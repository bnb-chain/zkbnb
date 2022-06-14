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
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	zecreyLegend "github.com/zecrey-labs/zecrey-eth-rpc/zecrey/core/zecrey-legend"
	"github.com/zecrey-labs/zecrey-legend/common/model/l1BlockMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2BlockEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/assetInfo"
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

type (
	ProviderClient = _rpc.ProviderClient
	AuthClient     = _rpc.AuthClient

	// event monitor
	L2TxEventMonitor    = l2TxEventMonitor.L2TxEventMonitor
	L2BlockEventMonitor = l2BlockEventMonitor.L2BlockEventMonitor
	L2AssetInfo         = asset.AssetInfo
	Sysconfig           = sysconfig.Sysconfig

	L2AssetInfoModel = asset.AssetInfoModel

	// model
	SysconfigModel           = sysconfig.SysconfigModel
	L1BlockMonitorModel      = l1BlockMonitor.L1BlockMonitorModel
	L2TxEventMonitorModel    = l2TxEventMonitor.L2TxEventMonitorModel
	L2BlockEventMonitorModel = l2BlockEventMonitor.L2BlockEventMonitorModel
)

const (
	// governance event name
	EventNameNewAsset              = "NewAsset"
	EventNameNewGovernor           = "NewGovernor"
	EventNameNewAssetGovernance    = "NewAssetGovernance"
	EventNameValidatorStatusUpdate = "ValidatorStatusUpdate"
	EventNameAssetPausedUpdate     = "AssetPausedUpdate"

	AddAssetEventType              = 4
	NewGovernorEventType           = 5
	NewAssetGovernanceEventType    = 6
	ValidatorStatusUpdateEventType = 7
	AssetPausedUpdateEventType     = 8
)

var (
	// err
	ErrNotFound = sqlx.ErrNotFound

	governanceLogNewAssetSig              = []byte("NewAsset(address,uint16)")
	governanceLogNewGovernorSig           = []byte("NewGovernor(address)")
	governanceLogNewAssetGovernanceSig    = []byte("NewAssetGovernance(address)")
	governanceLogValidatorStatusUpdateSig = []byte("ValidatorStatusUpdate(address,bool)")
	governanceLogAssetPausedUpdateSig     = []byte("AssetPausedUpdate(address,bool)")

	GovernanceLogNewAssetSigHash              = crypto.Keccak256Hash(governanceLogNewAssetSig)
	governanceLogNewGovernorSigHash           = crypto.Keccak256Hash(governanceLogNewGovernorSig)
	governanceLogNewAssetGovernanceSigHash    = crypto.Keccak256Hash(governanceLogNewAssetGovernanceSig)
	governanceLogValidatorStatusUpdateSigHash = crypto.Keccak256Hash(governanceLogValidatorStatusUpdateSig)
	governanceLogAssetPausedUpdateSigHash     = crypto.Keccak256Hash(governanceLogAssetPausedUpdateSig)

	GovernanceContractAbi, _ = abi.JSON(strings.NewReader(zecreyLegend.GovernanceABI))
)
