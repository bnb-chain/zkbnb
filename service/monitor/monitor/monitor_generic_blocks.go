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

package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/dao/priorityrequest"
	"github.com/bnb-chain/zkbnb/dao/proof"
	"github.com/bnb-chain/zkbnb/dao/tx"
	types2 "github.com/bnb-chain/zkbnb/types"
)

func (m *Monitor) MonitorGenericBlocks() (err error) {
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

	priorityRequestCount, err := getPriorityRequestCount(m.cli, m.zkbnbContractAddress, uint64(startHeight), uint64(endHeight))
	if err != nil {
		return fmt.Errorf("failed to get priority request count, err: %v", err)
	}

	logs, err := getZkBNBContractLogs(m.cli, m.zkbnbContractAddress, uint64(startHeight), uint64(endHeight))
	if err != nil {
		return fmt.Errorf("failed to get contract logs, err: %v", err)
	}
	l1GenericStartHeightMetric.Set(float64(startHeight))
	l1GenericEndHeightMetric.Set(float64(endHeight))
	l1GenericLenHeightMetric.Set(float64(len(logs)))

	logx.Infof("type is typeGeneric blocks from %d to %d and vlog len: %v", startHeight, endHeight, len(logs))
	for _, vlog := range logs {
		logx.Infof("type is typeGeneric blocks from %d to %d and vlog: %v", startHeight, endHeight, vlog)
	}
	var (
		l1Events         []*L1Event
		priorityRequests []*priorityrequest.PriorityRequest

		priorityRequestCountCheck = 0

		relatedBlocks        = make(map[int64]*block.Block)
		relatedBlockTxStatus = make(map[int64]int)
	)
	for _, vlog := range logs {
		l1EventInfo := &L1Event{
			TxHash: vlog.TxHash.Hex(),
		}

		logBlock, err := m.cli.GetBlockHeaderByNumber(big.NewInt(int64(vlog.BlockNumber)))
		if err != nil {
			return fmt.Errorf("failed to get block header, err: %v", err)
		}

		switch vlog.Topics[0].Hex() {
		case zkbnbLogNewPriorityRequestSigHash.Hex():
			priorityRequestCountCheck++
			l1EventInfo.EventType = EventTypeNewPriorityRequest

			l2TxEventMonitorInfo, err := convertLogToNewPriorityRequestEvent(vlog)
			if err != nil {
				return fmt.Errorf("failed to convert NewPriorityRequest log, err: %v", err)
			}
			priorityRequests = append(priorityRequests, l2TxEventMonitorInfo)
		case zkbnbLogWithdrawalSigHash.Hex():
		case zkbnbLogWithdrawalPendingSigHash.Hex():
		case zkbnbLogBlockCommitSigHash.Hex():
			l1EventInfo.EventType = EventTypeCommittedBlock

			var event zkbnb.ZkBNBBlockCommit
			if err := ZkBNBContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, vlog.Data); err != nil {
				return fmt.Errorf("failed to unpack ZkBNBBlockCommit event, err: %v", err)
			}

			// update block status
			blockHeight := int64(event.BlockNumber)
			if relatedBlocks[blockHeight] == nil {
				relatedBlocks[blockHeight], err = m.BlockModel.GetBlockByHeightWithoutTx(blockHeight)
				if err != nil {
					return fmt.Errorf("GetBlockByHeightWithoutTx err: %v", err)
				}
			}
			relatedBlocks[blockHeight].CommittedTxHash = vlog.TxHash.Hex()
			relatedBlocks[blockHeight].CommittedAt = int64(logBlock.Time)
			relatedBlocks[blockHeight].BlockStatus = block.StatusCommitted
			relatedBlockTxStatus[blockHeight] = tx.StatusCommitted
		case zkbnbLogBlockVerificationSigHash.Hex():
			l1EventInfo.EventType = EventTypeVerifiedBlock

			var event zkbnb.ZkBNBBlockVerification
			if err := ZkBNBContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data); err != nil {
				return fmt.Errorf("failed to unpack ZkBNBBlockVerification err: %v", err)
			}

			// update block status
			blockHeight := int64(event.BlockNumber)
			if relatedBlocks[blockHeight] == nil {
				relatedBlocks[blockHeight], err = m.BlockModel.GetBlockByHeightWithoutTx(blockHeight)
				if err != nil {
					return fmt.Errorf("failed to GetBlockByHeightWithoutTx: %v", err)
				}
			}
			relatedBlocks[blockHeight].VerifiedTxHash = vlog.TxHash.Hex()
			relatedBlocks[blockHeight].VerifiedAt = int64(logBlock.Time)
			relatedBlocks[blockHeight].BlockStatus = block.StatusVerifiedAndExecuted
			relatedBlockTxStatus[blockHeight] = tx.StatusVerified
		case zkbnbLogBlocksRevertSigHash.Hex():
			l1EventInfo.EventType = EventTypeRevertedBlock
		default:
		}

		l1Events = append(l1Events, l1EventInfo)
	}
	if priorityRequestCount != priorityRequestCountCheck {
		return fmt.Errorf("new priority requests events not match, try it again")
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

	// get pending update blocks
	pendingUpdateBlocks := make([]*block.Block, 0, len(relatedBlocks))
	pendingUpdateCommittedBlocks := make(map[string]*block.Block, 0)
	pendingUpdateVerifiedBlocks := make(map[string]*block.Block, 0)
	for _, pendingUpdateBlock := range relatedBlocks {
		pendingUpdateBlocks = append(pendingUpdateBlocks, pendingUpdateBlock)
		if pendingUpdateBlock.CommittedTxHash != "" {
			b, exist := pendingUpdateCommittedBlocks[pendingUpdateBlock.CommittedTxHash]
			if exist {
				if b.BlockHeight < pendingUpdateBlock.BlockHeight {
					pendingUpdateCommittedBlocks[pendingUpdateBlock.CommittedTxHash] = pendingUpdateBlock
				}
			} else {
				pendingUpdateCommittedBlocks[pendingUpdateBlock.CommittedTxHash] = pendingUpdateBlock
			}
		}
		if pendingUpdateBlock.VerifiedTxHash != "" {
			b, exist := pendingUpdateVerifiedBlocks[pendingUpdateBlock.VerifiedTxHash]
			if exist {
				if b.BlockHeight < pendingUpdateBlock.BlockHeight {
					pendingUpdateVerifiedBlocks[pendingUpdateBlock.VerifiedTxHash] = pendingUpdateBlock
				}
			} else {
				pendingUpdateVerifiedBlocks[pendingUpdateBlock.VerifiedTxHash] = pendingUpdateBlock
			}
		}
	}

	//update db
	err = m.db.Transaction(func(tx *gorm.DB) error {
		//create l1 synced block
		err := m.L1SyncedBlockModel.CreateL1SyncedBlockInTransact(tx, l1BlockMonitorInfo)
		if err != nil {
			return err
		}
		l1SyncedBlockHeightMetric.Set(float64(l1BlockMonitorInfo.L1BlockHeight))
		//create priority requests
		err = m.PriorityRequestModel.CreatePriorityRequestsInTransact(tx, priorityRequests)
		if err != nil {
			return err
		}
		for _, request := range priorityRequests {
			priorityOperationCreateMetric.Set(float64(request.RequestId))
			priorityOperationHeightCreateMetric.Set(float64(request.L1BlockHeight))
		}
		//update blocks
		err = m.BlockModel.UpdateBlocksWithoutTxsInTransact(tx, pendingUpdateBlocks)
		if err != nil {
			return err
		}
		// update l1 rollup tx status
		// maybe already updated by sender, or may be deleted by sender because of timeout
		for _, val := range pendingUpdateCommittedBlocks {
			_, err = m.L1RollupTxModel.GetL1RollupTxsByHash(val.CommittedTxHash)
			if err == types2.DbErrNotFound {
				logx.Info("monitor create commit rollup tx ", val.CommittedTxHash, val.BlockHeight)
				// the rollup tx is deleted by sender
				// so we insert it here
				err = m.L1RollupTxModel.CreateL1RollupTx(&l1rolluptx.L1RollupTx{
					L1TxHash:      val.CommittedTxHash,
					TxStatus:      l1rolluptx.StatusHandled,
					TxType:        l1rolluptx.TxTypeCommit,
					L2BlockHeight: val.BlockHeight,
				})
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
		pendingUpdateProofStatus := make(map[int64]int)
		for _, val := range pendingUpdateVerifiedBlocks {
			_, err = m.L1RollupTxModel.GetL1RollupTxsByHash(val.VerifiedTxHash)
			if err == types2.DbErrNotFound {
				logx.Info("monitor create verify rollup tx ", val.VerifiedTxHash, val.BlockHeight)
				// the rollup tx is deleted by sender
				// so we insert it here
				err = m.L1RollupTxModel.CreateL1RollupTx(&l1rolluptx.L1RollupTx{
					L1TxHash:      val.VerifiedTxHash,
					TxStatus:      l1rolluptx.StatusHandled,
					TxType:        l1rolluptx.TxTypeVerifyAndExecute,
					L2BlockHeight: val.BlockHeight,
				})
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			pendingUpdateProofStatus[val.BlockHeight] = proof.Confirmed
		}
		// update proof status
		if len(pendingUpdateProofStatus) != 0 {
			err = m.ProofModel.UpdateProofsInTransact(tx, pendingUpdateProofStatus)
			if err != nil {
				return err
			}
		}

		//update tx status
		err = m.TxModel.UpdateTxsStatusInTransact(tx, relatedBlockTxStatus)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to store monitor info, err: %v", err)
	}
	logx.Info("create txs count:", len(priorityRequests))
	return nil
}

func getZkBNBContractLogs(cli *rpc.ProviderClient, zkbnbContract string, startHeight, endHeight uint64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startHeight)),
		ToBlock:   big.NewInt(int64(endHeight)),
		Addresses: []common.Address{common.HexToAddress(zkbnbContract)},
	}
	logs, err := cli.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func getPriorityRequestCount(cli *rpc.ProviderClient, zkbnbContract string, startHeight, endHeight uint64) (int, error) {
	zkbnbInstance, err := zkbnb.LoadZkBNBInstance(cli, zkbnbContract)
	if err != nil {
		return 0, err
	}
	priorityRequests, err := zkbnbInstance.ZkBNBFilterer.
		FilterNewPriorityRequest(&bind.FilterOpts{Start: startHeight, End: &endHeight})
	if err != nil {
		return 0, err
	}
	priorityRequestCount := 0
	for priorityRequests.Next() {
		priorityRequestCount++
	}
	return priorityRequestCount, nil
}

func convertLogToNewPriorityRequestEvent(log types.Log) (*priorityrequest.PriorityRequest, error) {
	var event zkbnb.ZkBNBNewPriorityRequest
	if err := ZkBNBContractAbi.UnpackIntoInterface(&event, EventNameNewPriorityRequest, log.Data); err != nil {
		return nil, err
	}
	request := &priorityrequest.PriorityRequest{
		L1TxHash:        log.TxHash.Hex(),
		L1BlockHeight:   int64(log.BlockNumber),
		SenderAddress:   event.Sender.Hex(),
		RequestId:       int64(event.SerialId),
		TxType:          int64(event.TxType),
		Pubdata:         common.Bytes2Hex(event.PubData),
		ExpirationBlock: event.ExpirationBlock.Int64(),
		Status:          priorityrequest.PendingStatus,
	}
	return request, nil
}
