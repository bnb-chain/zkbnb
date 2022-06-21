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
	"fmt"
	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zeromicro/go-zero/core/logx"
	"sort"
	"time"
)

/*
	MonitorL2BlockEvents: monitor layer-2 block events
*/
func MonitorL2BlockEvents(
	bscCli *_rpc.ProviderClient,
	bscPendingBlocksCount uint64,
	mempoolModel MempoolModel,
	blockModel BlockModel,
	l1TxSenderModel L1TxSenderModel,
) (err error) {
	// get pending transactions from l1TxSender
	pendingSenders, err := l1TxSenderModel.GetL1TxSendersByTxStatus(L1TxSenderPendingStatus)
	if err != nil {
		logx.Errorf("[MonitorL2BlockEvents] unable to get l1 tx senders by tx status: %s", err.Error())
		return err
	}
	// scan each event
	var (
		// pending update blocks
		relatedBlocks                  = make(map[int64]*Block)
		pendingUpdateSenders           []*L1TxSender
		pendingUpdateProofSenderStatus = make(map[int64]int)
	)
	// handle each sender
	for _, pendingSender := range pendingSenders {
		txHash := pendingSender.L1TxHash
		var (
			l1BlockNumber uint64
			receipt       *types.Receipt
		)
		// check if the status of tx is success
		// get latest l1 block height(latest height - pendingBlocksCount)
		latestHeight, err := bscCli.GetHeight()
		if err != nil {
			errInfo := fmt.Sprintf("[MonitorL2BlockEvents]<=>[cli.GetHeight] %s", err.Error())
			logx.Error(errInfo)
			return err
		}
		_, isPending, err := bscCli.GetTransactionByHash(txHash)
		if err != nil {
			logx.Errorf("[MonitorL2BlockEvents] unable to get transaction by hash: %s", err.Error())
			continue
		}
		if isPending {
			logx.Errorf("[MonitorL2BlockEvents] the tx is still pending, just handle next sender: %s", txHash)
			continue
		}
		// get receipt
		receipt, err = bscCli.GetTransactionReceipt(txHash)
		if err != nil {
			logx.Errorf("[MonitorL2BlockEvents] unable to get tx receipt: %s", err.Error())
			continue
		}
		l1BlockNumber = receipt.BlockNumber.Uint64()
		// check if the height is over safe height
		if latestHeight < l1BlockNumber+bscPendingBlocksCount {
			logx.Infof("[MonitorL2BlockEvents] haven't reached to safe block height, should wait: %s", txHash)
			continue
		}
		// get events from the tx
		logs := receipt.Logs
		timeAt := time.Now().UnixMilli()
		var isValidSender bool
		for _, vlog := range logs {
			switch vlog.Topics[0].Hex() {
			case ZecreyLogBlockCommitSigHash.Hex():
				// parse event info
				var event ZecreyLegendBlockCommit
				err = ZecreyContractAbi.UnpackIntoInterface(&event, BlockCommitEventName, vlog.Data)
				if err != nil {
					errInfo := fmt.Sprintf("[MonitorL2BlockEvents]<=>[ZecreyContractAbi.UnpackIntoInterface] %s", err.Error())
					logx.Error(errInfo)
					return err
				}
				// get related blocks
				blockHeight := int64(event.BlockNumber)
				if relatedBlocks[blockHeight] == nil {
					relatedBlocks[blockHeight], err = blockModel.GetBlockByBlockHeightWithoutTx(blockHeight)
					if err != nil {
						logx.Errorf("[MonitorL2BlockEvents] unable to get block by block height: %s", err.Error())
						return err
					}
				}
				// check block height
				if blockHeight == pendingSender.L2BlockHeight {
					isValidSender = true
				}
				relatedBlocks[blockHeight].CommittedTxHash = receipt.TxHash.Hex()
				relatedBlocks[blockHeight].CommittedAt = timeAt
				relatedBlocks[blockHeight].BlockStatus = block.StatusCommitted
				break
			case ZecreyLogBlockVerificationSigHash.Hex():
				// parse event info
				var event ZecreyLegendBlockVerification
				err = ZecreyContractAbi.UnpackIntoInterface(&event, BlockVerificationEventName, vlog.Data)
				if err != nil {
					errInfo := fmt.Sprintf("[blockMoniter.MonitorL2BlockEvents]<=>[ZecreyContractAbi.UnpackIntoInterface] %s", err.Error())
					logx.Error(errInfo)
					return err
				}
				// get related blocks
				blockHeight := int64(event.BlockNumber)
				if relatedBlocks[blockHeight] == nil {
					relatedBlocks[blockHeight], err = blockModel.GetBlockByBlockHeightWithoutTx(blockHeight)
					if err != nil {
						logx.Errorf("[MonitorL2BlockEvents] unable to get block by block height: %s", err.Error())
						return err
					}
				}
				// check block height
				if blockHeight == pendingSender.L2BlockHeight {
					isValidSender = true
				}
				// update block status
				relatedBlocks[blockHeight].VerifiedTxHash = receipt.TxHash.Hex()
				relatedBlocks[blockHeight].VerifiedAt = timeAt
				relatedBlocks[blockHeight].BlockStatus = block.StatusVerifiedAndExecuted
				pendingUpdateProofSenderStatus[blockHeight] = proofSender.Confirmed
				break
			case ZecreyLogBlocksRevertSigHash.Hex():
				// TODO revert
				break
			default:
				break
			}
		}
		if isValidSender {
			// update sender status
			pendingSender.TxStatus = L1TxSenderHandledStatus
			pendingUpdateSenders = append(pendingUpdateSenders, pendingSender)
		}
	}
	// get pending update info
	var (
		pendingUpdateBlocks []*Block
	)
	for _, pendingUpdateBlock := range relatedBlocks {
		pendingUpdateBlocks = append(pendingUpdateBlocks, pendingUpdateBlock)
	}
	// sort for blocks
	if len(pendingUpdateBlocks) != 0 {
		sort.Sort(blockInfosByBlockHeight(pendingUpdateBlocks))
		logx.Infof("pending update blocks count: %v and height: %v", len(pendingUpdateBlocks), pendingUpdateBlocks[len(pendingUpdateBlocks)-1].BlockHeight)
	}

	// handle executed blocks
	var (
		pendingUpdateMempoolTxs []*MempoolTx
	)
	for _, pendingUpdateBlock := range pendingUpdateBlocks {
		if pendingUpdateBlock.BlockStatus == BlockVerifiedStatus {
			// delete related mempool txs
			rowsAffected, pendingDeleteMempoolTxs, err := mempoolModel.GetMempoolTxsByBlockHeight(pendingUpdateBlock.BlockHeight)
			if err != nil {
				errInfo := fmt.Sprintf("[MonitorL2BlockEvents] unable to get related mempool txs by height: %s", err.Error())
				logx.Error(errInfo)
				return err
			}
			if rowsAffected == 0 {
				logx.Error("[MonitorL2BlockEvents] invalid txs size or mempool has been deleted")
				continue
			}
			pendingUpdateMempoolTxs = append(pendingUpdateMempoolTxs, pendingDeleteMempoolTxs...)
		}
	}
	// update blocks, blockDetails, updateEvents, sender
	// update assets, locked assets, liquidity
	// delete mempool txs
	err = l1TxSenderModel.UpdateRelatedEventsAndResetRelatedAssetsAndTxs(
		pendingUpdateBlocks,
		pendingUpdateSenders,
		pendingUpdateMempoolTxs,
		pendingUpdateProofSenderStatus,
	)
	logx.Infof("[MonitorL2BlockEvents] update blocks count: %v", len(pendingUpdateBlocks))
	logx.Infof("[MonitorL2BlockEvents] update senders count: %v", len(pendingUpdateSenders))
	logx.Infof("[MonitorL2BlockEvents] update mempool txs count: %v", len(pendingUpdateMempoolTxs))
	if err != nil {
		logx.Errorf("[MonitorL2BlockEvents] unable to update everything: %s", err.Error())
		return err
	}
	return nil
}
