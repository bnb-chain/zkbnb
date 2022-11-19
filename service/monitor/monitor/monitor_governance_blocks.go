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
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/types"
)

func (m *Monitor) getNewL2Asset(event zkbnb.GovernanceNewAsset) (*asset.Asset, error) {
	// get asset info by contract address
	erc20Instance, err := zkbnb.LoadERC20(m.cli, event.AssetAddress.Hex())
	if err != nil {
		return nil, err
	}
	name, err := erc20Instance.Name(EmptyCallOpts())
	if err != nil {
		return nil, err
	}
	symbol, err := erc20Instance.Symbol(EmptyCallOpts())
	if err != nil {
		return nil, err
	}
	decimals, err := erc20Instance.Decimals(EmptyCallOpts())
	if err != nil {
		return nil, err
	}
	l2Asset := &asset.Asset{
		AssetId:     uint32(event.AssetId),
		L1Address:   event.AssetAddress.Hex(),
		AssetName:   name,
		AssetSymbol: strings.ToUpper(symbol),
		Decimals:    uint32(decimals),
		Status:      asset.StatusActive,
	}

	return l2Asset, nil
}

type GovernancePendingChanges struct {
	l2AssetMap                map[string]*asset.Asset
	pendingUpdateL2AssetMap   map[string]*asset.Asset
	pendingNewSysConfigMap    map[string]*sysconfig.SysConfig
	pendingUpdateSysConfigMap map[string]*sysconfig.SysConfig
}

func NewGovernancePendingChanges() *GovernancePendingChanges {
	return &GovernancePendingChanges{
		l2AssetMap:                make(map[string]*asset.Asset),
		pendingUpdateL2AssetMap:   make(map[string]*asset.Asset),
		pendingNewSysConfigMap:    make(map[string]*sysconfig.SysConfig),
		pendingUpdateSysConfigMap: make(map[string]*sysconfig.SysConfig),
	}
}

func (m *Monitor) MonitorGovernanceBlocks() (err error) {
	startHeight, endHeight, err := m.getBlockRangeToSync(l1syncedblock.TypeGovernance)
	if err != nil {
		logx.Errorf("get block range to sync error, err: %s", err.Error())
		return err
	}
	if endHeight < startHeight {
		logx.Infof("no blocks to sync, startHeight: %d, endHeight: %d", startHeight, endHeight)
		return nil
	}

	logx.Infof("syncing governance l1 blocks from %d to %d", big.NewInt(startHeight), big.NewInt(endHeight))
	contractAddress := common.HexToAddress(m.governanceContractAddress)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(startHeight),
		ToBlock:   big.NewInt(endHeight),
		Addresses: []common.Address{contractAddress},
	}
	logs, err := m.cli.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to query logs through rpc client: %v", err)
	}

	l1Events := make([]*L1Event, 0, len(logs))
	pendingChanges := NewGovernancePendingChanges()

	for _, vlog := range logs {
		switch vlog.Topics[0].Hex() {
		case governanceLogNewAssetSigHash.Hex():
			var event zkbnb.GovernanceNewAsset
			if err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameNewAsset, vlog.Data); err != nil {
				return fmt.Errorf("unpackIntoInterface err: %v", err)
			}
			l1EventInfo := &L1Event{
				EventType: EventTypeAddAsset,
				TxHash:    vlog.TxHash.Hex(),
			}
			newL2Asset, err := m.getNewL2Asset(event)
			if err != nil {
				logx.Infof("get new l2 asset error, err: %s", err.Error())
				return err
			}

			l1Events = append(l1Events, l1EventInfo)
			pendingChanges.l2AssetMap[event.AssetAddress.Hex()] = newL2Asset
		case governanceLogNewGovernorSigHash.Hex():
			// parse event info
			var event zkbnb.GovernanceNewGovernor
			if err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameNewGovernor, vlog.Data); err != nil {
				return fmt.Errorf("unpackIntoInterface err: %v", err)
			}
			// set up database info
			l1EventInfo := &L1Event{
				EventType: EventTypeNewGovernor,
				TxHash:    vlog.TxHash.Hex(),
			}
			configInfo := &sysconfig.SysConfig{
				Name:      types.Governor,
				Value:     event.NewGovernor.Hex(),
				ValueType: "string",
				Comment:   "governor",
			}
			// set into array
			l1Events = append(l1Events, l1EventInfo)
			pendingChanges.pendingNewSysConfigMap[configInfo.Name] = configInfo
		case governanceLogNewAssetGovernanceSigHash.Hex():
			// parse event info
			var event zkbnb.GovernanceNewAssetGovernance
			err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameNewAssetGovernance, vlog.Data)
			if err != nil {
				return fmt.Errorf("unpackIntoInterface err: %v", err)
			}
			l1EventInfo := &L1Event{
				EventType: EventTypeNewAssetGovernance,
				TxHash:    vlog.TxHash.Hex(),
			}
			configInfo := &sysconfig.SysConfig{
				Name:      types.AssetGovernanceContract,
				Value:     event.NewAssetGovernance.Hex(),
				ValueType: "string",
				Comment:   "asset governance contract",
			}
			// set into array
			l1Events = append(l1Events, l1EventInfo)
			pendingChanges.pendingNewSysConfigMap[configInfo.Name] = configInfo
		case governanceLogValidatorStatusUpdateSigHash.Hex():
			// parse event info
			var event zkbnb.GovernanceValidatorStatusUpdate
			if err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameValidatorStatusUpdate, vlog.Data); err != nil {
				return fmt.Errorf("unpack GovernanceValidatorStatusUpdate, err: %v", err)
			}
			// set up database info
			l1EventInfo := &L1Event{
				EventType: EventTypeValidatorStatusUpdate,
				TxHash:    vlog.TxHash.Hex(),
			}
			l1Events = append(l1Events, l1EventInfo)

			err = m.processValidatorUpdate(event, pendingChanges)
			if err != nil {
				return err
			}
		case governanceLogAssetPausedUpdateSigHash.Hex():
			// parse event info
			var event zkbnb.GovernanceAssetPausedUpdate
			err = GovernanceContractAbi.UnpackIntoInterface(&event, EventNameAssetPausedUpdate, vlog.Data)
			if err != nil {
				return fmt.Errorf("unpack GovernanceAssetPausedUpdate failed, err: %v", err)
			}
			// set up database info
			l1EventInfo := &L1Event{
				EventType: EventTypeAssetPausedUpdate,
				TxHash:    vlog.TxHash.Hex(),
			}
			l1Events = append(l1Events, l1EventInfo)

			err = m.processAssetPausedUpdate(event, pendingChanges)
			if err != nil {
				return err
			}
		default:
		}
	}
	// serialize into block info
	eventInfosBytes, err := json.Marshal(l1Events)
	if err != nil {
		return err
	}
	syncedBlock := &l1syncedblock.L1SyncedBlock{
		L1BlockHeight: endHeight,
		BlockInfo:     string(eventInfosBytes),
		Type:          l1syncedblock.TypeGovernance,
	}
	return m.storeChanges(syncedBlock, pendingChanges)
}

func (m *Monitor) processValidatorUpdate(event zkbnb.GovernanceValidatorStatusUpdate, pendingUpdates *GovernancePendingChanges) error {
	type ValidatorInfo struct {
		Address  string
		IsActive bool
	}
	// get data from db
	if pendingUpdates.pendingNewSysConfigMap[types.Validators] != nil {
		configInfo := pendingUpdates.pendingNewSysConfigMap[types.Validators]
		var validators map[string]*ValidatorInfo
		err := json.Unmarshal([]byte(configInfo.Value), &validators)
		if err != nil {
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
			return err
		}
		pendingUpdates.pendingNewSysConfigMap[types.Validators].Value = string(validatorBytes)
	} else {
		configInfo, err := m.SysConfigModel.GetSysConfigByName(types.Validators)
		if err != nil {
			if err != types.DbErrNotFound {
				return fmt.Errorf("unable to get sys config by name: %v", err)
			}

			validators := make(map[string]*ValidatorInfo)
			validators[event.ValidatorAddress.Hex()] = &ValidatorInfo{
				Address:  event.ValidatorAddress.Hex(),
				IsActive: event.IsActive,
			}
			validatorsBytes, err := json.Marshal(validators)
			if err != nil {
				return fmt.Errorf("unable to marshal validators: %v", err)
			}
			pendingUpdates.pendingNewSysConfigMap[types.Validators] = &sysconfig.SysConfig{
				Name:      types.Validators,
				Value:     string(validatorsBytes),
				ValueType: "map[string]*ValidatorInfo",
				Comment:   "validator info",
			}
		} else {
			var validators map[string]*ValidatorInfo
			err = json.Unmarshal([]byte(configInfo.Value), &validators)
			if err != nil {
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
				return err
			}
			if pendingUpdates.pendingUpdateSysConfigMap[types.Validators] == nil {
				pendingUpdates.pendingUpdateSysConfigMap[types.Validators] = configInfo
			}
			pendingUpdates.pendingUpdateSysConfigMap[types.Validators].Value = string(validatorBytes)
		}
	}
	return nil
}

func (m *Monitor) processAssetPausedUpdate(event zkbnb.GovernanceAssetPausedUpdate, pendingUpdates *GovernancePendingChanges) error {
	var assetInfo *asset.Asset
	if pendingUpdates.l2AssetMap[event.Token.Hex()] != nil {
		assetInfo = pendingUpdates.l2AssetMap[event.Token.Hex()]
	} else {
		assetInfo, err := m.L2AssetModel.GetAssetByAddress(event.Token.Hex())
		if err != nil {
			return fmt.Errorf("unable to get l2 asset by address, err: %v", err)
		}
		pendingUpdates.pendingUpdateL2AssetMap[event.Token.Hex()] = assetInfo
	}
	var status uint32
	if event.Paused {
		status = asset.StatusInactive
	} else {
		status = asset.StatusActive
	}
	assetInfo.Status = status
	return nil
}

func (m *Monitor) storeChanges(
	syncedBlock *l1syncedblock.L1SyncedBlock,
	pendingChanges *GovernancePendingChanges,
) (err error) {
	var (
		pendingNewAssets        []*asset.Asset
		pendingUpdateAssets     []*asset.Asset
		pendingNewSysConfigs    []*sysconfig.SysConfig
		pendingUpdateSysConfigs []*sysconfig.SysConfig
	)
	for _, l2Asset := range pendingChanges.l2AssetMap {
		pendingNewAssets = append(pendingNewAssets, l2Asset)
	}
	for _, pendingUpdateL2AssetInfo := range pendingChanges.pendingUpdateL2AssetMap {
		pendingUpdateAssets = append(pendingUpdateAssets, pendingUpdateL2AssetInfo)
	}
	for _, pendingNewSysConfigInfo := range pendingChanges.pendingNewSysConfigMap {
		pendingNewSysConfigs = append(pendingNewSysConfigs, pendingNewSysConfigInfo)
	}
	for _, pendingUpdateSysConfigInfo := range pendingChanges.pendingUpdateSysConfigMap {
		pendingUpdateSysConfigs = append(pendingUpdateSysConfigs, pendingUpdateSysConfigInfo)
	}
	if len(pendingNewAssets) > 0 || len(pendingUpdateAssets) > 0 {
		logx.Infof("l1 block info height: %v, l2 asset info size: %v, pending update l2 asset info size: %v",
			syncedBlock.L1BlockHeight,
			len(pendingNewAssets),
			len(pendingUpdateAssets),
		)
	}

	//update db
	err = m.db.Transaction(func(tx *gorm.DB) error {
		//create l1 synced block
		err := m.L1SyncedBlockModel.CreateL1SyncedBlockInTransact(tx, syncedBlock)
		if err != nil {
			return err
		}
		l1SyncedBlockHeightMetric.Set(float64(syncedBlock.L1BlockHeight))
		//create assets
		if len(pendingNewAssets) > 0 {
			err = m.L2AssetModel.CreateAssetsInTransact(tx, pendingNewAssets)
			if err != nil {
				return err
			}
		}
		//update assets
		if len(pendingUpdateAssets) > 0 {
			err = m.L2AssetModel.UpdateAssetsInTransact(tx, pendingUpdateAssets)
			if err != nil {
				return err
			}
		}
		//create sys configs
		if len(pendingNewSysConfigs) > 0 {
			err = m.SysConfigModel.CreateSysConfigsInTransact(tx, pendingNewSysConfigs)
			if err != nil {
				return err
			}
		}
		//update sys configs
		if len(pendingUpdateSysConfigs) > 0 {
			err = m.SysConfigModel.UpdateSysConfigsInTransact(tx, pendingUpdateSysConfigs)
		}
		return err
	})

	if err != nil {
		return fmt.Errorf("store governance monitor info error, err: %v", err)
	}
	return nil
}
