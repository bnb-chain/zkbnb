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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/l1BlockMonitor"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/util"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-zero/model/l2TxEventMonitor"
	zecreyLegend "github.com/zecrey-labs/zecrey-eth-rpc/zecrey/core/zecrey-legend"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
)

/*
	MonitorBlocks: monitor layer-1 block events
*/
func MonitorBlocks(
	cli *ProviderClient,
	nativeChainId *big.Int,
	startHeight int64, pendingBlocksCount uint64, maxHandledBlocksCount int64,
	zecreyContract string,
	l1BlockMonitorModel L1BlockMonitorModel,
) (err error) {

	// get latest handled l1 block from database by chain id
	latestHandledBlock, err := l1BlockMonitorModel.GetLatestL1BlockMonitor()
	var handledHeight int64
	if err != nil {
		if err == ErrNotFound {
			handledHeight = startHeight
		} else {
			logx.Errorf("[l1BlockMonitorModel.GetLatestL1BlockMonitor]: %s", err.Error())
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

	// compute safe height
	safeHeight := latestHeight - pendingBlocksCount
	safeHeight = uint64(util.MinInt64(int64(safeHeight), handledHeight+maxHandledBlocksCount))

	// check if safe height > handledHeight
	if safeHeight <= uint64(handledHeight) {
		logx.Error("[l2BlockMonitor.MonitorBlocks] no new blocks need to be handled")
		return nil
	}

	// filter query for Zecrey contract
	contractAddress := common.HexToAddress(zecreyContract)
	// set filter
	logx.Infof("[MonitorBlocks] fromBlock: %d, toBlock: %d", big.NewInt(handledHeight+1), big.NewInt(int64(safeHeight)))

	zecreyInstance, err := zecreyLegend.LoadZecreyLegendInstance(cli, zecreyContract)
	if err != nil {
		logx.Errorf("[MonitorBlocks] unable to load zecrey instance")
		return err
	}
	// deposit or lock logs
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
	// block query
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(handledHeight + 1),
		ToBlock:   big.NewInt(int64(safeHeight)),
		Addresses: []common.Address{contractAddress},
	}
	// get logs from client
	logs, err := cli.FilterLogs(context.Background(), query)
	if err != nil {
		errInfo := fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[cli.FilterLogs] %s", err.Error())
		logx.Error(errInfo)
		return err
	}
	// initialize L2TxEventMonitor & L2BlockEventMonitor & L1EventInfo
	var (
		l1EventInfos         []*L1EventInfo
		l2TxEventMonitors    []*L2TxEventMonitor
		l2BlockEventMonitors []*L2BlockEventMonitor
		//LondonSigner         = types.NewLondonSigner(nativeChainId)
	)
	for _, vlog := range logs {
		switch vlog.Topics[0].Hex() {
		// deposit or lock event
		case zecreyLogNewPriorityRequestSigHash.Hex():
			priorityRequestCountCheck++
			// parse event info
			var event zecreyLegend.ZecreyLegendNewPriorityRequest
			err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameNewPriorityRequest, vlog.Data)
			if err != nil {
				logx.Errorf("[blockMoniter.MonitorBlocks]<=>[ZecreyContractAbi.UnpackIntoInterface] %s", err.Error())
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: EventTypeNewPriorityRequest,
				TxHash:    vlog.TxHash.Hex(),
			}

			// compute balance delta
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
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			l2TxEventMonitors = append(l2TxEventMonitors, l2TxEventMonitorInfo)
			break
		case ZecreyLogWithdrawalSigHash.Hex():
			break
		case ZecreyLogWithdrawalPendingSigHash.Hex():
			break
		case ZecreyLogBlockCommitSigHash.Hex():
			// parse event info
			var event zecreyLegend.ZecreyLegendBlockCommit
			err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, vlog.Data)
			if err != nil {
				errInfo := fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[ZecreyContractAbi.UnpackIntoInterface] %s", err.Error())
				logx.Error(errInfo)
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: EventTypeCommittedBlock,
				TxHash:    vlog.TxHash.Hex(),
			}
			l2BlockEventMonitorInfo := &L2BlockEventMonitor{
				BlockEventType: EventTypeCommittedBlock,
				L1BlockHeight:  int64(vlog.BlockNumber),
				L1TxHash:       vlog.TxHash.Hex(),
				L2BlockHeight:  int64(event.BlockNumber),
				Status:         PendingStatusL2BlockEventMonitor,
			}
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			l2BlockEventMonitors = append(l2BlockEventMonitors, l2BlockEventMonitorInfo)
			break
		case ZecreyLogBlockVerificationSigHash.Hex():
			// parse event info
			var event zecreyLegend.ZecreyLegendBlockVerification
			err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data)
			if err != nil {
				errInfo := fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[ZecreyContractAbi.UnpackIntoInterface] %s", err.Error())
				logx.Error(errInfo)
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: EventTypeVerifiedBlock,
				TxHash:    vlog.TxHash.Hex(),
			}
			l2BlockEventMonitorInfo := &L2BlockEventMonitor{
				BlockEventType: EventTypeVerifiedBlock,
				L1BlockHeight:  int64(vlog.BlockNumber),
				L1TxHash:       vlog.TxHash.Hex(),
				L2BlockHeight:  int64(event.BlockNumber),
				Status:         PendingStatusL2BlockEventMonitor,
			}
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			l2BlockEventMonitors = append(l2BlockEventMonitors, l2BlockEventMonitorInfo)
			break
		case ZecreyLogBlocksRevertSigHash.Hex():
			// parse event info
			var event zecreyLegend.ZecreyLegendBlocksRevert
			err = ZecreyContractAbi.UnpackIntoInterface(&event, EventNameBlocksRevert, vlog.Data)
			if err != nil {
				errInfo := fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[ZecreyContractAbi.UnpackIntoInterface] %s", err.Error())
				logx.Error(errInfo)
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: EventTypeRevertedBlock,
				TxHash:    vlog.TxHash.Hex(),
			}
			l2BlockEventMonitorInfo := &L2BlockEventMonitor{
				BlockEventType: EventTypeRevertedBlock,
				L1BlockHeight:  int64(vlog.BlockNumber),
				L1TxHash:       vlog.TxHash.Hex(),
				L2BlockHeight:  int64(event.TotalBlocksCommitted),
				Status:         PendingStatusL2BlockEventMonitor,
			}
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			l2BlockEventMonitors = append(l2BlockEventMonitors, l2BlockEventMonitorInfo)
			break
		default:
			break
		}
	}
	// check deposit or lock events
	if priorityRequestCount != priorityRequestCountCheck {
		logx.Errorf("[MonitorBlocks] new priority requests events not match, try it again")
		return errors.New("[MonitorBlocks] new priority requests events not match, try it again")
	}
	// serialize into block info
	eventInfosBytes, err := json.Marshal(l1EventInfos)
	if err != nil {
		errInfo := fmt.Sprintf("[blockMoniter.MonitorBlocks]<=>[json.Marshal] %s", err.Error())
		logx.Error(errInfo)
		return err
	}
	l1BlockMonitorInfo := &l1BlockMonitor.L1BlockMonitor{
		L1BlockHeight: int64(safeHeight),
		BlockInfo:     string(eventInfosBytes),
	}
	// write into database, need to use transaction
	err = l1BlockMonitorModel.CreateMonitorsInfo(l1BlockMonitorInfo, l2TxEventMonitors, l2BlockEventMonitors)
	if err != nil {
		errInfo := fmt.Sprintf("[l1BlockMonitorModel.CreateMonitorsInfo] %s", err.Error())
		logx.Error(errInfo)
		return err
	}
	logx.Info("[MonitorBlocks] create txs count:", len(l2TxEventMonitors))
	logx.Info("[MonitorBlocks] create blocks events count:", len(l2BlockEventMonitors))
	return nil
}
