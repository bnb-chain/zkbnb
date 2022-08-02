/*
 * Copyright © 2021 Zkbas Protocol
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
	"context"
	"sort"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/repo/accountoperator"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/repo/l2eventoperator"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/repo/liquidityoperator"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/repo/mempooloperator"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/repo/nftoperator"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/svc"
)

type l2BlockEventsMonitor struct {
	logx.Logger
	ctx               context.Context
	svcCtx            *svc.ServiceContext
	accountOperator   accountoperator.Model
	mempoolOperator   mempooloperator.Model
	liquidityOperator liquidityoperator.Model
	nftOperator       nftoperator.Model
	l2eventOperator   l2eventoperator.Model
	commglobalmap     commglobalmap.Model
}

func Newl2BlockEventsMonitor(ctx context.Context, svcCtx *svc.ServiceContext) *l2BlockEventsMonitor {
	return &l2BlockEventsMonitor{
		Logger:            logx.WithContext(ctx),
		ctx:               ctx,
		svcCtx:            svcCtx,
		accountOperator:   accountoperator.New(svcCtx),
		mempoolOperator:   mempooloperator.New(svcCtx),
		liquidityOperator: liquidityoperator.New(svcCtx),
		nftOperator:       nftoperator.New(svcCtx),
		l2eventOperator:   l2eventoperator.New(svcCtx),
		commglobalmap:     commglobalmap.New(svcCtx),
	}
}

func MonitorL2BlockEvents(ctx context.Context, svcCtx *svc.ServiceContext,
	bscCli *_rpc.ProviderClient, bscPendingBlocksCount uint64, mempoolModel MempoolModel, blockModel BlockModel, l1TxSenderModel L1TxSenderModel) (err error) {
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
			logx.Errorf("[MonitorL2BlockEvents] GetTransactionByHash err: %s", err.Error())
			continue
		}
		if isPending {
			continue
		}
		receipt, err := bscCli.GetTransactionReceipt(txHash)
		if err != nil {
			logx.Errorf("[MonitorL2BlockEvents] GetTransactionReceipt err: %s", err.Error())
			continue
		}
		// get latest l1 block height(latest height - pendingBlocksCount)
		latestHeight, err := bscCli.GetHeight()
		if err != nil {
			logx.Errorf("[MonitorL2BlockEvents] GetHeight err: %s", err.Error())
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
					logx.Errorf("[MonitorL2BlockEvents] GetBlockHeaderByHash err: %s", err.Error())
					return err
				}
				isQueriedBlockHash[vlog.BlockHash.Hex()] = int64(onChainBlockInfo.Time)
			}
			timeAt := isQueriedBlockHash[vlog.BlockHash.Hex()]
			switch vlog.Topics[0].Hex() {
			case zkbasLogBlockCommitSigHash.Hex():
				var event ZkbasBlockCommit
				if err = ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, vlog.Data); err != nil {
					logx.Errorf("[MonitorL2BlockEvents] UnpackIntoInterface err: %s", err.Error())
					return err
				}
				blockHeight := int64(event.BlockNumber)
				if relatedBlocks[blockHeight] == nil {
					relatedBlocks[blockHeight], err = blockModel.GetBlockByBlockHeightWithoutTx(blockHeight)
					if err != nil {
						logx.Errorf("[MonitorL2BlockEvents] GetBlockByBlockHeightWithoutTx err: %s", err.Error())
						return err
					}
				}
				if blockHeight == pendingSender.L2BlockHeight {
					isValidSender = true
				}
				relatedBlocks[blockHeight].CommittedTxHash = receipt.TxHash.Hex()
				relatedBlocks[blockHeight].CommittedAt = timeAt
				relatedBlocks[blockHeight].BlockStatus = block.StatusCommitted
			case zkbasLogBlockVerificationSigHash.Hex():
				var event ZkbasBlockVerification
				if err = ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data); err != nil {
					logx.Errorf("[MonitorL2BlockEvents] UnpackIntoInterface err: %s", err.Error())
					return err
				}
				blockHeight := int64(event.BlockNumber)
				if relatedBlocks[blockHeight] == nil {
					relatedBlocks[blockHeight], err = blockModel.GetBlockByBlockHeightWithoutTx(blockHeight)
					if err != nil {
						logx.Errorf("[MonitorL2BlockEvents] GetBlockByBlockHeightWithoutTx err: %s", err.Error())
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
			case zkbasLogBlocksRevertSigHash.Hex():
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
				logx.Errorf("[MonitorL2BlockEvents] GetMempoolTxsByBlockHeight err: %s", err.Error())
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
	if err = l1TxSenderModel.UpdateRelatedEventsAndResetRelatedAssetsAndTxs(pendingUpdateBlocks,
		pendingUpdateSenders, pendingUpdateMempoolTxs, pendingUpdateProofSenderStatus); err != nil {
		logx.Errorf("[MonitorL2BlockEvents] UpdateRelatedEventsAndResetRelatedAssetsAndTxs err: %s", err.Error())
		return err
	}
	// update account cache for globalrpc sendtx interface
	m := Newl2BlockEventsMonitor(ctx, svcCtx)
	for _, mempooltx := range pendingUpdateMempoolTxs {
		if err := m.commglobalmap.SetLatestAccountInfoInToCache(ctx, mempooltx.AccountIndex); err != nil {
			logx.Errorf("[CreateMempoolTxs] unable to CreateMempoolTxs, error: %s", err.Error())
		}
	}
	logx.Errorf("[MonitorL2BlockEvents] update blocks count: %d", len(pendingUpdateBlocks))
	logx.Errorf("[MonitorL2BlockEvents] update senders count: %d", len(pendingUpdateSenders))
	logx.Errorf("[MonitorL2BlockEvents] update mempool txs count: %d", len(pendingUpdateMempoolTxs))
	logx.Errorf("========== end MonitorL2BlockEvents ==========")
	return nil
}
