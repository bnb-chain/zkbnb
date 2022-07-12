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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	zecreyLegend "github.com/zecrey-labs/zecrey-eth-rpc/zecrey/core/zecrey-legend"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zecrey-labs/zecrey-legend/common/model/l1BlockMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/util"
)

/*
	MonitorBlocks: monitor layer-1 block events
*/
func MonitorBlocks(cli *ProviderClient, startHeight int64, pendingBlocksCount uint64, maxHandledBlocksCount int64, zecreyContract string, l1BlockMonitorModel L1BlockMonitorModel) (err error) {
	latestHandledBlock, err := l1BlockMonitorModel.GetLatestL1BlockMonitorByBlock()
	logx.Errorf("========== start MonitorBlocks ==========")
	var handledHeight int64
	if err != nil {
		if err == ErrNotFound {
			handledHeight = startHeight
		} else {
			logx.Errorf("[l1BlockMonitorModel.GetLatestL1BlockMonitorByBlock]: %s", err.Error())
			return err
		}
	} else {
		handledHeight = latestHandledBlock.L1BlockHeight
	}
	// get latest l1 block height(latest height - pendingBlocksCount)
	latestHeight, err := cli.GetHeight()
	if err != nil {
		logx.Errorf("[blockMoniter.MonitorBlocks]<=>[cli.GetHeight] %s", err.Error())
		return err
	}
	safeHeight := latestHeight - pendingBlocksCount
	safeHeight = uint64(util.MinInt64(int64(safeHeight), handledHeight+maxHandledBlocksCount))
	if safeHeight <= uint64(handledHeight) {
		logx.Error("[l2BlockMonitor.MonitorBlocks] no new blocks need to be handled")
		return nil
	}
	contractAddress := common.HexToAddress(zecreyContract)
	logx.Infof("[MonitorBlocks] fromBlock: %d, toBlock: %d", big.NewInt(handledHeight+1), big.NewInt(int64(safeHeight)))
	zecreyInstance, err := zecreyLegend.LoadZecreyLegendInstance(cli, zecreyContract)
	if err != nil {
		logx.Errorf("[MonitorBlocks] unable to load zecrey instance")
		return err
	}
	priorityRequests, err := zecreyInstance.ZecreyLegendFilterer.
		FilterNewPriorityRequest(&bind.FilterOpts{Start: uint64(handledHeight + 1), End: &safeHeight})
	if err != nil {
		logx.Errorf("[MonitorBlocks] unable to filter deposit or lock events: %s", err.Error())
		return err
	}
	priorityRequestCount, priorityRequestCountCheck := 0, 0
	for priorityRequests.Next() {
		priorityRequestCount++
	}
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(handledHeight + 1),
		ToBlock:   big.NewInt(int64(safeHeight)),
		Addresses: []common.Address{contractAddress},
	}
	logs, err := cli.FilterLogs(context.Background(), query)
	if err != nil {
		errInfo := fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[cli.FilterLogs] %s", err.Error())
		logx.Error(errInfo)
		return err
	}
	var (
		l1EventInfos         []*L1EventInfo
		l2TxEventMonitors    []*L2TxEventMonitor
		l2BlockEventMonitors []*L2BlockEventMonitor
	)
	for _, vlog := range logs {
		l1EventInfo := &L1EventInfo{
			TxHash: vlog.TxHash.Hex(),
		}
		switch vlog.Topics[0].Hex() {
		case zecreyLogNewPriorityRequestSigHash.Hex():
			priorityRequestCountCheck++
			var event zecreyLegend.ZecreyLegendNewPriorityRequest
			if err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameNewPriorityRequest, vlog.Data); err != nil {
				logx.Errorf("[blockMoniter.MonitorBlocks]<=>[ZecreyContractAbi.UnpackIntoInterface] %v", err)
				return err
			}
			l1EventInfo.EventType = EventTypeNewPriorityRequest
			l2TxEventMonitorInfo := &L2TxEventMonitor{
				L1TxHash:        vlog.TxHash.Hex(),
				L1BlockHeight:   int64(vlog.BlockNumber),
				SenderAddress:   event.Sender.Hex(),
				RequestId:       int64(event.SerialId),
				TxType:          int64(event.TxType),
				Pubdata:         common.Bytes2Hex(event.PubData),
				ExpirationBlock: event.ExpirationBlock.Int64(),
				Status:          l2TxEventMonitor.PendingStatus,
			}
			l2TxEventMonitors = append(l2TxEventMonitors, l2TxEventMonitorInfo)
		case zecreyLogWithdrawalSigHash.Hex():
		case zecreyLogWithdrawalPendingSigHash.Hex():
		case zecreyLogBlockCommitSigHash.Hex():
			var event zecreyLegend.ZecreyLegendBlockCommit
			if err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, vlog.Data); err != nil {
				errInfo := fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[ZecreyContractAbi.UnpackIntoInterface] %s", err.Error())
				logx.Error(errInfo)
				return err
			}
			l1EventInfo.EventType = EventTypeCommittedBlock
			l2BlockEventMonitorInfo := &L2BlockEventMonitor{
				BlockEventType: EventTypeCommittedBlock,
				L1BlockHeight:  int64(vlog.BlockNumber),
				L1TxHash:       vlog.TxHash.Hex(),
				L2BlockHeight:  int64(event.BlockNumber),
				Status:         PendingStatusL2BlockEventMonitor,
			}
			l2BlockEventMonitors = append(l2BlockEventMonitors, l2BlockEventMonitorInfo)
		case zecreyLogBlockVerificationSigHash.Hex():
			var event zecreyLegend.ZecreyLegendBlockVerification
			if err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data); err != nil {
				errInfo := fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[ZecreyContractAbi.UnpackIntoInterface] %s", err.Error())
				logx.Error(errInfo)
				return err
			}
			l1EventInfo.EventType = EventTypeVerifiedBlock
			l2BlockEventMonitorInfo := &L2BlockEventMonitor{
				BlockEventType: EventTypeVerifiedBlock,
				L1BlockHeight:  int64(vlog.BlockNumber),
				L1TxHash:       vlog.TxHash.Hex(),
				L2BlockHeight:  int64(event.BlockNumber),
				Status:         PendingStatusL2BlockEventMonitor,
			}
			l2BlockEventMonitors = append(l2BlockEventMonitors, l2BlockEventMonitorInfo)
		case zecreyLogBlocksRevertSigHash.Hex():
			var event zecreyLegend.ZecreyLegendBlocksRevert
			if err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameBlocksRevert, vlog.Data); err != nil {
				errInfo := fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[ZecreyContractAbi.UnpackIntoInterface] %s", err.Error())
				logx.Error(errInfo)
				return err
			}
			l1EventInfo.EventType = EventTypeRevertedBlock
			l2BlockEventMonitorInfo := &L2BlockEventMonitor{
				BlockEventType: EventTypeRevertedBlock,
				L1BlockHeight:  int64(vlog.BlockNumber),
				L1TxHash:       vlog.TxHash.Hex(),
				L2BlockHeight:  int64(event.TotalBlocksCommitted),
				Status:         PendingStatusL2BlockEventMonitor,
			}
			l2BlockEventMonitors = append(l2BlockEventMonitors, l2BlockEventMonitorInfo)
		default:
		}
		l1EventInfos = append(l1EventInfos, l1EventInfo)
	}
	if priorityRequestCount != priorityRequestCountCheck {
		logx.Errorf("[MonitorBlocks] new priority requests events not match, try it again")
		return errors.New("[MonitorBlocks] new priority requests events not match, try it again")
	}
	eventInfosBytes, err := json.Marshal(l1EventInfos)
	if err != nil {
		logx.Error(fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[json.Marshal] %s", err.Error()))
		return err
	}
	l1BlockMonitorInfo := &l1BlockMonitor.L1BlockMonitor{
		L1BlockHeight: int64(safeHeight),
		BlockInfo:     string(eventInfosBytes),
		MonitorType:   l1BlockMonitor.MonitorTypeBlock,
	}
	if err = l1BlockMonitorModel.CreateMonitorsInfo(l1BlockMonitorInfo, l2TxEventMonitors, l2BlockEventMonitors); err != nil {
		errInfo := fmt.Sprintf("[l1BlockMonitorModel.CreateMonitorsInfo] %s", err.Error())
		logx.Error(errInfo)
		return err
	}
	logx.Info("[MonitorBlocks] create txs count:", len(l2TxEventMonitors))
	logx.Info("[MonitorBlocks] create blocks events count:", len(l2BlockEventMonitors))
	logx.Errorf("========== end MonitorBlocks ==========")
	return nil
}
