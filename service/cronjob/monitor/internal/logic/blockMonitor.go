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
	"encoding/json"
	"errors"
	"math/big"

	"github.com/bnb-chain/zkbas/common/model/l2BlockEventMonitor"

	"github.com/bnb-chain/zkbas/common/model/mempool"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/l1BlockMonitor"
	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/errorcode"
)

/*
	MonitorBlocks: monitor layer-1 block events
*/
func MonitorBlocks(
	cli *_rpc.ProviderClient,
	startHeight int64,
	pendingBlocksCount uint64,
	maxHandledBlocksCount int64,
	zkbasContract string,
	l1BlockMonitorModel l1BlockMonitor.L1BlockMonitorModel,
	blockModel block.BlockModel,
	mempoolModel mempool.MempoolModel,
) (err error) {
	latestHandledBlock, err := l1BlockMonitorModel.GetLatestL1BlockMonitorByBlock()
	logx.Info("========== start MonitorBlocks ==========")
	var handledHeight int64
	if err != nil {
		if err == errorcode.DbErrNotFound {
			handledHeight = startHeight
		} else {
			logx.Errorf("get latest l1 monitor block error, err: %s", err.Error())
			return err
		}
	} else {
		handledHeight = latestHandledBlock.L1BlockHeight
	}

	// get latest l1 block height(latest height - pendingBlocksCount)
	latestHeight, err := cli.GetHeight()
	if err != nil {
		logx.Errorf("get l1 height error, err: %s", err.Error())
		return err
	}

	safeHeight := latestHeight - pendingBlocksCount
	safeHeight = uint64(util.MinInt64(int64(safeHeight), handledHeight+maxHandledBlocksCount))
	if safeHeight <= uint64(handledHeight) {
		logx.Info("no new blocks need to be handled")
		return nil
	}
	logx.Infof("fromBlock: %d, toBlock: %d", big.NewInt(handledHeight+1), big.NewInt(int64(safeHeight)))

	priorityRequestCount, err := getPriorityRequestCount(cli, zkbasContract, uint64(handledHeight+1), safeHeight)
	if err != nil {
		logx.Errorf("get priority request count error, err: %s", err.Error())
		return err
	}

	logs, err := getZkbasContractLogs(cli, zkbasContract, uint64(handledHeight+1), safeHeight)
	if err != nil {
		logx.Error("get contract logs error, err: %s", err.Error())
		return err
	}
	var (
		l1EventInfos         []*L1EventInfo
		l2TxEventMonitors    []*l2TxEventMonitor.L2TxEventMonitor
		l2BlockEventMonitors []*l2BlockEventMonitor.L2BlockEventMonitor

		priorityRequestCountCheck = 0

		relatedBlocks = make(map[int64]*block.Block)
	)
	for _, vlog := range logs {
		l1EventInfo := &L1EventInfo{
			TxHash: vlog.TxHash.Hex(),
		}

		logBlock, err := cli.GetBlockHeaderByNumber(big.NewInt(int64(vlog.BlockNumber)))
		if err != nil {
			logx.Errorf("get block header error, err: %s", err.Error())
			return err
		}

		switch vlog.Topics[0].Hex() {
		case zkbasLogNewPriorityRequestSigHash.Hex():
			priorityRequestCountCheck++
			l1EventInfo.EventType = EventTypeNewPriorityRequest

			l2TxEventMonitorInfo, err := convertLogToNewPriorityRequestEvent(vlog)
			if err != nil {
				logx.Errorf("convert NewPriorityRequest log error, err: %s", err.Error())
				return err
			}
			l2TxEventMonitors = append(l2TxEventMonitors, l2TxEventMonitorInfo)
		case zkbasLogWithdrawalSigHash.Hex():
		case zkbasLogWithdrawalPendingSigHash.Hex():
		case zkbasLogBlockCommitSigHash.Hex():
			l1EventInfo.EventType = EventTypeCommittedBlock

			l2BlockEventMonitorInfo, err := convertLogToBlockCommitEvent(vlog)
			if err != nil {
				logx.Errorf("convert CommittedBlock log error, err: %s", err.Error())
				return err
			}
			l2BlockEventMonitors = append(l2BlockEventMonitors, l2BlockEventMonitorInfo)

			// update block status
			blockHeight := l2BlockEventMonitorInfo.L2BlockHeight
			if relatedBlocks[blockHeight] == nil {
				relatedBlocks[blockHeight], err = blockModel.GetBlockByBlockHeightWithoutTx(blockHeight)
				if err != nil {
					logx.Errorf("GetBlockByBlockHeightWithoutTx err: %s", err.Error())
					return err
				}
			}
			relatedBlocks[blockHeight].CommittedTxHash = vlog.TxHash.Hex()
			relatedBlocks[blockHeight].CommittedAt = int64(logBlock.Time)
			relatedBlocks[blockHeight].BlockStatus = block.StatusCommitted
		case zkbasLogBlockVerificationSigHash.Hex():
			l1EventInfo.EventType = EventTypeVerifiedBlock

			l2BlockEventMonitorInfo, err := convertLogToBlockVerificationEvent(vlog)
			if err != nil {
				logx.Errorf("convert TypeVerifiedBlock log error, err: %s", err.Error())
				return err
			}
			l2BlockEventMonitors = append(l2BlockEventMonitors, l2BlockEventMonitorInfo)

			// update block status
			blockHeight := l2BlockEventMonitorInfo.L2BlockHeight
			if relatedBlocks[blockHeight] == nil {
				relatedBlocks[blockHeight], err = blockModel.GetBlockByBlockHeightWithoutTx(blockHeight)
				if err != nil {
					logx.Errorf("GetBlockByBlockHeightWithoutTx err: %s", err.Error())
					return err
				}
			}
			relatedBlocks[blockHeight].VerifiedTxHash = vlog.TxHash.Hex()
			relatedBlocks[blockHeight].VerifiedAt = int64(logBlock.Time)
			relatedBlocks[blockHeight].BlockStatus = block.StatusVerifiedAndExecuted
		case zkbasLogBlocksRevertSigHash.Hex():
			l1EventInfo.EventType = EventTypeRevertedBlock
			l2BlockEventMonitorInfo, err := convertLogToBlocksRevertEvent(vlog)
			if err != nil {
				logx.Errorf("convert RevertedBlock log error, err: %s", err.Error())
				return err
			}
			l2BlockEventMonitors = append(l2BlockEventMonitors, l2BlockEventMonitorInfo)
		default:
		}

		l1EventInfos = append(l1EventInfos, l1EventInfo)
	}
	if priorityRequestCount != priorityRequestCountCheck {
		logx.Errorf("new priority requests events not match, try it again")
		return errors.New("new priority requests events not match, try it again")
	}

	eventInfosBytes, err := json.Marshal(l1EventInfos)
	if err != nil {
		logx.Errorf("marshal l1 events error, err: %s", err.Error())
		return err
	}
	l1BlockMonitorInfo := &l1BlockMonitor.L1BlockMonitor{
		L1BlockHeight: int64(safeHeight),
		BlockInfo:     string(eventInfosBytes),
		MonitorType:   l1BlockMonitor.MonitorTypeBlock,
	}

	// get pending update blocks
	pendingUpdateBlocks := make([]*block.Block, 0, len(relatedBlocks))
	for _, pendingUpdateBlock := range relatedBlocks {
		pendingUpdateBlocks = append(pendingUpdateBlocks, pendingUpdateBlock)
	}

	// get mempool txs to delete
	pendingDeleteMempoolTxs, err := getMempoolTxsToDelete(pendingUpdateBlocks, mempoolModel)
	if err != nil {
		logx.Errorf("get mempool txs to delete error, err: %s", err.Error())
		return err
	}

	if err = l1BlockMonitorModel.CreateMonitorsInfoAndUpdateBlocksAndTxs(l1BlockMonitorInfo, l2TxEventMonitors,
		l2BlockEventMonitors, pendingUpdateBlocks, pendingDeleteMempoolTxs); err != nil {
		logx.Error("store monitor info error, err: %s", err.Error())
		return err
	}
	logx.Info("create txs count:", len(l2TxEventMonitors))
	logx.Info("create blocks events count:", len(l2BlockEventMonitors))
	logx.Info("========== end MonitorBlocks ==========")
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

func getZkbasContractLogs(cli *_rpc.ProviderClient, zkbasContract string, startHeight, endHeight uint64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startHeight)),
		ToBlock:   big.NewInt(int64(endHeight)),
		Addresses: []common.Address{common.HexToAddress(zkbasContract)},
	}
	logs, err := cli.FilterLogs(context.Background(), query)
	if err != nil {
		logx.Error("filter logs error, err: %s", err.Error())
		return nil, err
	}
	return logs, nil
}

func getPriorityRequestCount(cli *_rpc.ProviderClient, zkbasContract string, startHeight, endHeight uint64) (int, error) {
	zkbasInstance, err := zkbas.LoadZkbasInstance(cli, zkbasContract)
	if err != nil {
		logx.Errorf("unable to load zkbas instance")
		return 0, err
	}
	priorityRequests, err := zkbasInstance.ZkbasFilterer.
		FilterNewPriorityRequest(&bind.FilterOpts{Start: startHeight, End: &endHeight})
	if err != nil {
		logx.Errorf("unable to filter deposit or lock events: %s", err.Error())
		return 0, err
	}
	priorityRequestCount := 0
	for priorityRequests.Next() {
		priorityRequestCount++
	}
	return priorityRequestCount, nil
}

func convertLogToNewPriorityRequestEvent(log types.Log) (*l2TxEventMonitor.L2TxEventMonitor, error) {
	var event zkbas.ZkbasNewPriorityRequest
	if err := ZkbasContractAbi.UnpackIntoInterface(&event, EventNameNewPriorityRequest, log.Data); err != nil {
		logx.Errorf("unpack ZkbasNewPriorityRequest err: %s", err.Error())
		return nil, err
	}
	l2TxEventMonitorInfo := &l2TxEventMonitor.L2TxEventMonitor{
		L1TxHash:        log.TxHash.Hex(),
		L1BlockHeight:   int64(log.BlockNumber),
		SenderAddress:   event.Sender.Hex(),
		RequestId:       int64(event.SerialId),
		TxType:          int64(event.TxType),
		Pubdata:         common.Bytes2Hex(event.PubData),
		ExpirationBlock: event.ExpirationBlock.Int64(),
		Status:          l2TxEventMonitor.PendingStatus,
	}
	return l2TxEventMonitorInfo, nil
}

func convertLogToBlockCommitEvent(log types.Log) (*l2BlockEventMonitor.L2BlockEventMonitor, error) {
	var event zkbas.ZkbasBlockCommit
	if err := ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, log.Data); err != nil {
		logx.Errorf("unpack ZkbasBlockCommit err: %s", err.Error())
		return nil, err
	}
	l2BlockEventMonitorInfo := &l2BlockEventMonitor.L2BlockEventMonitor{
		BlockEventType: EventTypeCommittedBlock,
		L1BlockHeight:  int64(log.BlockNumber),
		L1TxHash:       log.TxHash.Hex(),
		L2BlockHeight:  int64(event.BlockNumber),
		Status:         PendingStatusL2BlockEventMonitor,
	}
	return l2BlockEventMonitorInfo, nil
}

func convertLogToBlockVerificationEvent(log types.Log) (*l2BlockEventMonitor.L2BlockEventMonitor, error) {
	var event zkbas.ZkbasBlockVerification
	if err := ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, log.Data); err != nil {
		logx.Errorf("unpack ZkbasBlockVerification err: %s", err.Error())
		return nil, err
	}
	l2BlockEventMonitorInfo := &l2BlockEventMonitor.L2BlockEventMonitor{
		BlockEventType: EventTypeVerifiedBlock,
		L1BlockHeight:  int64(log.BlockNumber),
		L1TxHash:       log.TxHash.Hex(),
		L2BlockHeight:  int64(event.BlockNumber),
		Status:         PendingStatusL2BlockEventMonitor,
	}
	return l2BlockEventMonitorInfo, nil
}

func convertLogToBlocksRevertEvent(log types.Log) (*l2BlockEventMonitor.L2BlockEventMonitor, error) {
	var event zkbas.ZkbasBlocksRevert
	if err := ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlocksRevert, log.Data); err != nil {
		logx.Errorf("unpack ZkbasBlocksRevert err: %s", err.Error())
		return nil, err
	}
	l2BlockEventMonitorInfo := &l2BlockEventMonitor.L2BlockEventMonitor{
		BlockEventType: EventTypeRevertedBlock,
		L1BlockHeight:  int64(log.BlockNumber),
		L1TxHash:       log.TxHash.Hex(),
		L2BlockHeight:  int64(event.TotalBlocksCommitted),
		Status:         PendingStatusL2BlockEventMonitor,
	}
	return l2BlockEventMonitorInfo, nil
}
