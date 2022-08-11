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

package monitor

import (
	"context"
	"encoding/json"
	"math/big"

	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/zero/basic"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/asset"
	"github.com/bnb-chain/zkbas/common/model/l1SyncedBlock"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/sysConfigName"
	"github.com/bnb-chain/zkbas/common/util"
)

func (m *Monitor) MonitorGovernanceBlocks() (err error) {
	// get latest handled l1 block from database by chain id
	latestHandledBlock, err := m.L1SyncedBlockModel.GetLatestL1BlockByType(l1SyncedBlock.TypeGovernance)
	var handledHeight int64
	if err != nil {
		if err == errorcode.DbErrNotFound {
			handledHeight = m.Config.ChainConfig.StartL1BlockHeight
		} else {
			logx.Errorf("[l1SyncedBlockModel.GetLatestL1BlockByType]: %s", err.Error())
			return err
		}
	} else {
		handledHeight = latestHandledBlock.L1BlockHeight
	}
	// get latest l1 block height(latest height - pendingBlocksCount)
	latestHeight, err := m.cli.GetHeight()
	if err != nil {
		logx.Errorf("get l1 block height err: %s", err.Error())
		return err
	}
	// compute safe height
	safeHeight := latestHeight - m.Config.ChainConfig.ConfirmBlocksCount
	safeHeight = uint64(util.MinInt64(int64(safeHeight), handledHeight+m.Config.ChainConfig.MaxHandledBlocksCount))
	// check if safe height > handledHeight
	if safeHeight <= uint64(handledHeight) {
		return nil
	}
	contractAddress := common.HexToAddress(m.governanceContractAddress)
	logx.Infof("fromBlock: %d, toBlock: %d", big.NewInt(handledHeight+1), big.NewInt(int64(safeHeight)))
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(handledHeight + 1),
		ToBlock:   big.NewInt(int64(safeHeight)),
		Addresses: []common.Address{contractAddress},
	}
	logs, err := m.cli.FilterLogs(context.Background(), query)
	if err != nil {
		logx.Errorf("FilterLogs err: %s", err.Error())
		return err
	}
	var (
		l1EventInfos              []*L1EventInfo
		l2AssetInfoMap            = make(map[string]*asset.Asset)
		pendingUpdateL2AssetMap   = make(map[string]*asset.Asset)
		pendingNewSysConfigMap    = make(map[string]*sysconfig.SysConfig)
		pendingUpdateSysConfigMap = make(map[string]*sysconfig.SysConfig)
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
			erc20Instance, err := zkbas.LoadERC20(m.cli, event.AssetAddress.Hex())
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
			l2AssetInfo := &asset.Asset{
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
			configInfo := &sysconfig.SysConfig{
				Name:      sysConfigName.Governor,
				Value:     event.NewGovernor.Hex(),
				ValueType: "string",
				Comment:   "governor",
			}
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			pendingNewSysConfigMap[configInfo.Name] = configInfo
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
			configInfo := &sysconfig.SysConfig{
				Name:      sysConfigName.AssetGovernanceContract,
				Value:     event.NewAssetGovernance.Hex(),
				ValueType: "string",
				Comment:   "asset governance contract",
			}
			// set into array
			l1EventInfos = append(l1EventInfos, l1EventInfo)
			pendingNewSysConfigMap[configInfo.Name] = configInfo
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
			if pendingNewSysConfigMap[sysConfigName.Validators] != nil {
				configInfo := pendingNewSysConfigMap[sysConfigName.Validators]
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
				pendingNewSysConfigMap[sysConfigName.Validators].Value = string(validatorBytes)
			} else {
				configInfo, err := m.SysConfigModel.GetSysConfigByName(sysConfigName.Validators)
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
						pendingNewSysConfigMap[sysConfigName.Validators] = &sysconfig.SysConfig{
							Name:      sysConfigName.Validators,
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
					pendingUpdateSysConfigMap[sysConfigName.Validators].Value = string(validatorBytes)
				}
			}
			l1EventInfos = append(l1EventInfos, l1EventInfo)
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
			var assetInfo *asset.Asset
			if l2AssetInfoMap[event.Token.Hex()] != nil {
				assetInfo = l2AssetInfoMap[event.Token.Hex()]
			} else {
				assetInfo, err = m.L2AssetModel.GetAssetByAddress(event.Token.Hex())
				if err != nil {
					logx.Errorf("unable to get l2 asset by address, err: %s", err.Error())
					return err
				}
				pendingUpdateL2AssetMap[event.Token.Hex()] = assetInfo
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
		default:
		}
	}
	// serialize into block info
	eventInfosBytes, err := json.Marshal(l1EventInfos)
	if err != nil {
		logx.Errorf("marshal l1 events error, err: %s", err.Error())
		return err
	}
	syncedBlock := &l1SyncedBlock.L1SyncedBlock{
		L1BlockHeight: int64(safeHeight),
		BlockInfo:     string(eventInfosBytes),
		Type:          l1SyncedBlock.TypeGovernance,
	}
	var (
		l2AssetInfos                []*asset.Asset
		pendingUpdateL2AssetInfos   []*asset.Asset
		pendingNewSysconfigInfos    []*sysconfig.SysConfig
		pendingUpdateSysconfigInfos []*sysconfig.SysConfig
	)
	for _, l2AssetInfo := range l2AssetInfoMap {
		l2AssetInfos = append(l2AssetInfos, l2AssetInfo)
	}
	for _, pendingUpdateL2AssetInfo := range pendingUpdateL2AssetMap {
		pendingUpdateL2AssetInfos = append(pendingUpdateL2AssetInfos, pendingUpdateL2AssetInfo)
	}
	for _, pendingNewSysconfigInfo := range pendingNewSysConfigMap {
		pendingNewSysconfigInfos = append(pendingNewSysconfigInfos, pendingNewSysconfigInfo)
	}
	for _, pendingUpdateSysconfigInfo := range pendingUpdateSysConfigMap {
		pendingUpdateSysconfigInfos = append(pendingUpdateSysconfigInfos, pendingUpdateSysconfigInfo)
	}
	logx.Infof("l1 block info height: %v, l2 asset info size: %v, pending update l2 asset info size: %v",
		syncedBlock.L1BlockHeight,
		len(l2AssetInfos),
		len(pendingUpdateL2AssetInfos),
	)
	if err = m.L1SyncedBlockModel.CreateGovernanceBlock(syncedBlock, l2AssetInfos,
		pendingUpdateL2AssetInfos, pendingNewSysconfigInfos, pendingUpdateSysconfigInfos); err != nil {
		logx.Errorf("store governance monitor info error, err: %s", err.Error())
		return err
	}
	return nil
}
