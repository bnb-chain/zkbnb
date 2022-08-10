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
	"math/big"

	"github.com/bnb-chain/zkbas/common/model/sysconfig"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/zero/basic"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	asset "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/l1BlockMonitor"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/errorcode"
)

/*
	MonitorGovernanceContract: monitor layer-1 governance related events
*/
func MonitorGovernanceContract(cli *_rpc.ProviderClient, startHeight int64, pendingBlocksCount uint64, maxHandledBlocksCount int64,
	governanceContract string, l1BlockMonitorModel l1BlockMonitor.L1BlockMonitorModel, sysconfigModel sysconfig.SysconfigModel, l2AssetInfoModel asset.AssetInfoModel) (err error) {
	logx.Info("========================= start MonitorGovernanceContract =========================")
	// get latest handled l1 block from database by chain id
	latestHandledBlock, err := l1BlockMonitorModel.GetLatestL1BlockMonitorByGovernance()
	var handledHeight int64
	if err != nil {
		if err == errorcode.DbErrNotFound {
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
		logx.Errorf("get l1 block height err: %s", err.Error())
		return err
	}
	// compute safe height
	safeHeight := latestHeight - pendingBlocksCount
	safeHeight = uint64(util.MinInt64(int64(safeHeight), handledHeight+maxHandledBlocksCount))
	// check if safe height > handledHeight
	if safeHeight <= uint64(handledHeight) {
		return nil
	}
	contractAddress := common.HexToAddress(governanceContract)
	logx.Infof("fromBlock: %d, toBlock: %d", big.NewInt(handledHeight+1), big.NewInt(int64(safeHeight)))
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(handledHeight + 1),
		ToBlock:   big.NewInt(int64(safeHeight)),
		Addresses: []common.Address{contractAddress},
	}
	logs, err := cli.FilterLogs(context.Background(), query)
	if err != nil {
		logx.Errorf("FilterLogs err: %s", err.Error())
		return err
	}
	var (
		l1EventInfos                  []*L1EventInfo
		l2AssetInfoMap                = make(map[string]*asset.AssetInfo)
		pendingUpdateL2AssetInfoMap   = make(map[string]*asset.AssetInfo)
		pendingNewSysconfigInfoMap    = make(map[string]*sysconfig.Sysconfig)
		pendingUpdateSysconfigInfoMap = make(map[string]*sysconfig.Sysconfig)
	)
	for _, vlog := range logs {
		switch vlog.Topics[0].Hex() {
		case governanceLogNewAssetSigHash.Hex():
			var event zkbas.GovernanceNewAsset
			if err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameNewAsset, vlog.Data); err != nil {
				logx.Errorf("UnpackIntoInterface err: %s", err.Error())
				return err
			}
			l1EventInfo := &L1EventInfo{
				EventType: EventTypeAddAsset,
				TxHash:    vlog.TxHash.Hex(),
			}
			// get asset info by contract address
			erc20Instance, err := zkbas.LoadERC20(cli, event.AssetAddress.Hex())
			if err != nil {
				logx.Errorf("LoadERC20 err: %s", err.Error())
				return err
			}
			name, err := erc20Instance.Name(basic.EmptyCallOpts())
			if err != nil {
				logx.Errorf("erc20Instance.Name err: %s", err.Error())
				return err
			}
			symbol, err := erc20Instance.Symbol(basic.EmptyCallOpts())
			if err != nil {
				logx.Errorf("erc20Instance.Symbol err: %s", err.Error())
				return err
			}
			decimals, err := erc20Instance.Decimals(basic.EmptyCallOpts())
			if err != nil {
				logx.Errorf("erc20Instance.Decimals err: %s", err.Error())
				return err
			}
			l2AssetInfo := &asset.AssetInfo{
				AssetId:     uint32(event.AssetId),
				L1Address:   event.AssetAddress.Hex(),
				AssetName:   name,
				AssetSymbol: symbol,
				Decimals:    uint32(decimals),
				Status:      asset.StatusActive,
			}
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			l2AssetInfoMap[event.AssetAddress.Hex()] = l2AssetInfo
		case governanceLogNewGovernorSigHash.Hex():
			// parse event info
			var event zkbas.GovernanceNewGovernor
			if err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameNewGovernor, vlog.Data); err != nil {
				logx.Errorf("UnpackIntoInterface err: %s", err.Error())
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: EventTypeNewGovernor,
				TxHash:    vlog.TxHash.Hex(),
			}
			configInfo := &sysconfig.Sysconfig{
				Name:      sysconfigName.Governor,
				Value:     event.NewGovernor.Hex(),
				ValueType: "string",
				Comment:   "governor",
			}
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			pendingNewSysconfigInfoMap[configInfo.Name] = configInfo
		case governanceLogNewAssetGovernanceSigHash.Hex():
			// parse event info
			var event zkbas.GovernanceNewAssetGovernance
			err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameNewAssetGovernance, vlog.Data)
			if err != nil {
				logx.Errorf("UnpackIntoInterface err: %s", err.Error())
				return err
			}
			l1EventInfo := &L1EventInfo{
				EventType: EventTypeNewAssetGovernance,
				TxHash:    vlog.TxHash.Hex(),
			}
			configInfo := &sysconfig.Sysconfig{
				Name:      sysconfigName.AssetGovernanceContract,
				Value:     event.NewAssetGovernance.Hex(),
				ValueType: "string",
				Comment:   "asset governance contract",
			}
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			pendingNewSysconfigInfoMap[configInfo.Name] = configInfo
		case governanceLogValidatorStatusUpdateSigHash.Hex():
			// parse event info
			var event zkbas.GovernanceValidatorStatusUpdate
			if err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameValidatorStatusUpdate, vlog.Data); err != nil {
				logx.Errorf("unpack GovernanceValidatorStatusUpdate error, err: %s", err.Error())
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: EventTypeValidatorStatusUpdate,
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
					logx.Errorf("unable to unmarshal: %s", err.Error())
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
				validatorBytes, err := json.Marshal(validators)
				if err != nil {
					logx.Errorf("unable to marshal validators: %s", err.Error())
					return err
				}
				pendingNewSysconfigInfoMap[sysconfigName.Validators].Value = string(validatorBytes)
			} else {
				configInfo, err := sysconfigModel.GetSysconfigByName(sysconfigName.Validators)
				if err != nil {
					if err != errorcode.DbErrNotFound {
						logx.Errorf("unable to get sys config by name: %s", err.Error())
						return err
					} else {
						validators := make(map[string]*ValidatorInfo)
						validators[event.ValidatorAddress.Hex()] = &ValidatorInfo{
							Address:  event.ValidatorAddress.Hex(),
							IsActive: event.IsActive,
						}
						validatorsBytes, err := json.Marshal(validators)
						if err != nil {
							logx.Errorf("unable to marshal validators: %s", err.Error())
							return err
						}
						pendingNewSysconfigInfoMap[sysconfigName.Validators] = &sysconfig.Sysconfig{
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
						logx.Errorf("unable to unmarshal validators: %s", err.Error())
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
						logx.Errorf("unable to marshal validators: %s", err.Error())
						return err
					}
					pendingUpdateSysconfigInfoMap[sysconfigName.Validators].Value = string(validatorBytes)
				}
			}
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			break
		case governanceLogAssetPausedUpdateSigHash.Hex():
			// parse event info
			var event zkbas.GovernanceAssetPausedUpdate
			err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameAssetPausedUpdate, vlog.Data)
			if err != nil {
				logx.Errorf("unpack GovernanceAssetPausedUpdate error, err: %s", err.Error())
				return err
			}
			// set up database info
			l1EventInfo := &L1EventInfo{
				EventType: EventTypeAssetPausedUpdate,
				TxHash:    vlog.TxHash.Hex(),
			}
			var assetInfo *asset.AssetInfo
			if l2AssetInfoMap[event.Token.Hex()] != nil {
				assetInfo = l2AssetInfoMap[event.Token.Hex()]
			} else {
				assetInfo, err = l2AssetInfoModel.GetAssetByAddress(event.Token.Hex())
				if err != nil {
					logx.Errorf("unable to get l2 asset by address, err: %s", err.Error())
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
		logx.Errorf("marshal l1 events error, err: %s", err.Error())
		return err
	}
	l1BlockMonitorInfo := &l1BlockMonitor.L1BlockMonitor{
		L1BlockHeight: int64(safeHeight),
		BlockInfo:     string(eventInfosBytes),
		MonitorType:   l1BlockMonitor.MonitorTypeGovernance,
	}
	var (
		l2AssetInfos                []*asset.AssetInfo
		pendingUpdateL2AssetInfos   []*asset.AssetInfo
		pendingNewSysconfigInfos    []*sysconfig.Sysconfig
		pendingUpdateSysconfigInfos []*sysconfig.Sysconfig
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
	logx.Infof("l1 block info height: %v, l2 asset info size: %v, pending update l2 asset info size: %v",
		l1BlockMonitorInfo.L1BlockHeight,
		len(l2AssetInfos),
		len(pendingUpdateL2AssetInfos),
	)
	if err = l1BlockMonitorModel.CreateGovernanceMonitorInfo(l1BlockMonitorInfo, l2AssetInfos,
		pendingUpdateL2AssetInfos, pendingNewSysconfigInfos, pendingUpdateSysconfigInfos); err != nil {
		logx.Errorf("store governance monitor info error, err: %s", err.Error())
		return err
	}
	logx.Info("========================= end MonitorGovernanceContract =========================")
	return nil
}
