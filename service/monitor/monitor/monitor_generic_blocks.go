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
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb-eth-rpc/_rpc"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/zkbnb/core/legend"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
	"github.com/bnb-chain/zkbnb/dao/mempool"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/priorityrequest"
	"github.com/bnb-chain/zkbnb/tree"
	types2 "github.com/bnb-chain/zkbnb/types"
)

func (m *Monitor) MonitorGenericBlocks() (err error) {
	latestHandledBlock, err := m.L1SyncedBlockModel.GetLatestL1BlockByType(l1syncedblock.TypeGeneric)
	var handledHeight int64
	if err != nil {
		if err == types2.DbErrNotFound {
			handledHeight = m.Config.ChainConfig.StartL1BlockHeight
		} else {
			return fmt.Errorf("failed to get latest l1 monitor block, err: %v", err)
		}
	} else {
		handledHeight = latestHandledBlock.L1BlockHeight
	}

	// get latest l1 block height(latest height - pendingBlocksCount)
	latestHeight, err := m.cli.GetHeight()
	if err != nil {
		return fmt.Errorf("failed to get l1 height, err: %v", err)
	}

	safeHeight := latestHeight - m.Config.ChainConfig.ConfirmBlocksCount
	safeHeight = uint64(common2.MinInt64(int64(safeHeight), handledHeight+m.Config.ChainConfig.MaxHandledBlocksCount))
	if safeHeight <= uint64(handledHeight) {
		return nil
	}

	logx.Infof("syncing l1 blocks from %d to %d", big.NewInt(handledHeight+1), big.NewInt(int64(safeHeight)))

	priorityRequestCount, err := getPriorityRequestCount(m.cli, m.zkbnbContractAddress, uint64(handledHeight+1), safeHeight)
	if err != nil {
		return fmt.Errorf("failed to get priority request count, err: %v", err)
	}

	logs, err := getZkBNBContractLogs(m.cli, m.zkbnbContractAddress, uint64(handledHeight+1), safeHeight)
	if err != nil {
		return fmt.Errorf("failed to get contract logs, err: %v", err)
	}
	var (
		l1EventInfos     []*L1EventInfo
		priorityRequests []*priorityrequest.PriorityRequest

		priorityRequestCountCheck = 0

		relatedBlocks = make(map[int64]*block.Block)

		// revert
		revertTo                 int64
		revertBlocks             []*block.Block
		revertCompressedBlocks   []*compressedblock.CompressedBlock
		revertAccountHistories   []*account.AccountHistory
		revertLiquidityHistories []*liquidity.LiquidityHistory
		revertNftHistories       []*nft.L2NftHistory
		revertMempoolTxs         []*mempool.MempoolTx

		pendingResetAccountRegisterIndex []int64
		pendingResetAccounts             []*account.AccountHistory
		pendingResetLiquidities          []*liquidity.LiquidityHistory
		pendingResetNfts                 []*nft.L2NftHistory
	)
	for _, vlog := range logs {
		l1EventInfo := &L1EventInfo{
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
		case zkbnbLogBlocksRevertSigHash.Hex():
			var event zkbnb.ZkBNBBlocksRevert
			l1EventInfo.EventType = EventTypeRevertedBlock
			if err = ZkBNBContractAbi.UnpackIntoInterface(&event, EventNameBlocksRevert, vlog.Data); err != nil {
				logx.Errorf("[MonitorL2BlockEvents] UnpackIntoInterface err:%v", err)
				return err
			}

			// get revert to height
			revertTo = int64(event.TotalBlocksCommitted)
			// get pending delete block & tx info
			revertBlocks, err = m.BlockModel.GetBlocksForRevertWithTx(revertTo)
			if err != nil {
				logx.Errorf("[MonitorL2BlockEvents] unable to get blocks for revert: %s", err.Error())
				return err
			}
			// get pending delete block for commit
			revertCompressedBlocks, err = m.CompressedBlockModel.GetCompressedBlockForRevert(revertTo)
			if err != nil {
				logx.Errorf("[MonitorL2BlockEvents] unable to get blocks for commit for revert: %s", err.Error())
				return err
			}
			// get pending delete account history
			revertAccountHistories, err = m.AccountHistoryModel.GetAccountsForRevert(revertTo)
			if err != nil && err != types2.DbErrNotFound {
				logx.Errorf("[MonitorL2BlockEvents] unable to get accounts for revert: %s", err.Error())
				return err
			}
			// get pending delete liquidity history
			revertLiquidityHistories, err = m.LiquidityHistoryModel.GetLiquidityForRevert(revertTo)
			if err != nil && err != types2.DbErrNotFound {
				logx.Errorf("[MonitorL2BlockEvents] unable to get liquidity for revert: %s", err.Error())
				return err
			}
			// get pending delete nft history
			revertNftHistories, err = m.L2NftHistoryModel.GetNftsForRevert(revertTo)
			if err != nil && err != types2.DbErrNotFound {
				logx.Errorf("[MonitorL2BlockEvents] unable to get nfts for revert: %s", err.Error())
				return err
			}
			// get pending reset mempool tx and tx detail
			revertMempoolTxs, err = m.MempoolModel.GetMempoolTxsForRevert(revertTo)
			if err != nil && err != types2.DbErrNotFound {
				logx.Errorf("[MonitorL2BlockEvents] unable to get mempool txs for revert: %s", err.Error())
				return err
			}
			// reset account / liquidity / nft info
			var (
				pendingResetAccountIndexMap   = make(map[int64]bool)
				pendingResetLiquidityIndexMap = make(map[int64]bool)
				pendingResetNftIndexMap       = make(map[int64]bool)
				pendingResetAccountIndex      []int64
				pendingResetLiquidityIndex    []int64
				pendingResetNftIndex          []int64
			)
			for _, accountHistory := range revertAccountHistories {
				pendingResetAccountIndexMap[accountHistory.AccountIndex] = true
			}
			for _, liquidityHistory := range revertLiquidityHistories {
				pendingResetLiquidityIndexMap[liquidityHistory.PairIndex] = true
			}
			for _, nftHistory := range revertNftHistories {
				pendingResetNftIndexMap[nftHistory.NftIndex] = true
			}
			for index := range pendingResetAccountIndexMap {
				pendingResetAccountIndex = append(pendingResetAccountIndex, index)
			}
			for index := range pendingResetLiquidityIndexMap {
				pendingResetLiquidityIndex = append(pendingResetLiquidityIndex, index)
			}
			for index := range pendingResetNftIndexMap {
				pendingResetNftIndex = append(pendingResetNftIndex, index)
			}
			// get last account history info
			for _, index := range pendingResetAccountIndex {
				accountInfo, err := m.AccountHistoryModel.GetLatestAccountInfoByAccountIndexAndHeight(index, revertTo)
				if err != nil {
					if !errors.Is(err, types2.DbErrNotFound) {
						logx.Errorf("unable to get account info : %s", err.Error())
						return err
					}

					accountInfo = &account.AccountHistory{
						AccountIndex:    index,
						Nonce:           0,
						CollectionNonce: 0,
						AssetInfo:       types2.NilAssetInfo,
						AssetRoot:       common.Bytes2Hex(tree.NilAccountAssetRoot),
					}
				}
				pendingResetAccounts = append(pendingResetAccounts, accountInfo)
			}
			// get last liquidity history info
			for _, index := range pendingResetLiquidityIndex {
				liquidityInfo, err := m.LiquidityHistoryModel.GetLatestLiquidityInfoByPairIndexAndHeight(index, revertTo)
				if err != nil {
					if !errors.Is(err, types2.DbErrNotFound) {
						logx.Errorf("unable to get liquidity info : %s", err.Error())
						return err
					}
					continue
				}
				pendingResetLiquidities = append(pendingResetLiquidities, liquidityInfo)
			}
			// get last nft history info
			for _, index := range pendingResetNftIndex {
				nftInfo, err := m.L2NftHistoryModel.GetLatestNftAssetByIndexAndHeight(index, revertTo)
				if err != nil {
					if !errors.Is(err, types2.DbErrNotFound) {
						logx.Errorf("unable to get nft info : %s", err.Error())
						return err
					}
					continue
				}
				pendingResetNfts = append(pendingResetNfts, nftInfo)
			}
			// reset info
			for _, mempoolTx := range revertMempoolTxs {
				mempoolTx.L2BlockHeight = types2.NilBlockHeight
				mempoolTx.Status = mempool.PendingTxStatus
				// reset account status
				if mempoolTx.TxType == types2.TxTypeRegisterZns {
					pendingResetAccountRegisterIndex = append(pendingResetAccountRegisterIndex, mempoolTx.AccountIndex)
				}
			}
		default:
		}

		l1EventInfos = append(l1EventInfos, l1EventInfo)
	}
	if priorityRequestCount != priorityRequestCountCheck {
		return fmt.Errorf("new priority requests events not match, try it again")
	}

	eventInfosBytes, err := json.Marshal(l1EventInfos)
	if err != nil {
		return err
	}
	l1BlockMonitorInfo := &l1syncedblock.L1SyncedBlock{
		L1BlockHeight: int64(safeHeight),
		BlockInfo:     string(eventInfosBytes),
		Type:          l1syncedblock.TypeGeneric,
	}

	// get pending update blocks
	pendingUpdateBlocks := make([]*block.Block, 0, len(relatedBlocks))
	for _, pendingUpdateBlock := range relatedBlocks {
		pendingUpdateBlocks = append(pendingUpdateBlocks, pendingUpdateBlock)
	}

	// get mempool txs to delete
	pendingDeleteMempoolTxs, err := getMempoolTxsToDelete(pendingUpdateBlocks, m.MempoolModel)
	if err != nil {
		return fmt.Errorf("failed to get mempool txs to delete, err: %v", err)
	}

	// update db
	err = m.db.Transaction(func(tx *gorm.DB) error {
		// create l1 synced block
		err := m.L1SyncedBlockModel.CreateL1SyncedBlockInTransact(tx, l1BlockMonitorInfo)
		if err != nil {
			return err
		}
		// create priority requests
		err = m.PriorityRequestModel.CreatePriorityRequestsInTransact(tx, priorityRequests)
		if err != nil {
			return err
		}
		// update blocks
		err = m.BlockModel.UpdateBlocksWithoutTxsInTransact(tx, pendingUpdateBlocks)
		if err != nil {
			return err
		}
		// delete mempool txs
		err = m.MempoolModel.DeleteMempoolTxsInTransact(tx, pendingDeleteMempoolTxs)
		if err != nil {
			return err
		}

		if len(revertBlocks) == 0 {
			// skip revert
			return nil
		}

		// delete blocks
		err = m.BlockModel.DeleteBlocksInTransact(tx, revertBlocks)
		if err != nil {
			return err
		}
		for _, revertBlock := range revertBlocks {
			for _, revertTx := range revertBlock.Txs {
				// delete txDetails in tx
				err = m.TxDetailModel.DeleteTxsInTransact(tx, revertTx.TxDetails)
				if err != nil {
					return err
				}
			}
			// delete txs in block
			err = m.TxModel.DeleteTxsInTransact(tx, revertBlock.Txs)
			if err != nil {
				return err
			}
		}

		// delete compressedBlocks
		err = m.CompressedBlockModel.DeleteCompressedBlockInTransact(tx, revertCompressedBlocks)
		if err != nil {
			return err
		}
		// delete proof
		err = m.ProofModel.DeleteProofOverBlockHeightInTransact(tx, revertTo)
		if err != nil {
			return err
		}
		// delete witness
		err = m.BlockWitnessModel.DeleteBlockWitnessOverHeightInTransact(tx, revertTo)
		if err != nil {
			return err
		}
		// delete account history
		err = m.AccountHistoryModel.DeleteAccountHistoryInTransact(tx, revertAccountHistories)
		if err != nil {
			return err
		}
		// delete liquidity history
		err = m.LiquidityHistoryModel.DeleteLiquidityHistoriesInTransact(tx, revertLiquidityHistories)
		if err != nil {
			return err
		}
		// delete nft history
		err = m.L2NftHistoryModel.DeleteNftHistoriesInTransact(tx, revertNftHistories)
		if err != nil {
			return err
		}
		// reset mempool tx
		err = m.MempoolModel.UpdateMempoolTxsInTransact(tx, revertMempoolTxs)
		if err != nil {
			return err
		}
		// reset account info
		accountUpdates := make(map[int64]map[string]interface{}, len(pendingResetAccounts))
		for _, accountInfo := range pendingResetAccounts {
			accountUpdates[accountInfo.AccountIndex] = map[string]interface{}{
				"asset_info":       accountInfo.AssetInfo,
				"asset_root":       accountInfo.AssetRoot,
				"nonce":            accountInfo.Nonce,
				"collection_nonce": accountInfo.CollectionNonce,
			}
		}
		// reset account to pending status
		accountUpdates = make(map[int64]map[string]interface{}, len(pendingResetAccountRegisterIndex))
		for _, accountIndex := range pendingResetAccountRegisterIndex {
			accountUpdates[accountIndex] = map[string]interface{}{
				"status": account.AccountStatusPending,
			}
		}
		err = m.AccountModel.ResetAccountInTransact(tx, accountUpdates)
		if err != nil {
			return err
		}
		// reset liquidity info
		liquidityUpdates := make(map[int64]map[string]interface{}, len(pendingResetLiquidities))
		for _, liquidityInfo := range pendingResetLiquidities {
			liquidityUpdates[liquidityInfo.PairIndex] = map[string]interface{}{
				"asset_a":                liquidityInfo.AssetA,
				"asset_b":                liquidityInfo.AssetB,
				"lp_amount":              liquidityInfo.LpAmount,
				"k_last":                 liquidityInfo.KLast,
				"treasury_account_index": liquidityInfo.TreasuryAccountIndex,
				"treasury_rate":          liquidityInfo.TreasuryRate,
			}
		}
		err = m.LiquidityModel.ResetLiquidityInTransact(tx, liquidityUpdates)
		if err != nil {
			return err
		}
		// reset nft info
		nftUpdates := make(map[int64]map[string]interface{}, len(pendingResetNfts))
		for _, nftInfo := range pendingResetNfts {
			nftUpdates[nftInfo.NftIndex] = map[string]interface{}{
				"creator_account_index": nftInfo.CreatorAccountIndex,
				"owner_account_index":   nftInfo.OwnerAccountIndex,
				"nft_content_hash":      nftInfo.NftContentHash,
				"nft_l1_address":        nftInfo.NftL1Address,
				"nft_l1_token_id":       nftInfo.NftL1TokenId,
				"creator_treasury_rate": nftInfo.CreatorTreasuryRate,
				"collection_id":         nftInfo.CollectionId,
			}
		}
		err = m.L2NftModel.ResetNftsInTransact(tx, nftUpdates)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to store monitor info, err: %v", err)
	}
	logx.Info("create txs count:", len(priorityRequests))
	return nil
}

func getMempoolTxsToDelete(blocks []*block.Block, mempoolModel mempool.MempoolModel) ([]*mempool.MempoolTx, error) {
	var toDeleteMempoolTxs []*mempool.MempoolTx
	for _, pendingUpdateBlock := range blocks {
		if pendingUpdateBlock.BlockStatus == BlockVerifiedStatus {
			_, blockToDleteMempoolTxs, err := mempoolModel.GetMempoolTxsByBlockHeight(pendingUpdateBlock.BlockHeight)
			if err != nil {
				logx.Errorf("GetMempoolTxsByBlockHeight err: %s", err.Error())
				return nil, err
			}
			if len(blockToDleteMempoolTxs) == 0 {
				continue
			}
			toDeleteMempoolTxs = append(toDeleteMempoolTxs, blockToDleteMempoolTxs...)
		}
	}
	return toDeleteMempoolTxs, nil
}

func getZkBNBContractLogs(cli *_rpc.ProviderClient, zkbnbContract string, startHeight, endHeight uint64) ([]types.Log, error) {
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

func getPriorityRequestCount(cli *_rpc.ProviderClient, zkbnbContract string, startHeight, endHeight uint64) (int, error) {
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
