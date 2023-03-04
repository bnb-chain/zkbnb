/*
 * Copyright © 2021 ZkBNB Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package generateproof

import (
	"context"
	"encoding/json"
	"fmt"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb/common/abicoder"
	monitor2 "github.com/bnb-chain/zkbnb/common/monitor"
	"github.com/bnb-chain/zkbnb/dao/exodusexit"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/service/monitor/monitor"
	types2 "github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"math/big"

	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
)

func (m *Monitor) MonitorGenericBlocks() (err error) {
	exodusExitBlock, err := m.ExodusExitBlockModel.GetLatestBlock()
	if err != nil && err != types2.DbErrNotFound {
		logx.Errorf("get latest verify block failed:%s", err.Error())
		return err
	}
	if exodusExitBlock != nil {
		if exodusExitBlock.BlockHeight == m.Config.ChainConfig.EndL2BlockHeight &&
			(exodusExitBlock.BlockStatus == exodusexit.StatusVerified || exodusExitBlock.BlockStatus == exodusexit.StatusExecuted) {
			logx.Info("get all the l2 blocks from l1 successfully")
			panic("success")
			return nil
		}
	}
	for {
		startHeight, endHeight, err := m.getBlockRangeToSync(l1syncedblock.TypeGeneric)
		if err != nil {
			logx.Errorf("get block range to sync error, err: %s", err.Error())
			return err
		}
		if endHeight < startHeight {
			logx.Infof("no blocks to sync, startHeight: %d, endHeight: %d", startHeight, endHeight)
			return nil
		}

		logx.Infof("syncing generic l1 blocks from %d to %d", big.NewInt(startHeight), big.NewInt(endHeight))

		logs, err := monitor.GetZkBNBContractLogs(m.cli, m.ZkBnbContractAddress, uint64(startHeight), uint64(endHeight))
		if err != nil {
			return fmt.Errorf("failed to get contract logs, err: %v", err)
		}

		logx.Infof("type is typeGeneric blocks from %d to %d and vlog len: %v", startHeight, endHeight, len(logs))
		for _, vlog := range logs {
			logx.Infof("type is typeGeneric blocks from %d to %d and vlog: %v", startHeight, endHeight, vlog)
		}
		var (
			l1Events      []*monitor2.L1Event
			relatedBlocks = make(map[int64]*exodusexit.ExodusExitBlock)
		)
		exit := false
		for _, vlog := range logs {
			l1EventInfo := &monitor2.L1Event{
				TxHash: vlog.TxHash.Hex(),
				Index:  vlog.Index,
			}
			if vlog.Removed {
				logx.Errorf("Removed to get vlog,TxHash:%v,Index:%v", l1EventInfo.TxHash, l1EventInfo.Index)
				continue
			}
			logBlock, err := m.cli.GetBlockHeaderByNumber(big.NewInt(int64(vlog.BlockNumber)))
			if err != nil {
				return fmt.Errorf("failed to get block header, err: %v", err)
			}

			switch vlog.Topics[0].Hex() {
			case monitor2.ZkbnbLogWithdrawalSigHash.Hex():
			case monitor2.ZkbnbLogWithdrawalPendingSigHash.Hex():
			case monitor2.ZkbnbLogBlockCommitSigHash.Hex():
				l1EventInfo.EventType = monitor2.EventTypeCommittedBlock

				var event zkbnb.ZkBNBBlockCommit
				if err := monitor2.ZkBNBContractAbi.UnpackIntoInterface(&event, monitor2.EventNameBlockCommit, vlog.Data); err != nil {
					return fmt.Errorf("failed to unpack ZkBNBBlockCommit event, err: %v", err)
				}

				// update block status
				blockHeight := int64(event.BlockNumber)
				if relatedBlocks[blockHeight] == nil {
					relatedBlocks[blockHeight] = &exodusexit.ExodusExitBlock{}
				}
				relatedBlocks[blockHeight].CommittedTxHash = vlog.TxHash.Hex()
				relatedBlocks[blockHeight].CommittedAt = int64(logBlock.Time)
				relatedBlocks[blockHeight].BlockStatus = exodusexit.StatusCommitted
				relatedBlocks[blockHeight].BlockHeight = blockHeight
			case monitor2.ZkbnbLogBlockVerificationSigHash.Hex():
				l1EventInfo.EventType = monitor2.EventTypeVerifiedBlock

				var event zkbnb.ZkBNBBlockVerification
				if err := monitor2.ZkBNBContractAbi.UnpackIntoInterface(&event, monitor2.EventNameBlockVerification, vlog.Data); err != nil {
					return fmt.Errorf("failed to unpack ZkBNBBlockVerification err: %v", err)
				}

				// update block status
				blockHeight := int64(event.BlockNumber)
				if relatedBlocks[blockHeight] == nil {
					relatedBlocks[blockHeight] = &exodusexit.ExodusExitBlock{}
				}
				relatedBlocks[blockHeight].VerifiedTxHash = vlog.TxHash.Hex()
				relatedBlocks[blockHeight].VerifiedAt = int64(logBlock.Time)
				relatedBlocks[blockHeight].BlockStatus = exodusexit.StatusVerified
				relatedBlocks[blockHeight].BlockHeight = blockHeight
				if blockHeight > m.Config.ChainConfig.EndL2BlockHeight {
					logx.Info("get all the l2 blocks from l1 successfully")
					return nil
				}
				if m.Config.ChainConfig.EndL2BlockHeight == blockHeight {
					exit = true
				}
			case monitor2.ZkbnbLogBlocksRevertSigHash.Hex():
				l1EventInfo.EventType = monitor2.EventTypeRevertedBlock
			default:
			}

			l1Events = append(l1Events, l1EventInfo)
		}
		heights := make([]int64, 0, len(relatedBlocks))

		for height, _ := range relatedBlocks {
			heights = append(heights, height)
		}

		blocks, err := m.ExodusExitBlockModel.GetBlocksByHeights(heights)
		if err != nil && err != types2.DbErrNotFound {
			return fmt.Errorf("failed to get blocks by heights, err: %v", err)
		}
		committedTxHashMap := make(map[string]bool, 0)

		for height, pendingUpdateBlock := range relatedBlocks {
			for _, block := range blocks {
				if block.BlockHeight == height {
					pendingUpdateBlock.ID = block.ID
					if pendingUpdateBlock.CommittedTxHash == "" {
						pendingUpdateBlock.CommittedTxHash = block.CommittedTxHash
					}
					break
				}
			}
			if exodusexit.StatusVerified == pendingUpdateBlock.BlockStatus {
				if pendingUpdateBlock.CommittedTxHash == "" {
					return fmt.Errorf("committed tx hash is blank, block height: %d", pendingUpdateBlock.BlockHeight)
				}
				committedTxHashMap[pendingUpdateBlock.CommittedTxHash] = true
			}
		}
		commitBlockInfoList := make([]OldZkBNBCommitBlockInfo, 0)
		for committedTx, _ := range committedTxHashMap {
			commitBlocksCallData, err := getCommitBlocksCallData(m.cli, committedTx)
			if err != nil {
				return err
			}
			commitBlockInfoList = append(commitBlockInfoList, commitBlocksCallData.NewBlocksData...)
		}

		updateBlocks := make([]*exodusexit.ExodusExitBlock, 0)

		for height, pendingUpdateBlock := range relatedBlocks {
			for _, commitBlockInfo := range commitBlockInfoList {
				if commitBlockInfo.BlockNumber == uint32(height) {
					pendingUpdateBlock.BlockSize = commitBlockInfo.BlockSize
					pendingUpdateBlock.PubData = common.Bytes2Hex(commitBlockInfo.PublicData)
					break
				}
			}
			updateBlocks = append(updateBlocks, pendingUpdateBlock)
		}
		eventInfosBytes, err := json.Marshal(l1Events)
		if err != nil {
			return err
		}
		l1BlockMonitorInfo := &l1syncedblock.L1SyncedBlock{
			L1BlockHeight: endHeight,
			BlockInfo:     string(eventInfosBytes),
			Type:          l1syncedblock.TypeGeneric,
		}

		//update db
		err = m.db.Transaction(func(tx *gorm.DB) error {
			//create l1 synced block
			err := m.L1SyncedBlockModel.CreateL1SyncedBlockInTransact(tx, l1BlockMonitorInfo)
			if err != nil {
				return err
			}
			//update blocks
			err = m.ExodusExitBlockModel.BatchInsertOrUpdateInTransact(tx, updateBlocks)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to store monitor info, err: %v", err)
		}
		if exit {
			logx.Info("get all the l2 blocks from l1 successfully")
			return nil
		}
	}
}

func getCommitBlocksCallData(cli *rpc.ProviderClient, hash string) (*CommitBlocksCallData, error) {
	newABIDecoder := abicoder.NewABIDecoder(monitor2.ZkBNBContractAbi)
	transaction, _, err := cli.Client.TransactionByHash(context.Background(), common.HexToHash(hash))
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	storageStoredBlockInfo := StorageStoredBlockInfo{}
	newBlocksData := make([]OldZkBNBCommitBlockInfo, 0)
	callData := CommitBlocksCallData{LastCommittedBlockData: &storageStoredBlockInfo, NewBlocksData: newBlocksData}
	if err := newABIDecoder.UnpackIntoInterface(&callData, "commitBlocks", transaction.Data()[4:]); err != nil {
		logx.Severe(err)
		return nil, err
	}
	return &callData, nil
}

func getVerifyAndExecuteBlocksCallData(cli *rpc.ProviderClient, hash string) (*VerifyAndExecuteBlocksCallData, error) {
	newABIDecoder := abicoder.NewABIDecoder(monitor2.ZkBNBContractAbi)
	transaction, _, err := cli.Client.TransactionByHash(context.Background(), common.HexToHash(hash))
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	newBlocksData := make([]ZkBNBVerifyAndExecuteBlockInfo, 0)
	proofs := make([]*big.Int, 0)
	callData := VerifyAndExecuteBlocksCallData{Proofs: proofs, VerifyAndExecuteBlocksInfo: newBlocksData}
	if err := newABIDecoder.UnpackIntoInterface(&callData, "verifyAndExecuteBlocks", transaction.Data()[4:]); err != nil {
		logx.Severe(err)
		return nil, err
	}
	return &callData, nil
}
