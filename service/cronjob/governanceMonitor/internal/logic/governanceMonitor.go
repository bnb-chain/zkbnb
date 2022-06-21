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
	"fmt"
	zecreyLegend "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/zero/basic"
	asset "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/l1BlockMonitor"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/common/util"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
)

/*
	MonitorGovernanceContract: monitor layer-1 governance related events
*/
func MonitorGovernanceContract(
	cli *ProviderClient,
	startHeight int64, pendingBlocksCount uint64, maxHandledBlocksCount int64,
	governanceContract string,
	l1BlockMonitorModel L1BlockMonitorModel,
	sysconfigModel SysconfigModel,
	l2AssetInfoModel L2AssetInfoModel,
) (err error) {

	// get latest handled l1 block from database by chain id
	latestHandledBlock, err := l1BlockMonitorModel.GetLatestL1BlockMonitorByGovernance()
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
		logx.Errorf("[blockMoniter.MonitorGovernanceContract]<=>[cli.GetHeight] %s", err.Error())
		return err
	}

	// compute safe height
	safeHeight := latestHeight - pendingBlocksCount
	safeHeight = uint64(util.MinInt64(int64(safeHeight), handledHeight+maxHandledBlocksCount))

	// check if safe height > handledHeight
	if safeHeight <= uint64(handledHeight) {
		logx.Error("[l2BlockMonitor.MonitorGovernanceContract] no new blocks need to be handled")
		return nil
	}
	// filter query for Governance contract
	contractAddress := common.HexToAddress(governanceContract)
	// set filter
	logx.Infof("[MonitorGovernanceContract] fromBlock: %d, toBlock: %d", big.NewInt(handledHeight+1), big.NewInt(int64(safeHeight)))

	// block query
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(handledHeight + 1),
		ToBlock:   big.NewInt(int64(safeHeight)),
		Addresses: []common.Address{contractAddress},
	}
	// get logs from client
	logs, err := cli.FilterLogs(context.Background(), query)
	if err != nil {
		errInfo := fmt.Sprintf("[blockMoniter.MonitorGovernanceContract]<=>[cli.FilterLogs] %s", err.Error())
		logx.Error(errInfo)
		return err
	}
	// initialize L2TxEventMonitor & L2BlockEventMonitor & L1EventInfo
	var (
		l1EventInfos                  []*L1EventInfo
		l2AssetInfoMap                = make(map[string]*L2AssetInfo)
		pendingUpdateL2AssetInfoMap   = make(map[string]*L2AssetInfo)
		pendingNewSysconfigInfoMap    = make(map[string]*Sysconfig)
		pendingUpdateSysconfigInfoMap = make(map[string]*Sysconfig)
	)

	for _, vlog := range logs {
		switch vlog.Topics[0].Hex() {
		// deposit or lock event
		case GovernanceLogNewAssetSigHash.Hex():
			// parse event info
			var event zecreyLegend.GovernanceNewAsset
			err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameNewAsset, vlog.Data)
			if err != nil {
				logx.Errorf("[blockMoniter.MonitorGovernanceContract]<=>[GovernanceContractAbi.UnpackIntoInterface] %s", err.Error())
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: AddAssetEventType,
				TxHash:    vlog.TxHash.Hex(),
			}

			// get asset info by contract address
			erc20Instance, err := zecreyLegend.LoadERC20(cli, event.AssetAddress.Hex())
			if err != nil {
				logx.Errorf("[MonitorGovernanceContract] unable to load erc20: %s", err.Error())
				return err
			}
			name, err := erc20Instance.Name(basic.EmptyCallOpts())
			if err != nil {
				logx.Errorf("[MonitorGovernanceContract] unable to call: %s", err.Error())
				return err
			}
			symbol, err := erc20Instance.Symbol(basic.EmptyCallOpts())
			if err != nil {
				logx.Errorf("[MonitorGovernanceContract] unable to call: %s", err.Error())
				return err
			}
			decimals, err := erc20Instance.Decimals(basic.EmptyCallOpts())
			if err != nil {
				logx.Errorf("[MonitorGovernanceContract] unable to call: %s", err.Error())
				return err
			}

			l2AssetInfo := &L2AssetInfo{
				AssetId:     uint32(event.AssetId),
				L1Address:   event.AssetAddress.Hex(),
				AssetName:   name,
				AssetSymbol: symbol,
				Decimals:    uint32(decimals),
				Status:      asset.StatusActive,
			}
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			l2AssetInfoMap[event.AssetAddress.Hex()] = l2AssetInfo
			break
		case governanceLogNewGovernorSigHash.Hex():
			// parse event info
			var event zecreyLegend.GovernanceNewGovernor
			err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameNewGovernor, vlog.Data)
			if err != nil {
				logx.Errorf("[blockMoniter.MonitorGovernanceContract]<=>[GovernanceContractAbi.UnpackIntoInterface] %s", err.Error())
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: NewGovernorEventType,
				TxHash:    vlog.TxHash.Hex(),
			}

			configInfo := &Sysconfig{
				Name:      sysconfigName.Governor,
				Value:     event.NewGovernor.Hex(),
				ValueType: "string",
				Comment:   "governor",
			}

			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			pendingNewSysconfigInfoMap[configInfo.Name] = configInfo
			break
		case governanceLogNewAssetGovernanceSigHash.Hex():
			// parse event info
			var event zecreyLegend.GovernanceNewAssetGovernance
			err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameNewAssetGovernance, vlog.Data)
			if err != nil {
				logx.Errorf("[blockMoniter.MonitorGovernanceContract]<=>[GovernanceContractAbi.UnpackIntoInterface] %s", err.Error())
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: NewAssetGovernanceEventType,
				TxHash:    vlog.TxHash.Hex(),
			}

			configInfo := &Sysconfig{
				Name:      sysconfigName.AssetGovernanceContract,
				Value:     event.NewAssetGovernance.Hex(),
				ValueType: "string",
				Comment:   "asset governance contract",
			}

			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			pendingNewSysconfigInfoMap[configInfo.Name] = configInfo
			break
		case governanceLogValidatorStatusUpdateSigHash.Hex():
			// parse event info
			var event zecreyLegend.GovernanceValidatorStatusUpdate
			err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameValidatorStatusUpdate, vlog.Data)
			if err != nil {
				logx.Errorf("[blockMoniter.MonitorGovernanceContract]<=>[GovernanceContractAbi.UnpackIntoInterface] %s", err.Error())
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: ValidatorStatusUpdateEventType,
				TxHash:    vlog.TxHash.Hex(),
			}
			type ValidatorInfo struct {
				Address  string
				IsActive bool
			}
			// get data from db
			if pendingNewSysconfigInfoMap[sysconfigName.Validators] != nil {
				configInfo := pendingNewSysconfigInfoMap[sysconfigName.Validators]
				var validators map[string]*ValidatorInfo
				err = json.Unmarshal([]byte(configInfo.Value), &validators)
				if err != nil {
					logx.Errorf("[MonitorGovernanceContract] unable to unmarshal: %s", err.Error())
					return err
				}
				if validators[event.ValidatorAddress.Hex()] == nil {
					validators[event.ValidatorAddress.Hex()] = &ValidatorInfo{
						Address:  event.ValidatorAddress.Hex(),
						IsActive: event.IsActive,
					}
				} else {
					validators[event.ValidatorAddress.Hex()].IsActive = event.IsActive
				}
				// reset into map
				validatorBytes, err := json.Marshal(validators)
				if err != nil {
					logx.Errorf("[MonitorGovernanceContract] unable to marshal: %s", err.Error())
					return err
				}
				pendingNewSysconfigInfoMap[sysconfigName.Validators].Value = string(validatorBytes)
			} else {
				configInfo, err := sysconfigModel.GetSysconfigByName(sysconfigName.Validators)
				if err != nil {
					if err != ErrNotFound {
						logx.Errorf("[MonitorGovernanceContract] unable to get sysconfig by name: %s", err.Error())
						return err
					} else {
						validators := make(map[string]*ValidatorInfo)
						validators[event.ValidatorAddress.Hex()] = &ValidatorInfo{
							Address:  event.ValidatorAddress.Hex(),
							IsActive: event.IsActive,
						}
						validatorsBytes, err := json.Marshal(validators)
						if err != nil {
							logx.Errorf("[MonitorGovernanceContract] unable to marshal: %s", err.Error())
							return err
						}
						pendingNewSysconfigInfoMap[sysconfigName.Validators] = &Sysconfig{
							Name:      sysconfigName.Validators,
							Value:     string(validatorsBytes),
							ValueType: "map[string]*ValidatorInfo",
							Comment:   "validator info",
						}
					}
				} else {
					var validators map[string]*ValidatorInfo
					err = json.Unmarshal([]byte(configInfo.Value), &validators)
					if err != nil {
						logx.Errorf("[MonitorGovernanceContract] unable to unmarshal: %s", err.Error())
						return err
					}
					if validators[event.ValidatorAddress.Hex()] == nil {
						validators[event.ValidatorAddress.Hex()] = &ValidatorInfo{
							Address:  event.ValidatorAddress.Hex(),
							IsActive: event.IsActive,
						}
					} else {
						validators[event.ValidatorAddress.Hex()].IsActive = event.IsActive
					}
					// reset into map
					validatorBytes, err := json.Marshal(validators)
					if err != nil {
						logx.Errorf("[MonitorGovernanceContract] unable to marshal: %s", err.Error())
						return err
					}
					pendingUpdateSysconfigInfoMap[sysconfigName.Validators].Value = string(validatorBytes)
				}
			}
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			break
		case governanceLogAssetPausedUpdateSigHash.Hex():
			// parse event info
			var event zecreyLegend.GovernanceAssetPausedUpdate
			err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameAssetPausedUpdate, vlog.Data)
			if err != nil {
				logx.Errorf("[blockMoniter.MonitorGovernanceContract]<=>[GovernanceContractAbi.UnpackIntoInterface] %s", err.Error())
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: AssetPausedUpdateEventType,
				TxHash:    vlog.TxHash.Hex(),
			}

			var assetInfo *L2AssetInfo
			if l2AssetInfoMap[event.Token.Hex()] != nil {
				assetInfo = l2AssetInfoMap[event.Token.Hex()]
			} else {
				assetInfo, err = l2AssetInfoModel.GetAssetByAddress(event.Token.Hex())
				if err != nil {
					logx.Errorf("[MonitorGovernanceContract] unable to get l2 asset by address: %s", err.Error())
					return err
				}
				pendingUpdateL2AssetInfoMap[event.Token.Hex()] = assetInfo
			}
			var status uint32
			if event.Paused {
				status = asset.StatusInactive
			} else {
				status = asset.StatusActive
			}
			assetInfo.Status = status
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			break
		default:
			break
		}
	}
	// serialize into block info
	eventInfosBytes, err := json.Marshal(l1EventInfos)
	if err != nil {
		errInfo := fmt.Sprintf("[blockMoniter.MonitorGovernanceContract]<=>[json.Marshal] %s", err.Error())
		logx.Error(errInfo)
		return err
	}
	l1BlockMonitorInfo := &l1BlockMonitor.L1BlockMonitor{
		L1BlockHeight: int64(safeHeight),
		BlockInfo:     string(eventInfosBytes),
		MonitorType:   l1BlockMonitor.MonitorTypeGovernance,
	}
	var (
		l2AssetInfos                []*L2AssetInfo
		pendingUpdateL2AssetInfos   []*L2AssetInfo
		pendingNewSysconfigInfos    []*Sysconfig
		pendingUpdateSysconfigInfos []*Sysconfig
	)
	for _, l2AssetInfo := range l2AssetInfoMap {
		l2AssetInfos = append(l2AssetInfos, l2AssetInfo)
	}
	for _, pendingUpdateL2AssetInfo := range pendingUpdateL2AssetInfoMap {
		pendingUpdateL2AssetInfos = append(pendingUpdateL2AssetInfos, pendingUpdateL2AssetInfo)
	}
	for _, pendingNewSysconfigInfo := range pendingNewSysconfigInfoMap {
		pendingNewSysconfigInfos = append(pendingNewSysconfigInfos, pendingNewSysconfigInfo)
	}
	for _, pendingUpdateSysconfigInfo := range pendingUpdateSysconfigInfoMap {
		pendingUpdateSysconfigInfos = append(pendingUpdateSysconfigInfos, pendingUpdateSysconfigInfo)
	}

	logx.Infof("[MonitorGovernanceContract] l1 block info height: %v, l2 asset info size: %v, pending update l2 asset info size: %v",
		l1BlockMonitorInfo.L1BlockHeight,
		len(l2AssetInfos),
		len(pendingUpdateL2AssetInfos),
	)
	// write into database, need to use transaction
	err = l1BlockMonitorModel.CreateGovernanceMonitorInfo(
		l1BlockMonitorInfo,
		l2AssetInfos,
		pendingUpdateL2AssetInfos,
		pendingNewSysconfigInfos,
		pendingUpdateSysconfigInfos,
	)
	if err != nil {
		errInfo := fmt.Sprintf("[l1BlockMonitorModel.CreateMonitorsInfo] %s", err.Error())
		logx.Error(errInfo)
		return err
	}
	return nil
}
