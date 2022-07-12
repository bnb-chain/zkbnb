/*
 * Copyright © 2021 Zecrey Protocol
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

package logic

import (
	"sort"

	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/model/proofSender"
)

func MonitorL2BlockEvents(bscCli *_rpc.ProviderClient, bscPendingBlocksCount uint64, mempoolModel MempoolModel, blockModel BlockModel, l1TxSenderModel L1TxSenderModel) (err error) {
	logx.Errorf("========== start MonitorL2BlockEvents ==========")
	pendingSenders, err := l1TxSenderModel.GetL1TxSendersByTxStatus(L1TxSenderPendingStatus)
	if err != nil {
		logx.Errorf("[MonitorL2BlockEvents] unable to get l1 tx senders by tx status: %s", err.Error())
		return err
	}
	var (
		relatedBlocks                  = make(map[int64]*Block)
		pendingUpdateSenders           []*L1TxSender
		pendingUpdateProofSenderStatus = make(map[int64]int)
	)
	for _, pendingSender := range pendingSenders {
		txHash := pendingSender.L1TxHash
		// check if the status of tx is success
		_, isPending, err := bscCli.GetTransactionByHash(txHash)
		if err != nil {
			logx.Errorf("[MonitorL2BlockEvents] GetTransactionByHash err: %s", err)
			continue
		}
		if isPending {
			continue
		}
		receipt, err := bscCli.GetTransactionReceipt(txHash)
		if err != nil {
			logx.Errorf("[MonitorL2BlockEvents] GetTransactionReceipt err: %s", err)
			continue
		}
		// get latest l1 block height(latest height - pendingBlocksCount)
		latestHeight, err := bscCli.GetHeight()
		if err != nil {
			logx.Errorf("[MonitorL2BlockEvents] GetHeight err: %v", err)
			return err
		}
		if latestHeight < receipt.BlockNumber.Uint64()+bscPendingBlocksCount {
			continue
		}
		var (
			isValidSender      bool
			isQueriedBlockHash = make(map[string]int64)
		)
		for _, vlog := range receipt.Logs {
			if isQueriedBlockHash[vlog.BlockHash.Hex()] == 0 {
				onChainBlockInfo, err := bscCli.GetBlockHeaderByHash(vlog.BlockHash.Hex())
				if err != nil {
					logx.Errorf("[MonitorL2BlockEvents] GetBlockHeaderByHash err: %v", err)
					return err
				}
				isQueriedBlockHash[vlog.BlockHash.Hex()] = int64(onChainBlockInfo.Time)
			}
			timeAt := isQueriedBlockHash[vlog.BlockHash.Hex()]
			switch vlog.Topics[0].Hex() {
			case zecreyLogBlockCommitSigHash.Hex():
				var event ZecreyLegendBlockCommit
				if err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, vlog.Data); err != nil {
					logx.Errorf("[MonitorL2BlockEvents] UnpackIntoInterface err:%v", err)
					return err
				}
				blockHeight := int64(event.BlockNumber)
				if relatedBlocks[blockHeight] == nil {
					relatedBlocks[blockHeight], err = blockModel.GetBlockByBlockHeightWithoutTx(blockHeight)
					if err != nil {
						logx.Errorf("[MonitorL2BlockEvents] GetBlockByBlockHeightWithoutTx err:%v", err)
						return err
					}
				}
				if blockHeight == pendingSender.L2BlockHeight {
					isValidSender = true
				}
				relatedBlocks[blockHeight].CommittedTxHash = receipt.TxHash.Hex()
				relatedBlocks[blockHeight].CommittedAt = timeAt
				relatedBlocks[blockHeight].BlockStatus = block.StatusCommitted
			case zecreyLogBlockVerificationSigHash.Hex():
				var event ZecreyLegendBlockVerification
				if err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data); err != nil {
					logx.Errorf("[MonitorL2BlockEvents] UnpackIntoInterface err:%v", err)
					return err
				}
				blockHeight := int64(event.BlockNumber)
				if relatedBlocks[blockHeight] == nil {
					relatedBlocks[blockHeight], err = blockModel.GetBlockByBlockHeightWithoutTx(blockHeight)
					if err != nil {
						logx.Errorf("[MonitorL2BlockEvents] GetBlockByBlockHeightWithoutTx err:%v", err)
						return err
					}
				}
				if blockHeight == pendingSender.L2BlockHeight {
					isValidSender = true
				}
				relatedBlocks[blockHeight].VerifiedTxHash = receipt.TxHash.Hex()
				relatedBlocks[blockHeight].VerifiedAt = timeAt
				relatedBlocks[blockHeight].BlockStatus = block.StatusVerifiedAndExecuted
				pendingUpdateProofSenderStatus[blockHeight] = proofSender.Confirmed
			case zecreyLogBlocksRevertSigHash.Hex():
				// TODO revert
			default:
			}
		}
		if isValidSender {
			pendingSender.TxStatus = L1TxSenderHandledStatus
			pendingUpdateSenders = append(pendingUpdateSenders, pendingSender)
		}
	}
	var pendingUpdateBlocks []*Block
	for _, pendingUpdateBlock := range relatedBlocks {
		pendingUpdateBlocks = append(pendingUpdateBlocks, pendingUpdateBlock)
	}
	if len(pendingUpdateBlocks) != 0 {
		sort.Slice(pendingUpdateBlocks, func(i, j int) bool {
			return pendingUpdateBlocks[i].BlockHeight < pendingUpdateBlocks[j].BlockHeight
		})
	}
	// handle executed blocks
	var pendingUpdateMempoolTxs []*MempoolTx
	for _, pendingUpdateBlock := range pendingUpdateBlocks {
		if pendingUpdateBlock.BlockStatus == BlockVerifiedStatus {
			rowsAffected, pendingDeleteMempoolTxs, err := mempoolModel.GetMempoolTxsByBlockHeight(pendingUpdateBlock.BlockHeight)
			if err != nil {
				logx.Errorf("[MonitorL2BlockEvents] GetMempoolTxsByBlockHeight err:%v", err)
				return err
			}
			if rowsAffected == 0 {
				continue
			}
			pendingUpdateMempoolTxs = append(pendingUpdateMempoolTxs, pendingDeleteMempoolTxs...)
		}
	}
	// update blocks, blockDetails, updateEvents, sender
	// update assets, locked assets, liquidity
	// delete mempool txs
	if err = l1TxSenderModel.UpdateRelatedEventsAndResetRelatedAssetsAndTxs(pendingUpdateBlocks, pendingUpdateSenders, pendingUpdateMempoolTxs, pendingUpdateProofSenderStatus); err != nil {
		logx.Errorf("[MonitorL2BlockEvents] UpdateRelatedEventsAndResetRelatedAssetsAndTxs err:%v", err)
		return err
	}
	logx.Errorf("[MonitorL2BlockEvents] update blocks count: %v", len(pendingUpdateBlocks))
	logx.Errorf("[MonitorL2BlockEvents] update senders count: %v", len(pendingUpdateSenders))
	logx.Errorf("[MonitorL2BlockEvents] update mempool txs count: %v", len(pendingUpdateMempoolTxs))
	logx.Errorf("========== end MonitorL2BlockEvents ==========")
	return nil
}
