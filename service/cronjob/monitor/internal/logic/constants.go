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
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	zecreyLegend "github.com/zecrey-labs/zecrey-eth-rpc/zecrey/core/zecrey-legend"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	asset "github.com/zecrey-labs/zecrey-legend/common/model/assetInfo"
	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/model/l1BlockMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l1TxSender"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2BlockEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
)

type (
	ProviderClient = _rpc.ProviderClient
	AuthClient     = _rpc.AuthClient

	L1BlockMonitorModel      = l1BlockMonitor.L1BlockMonitorModel
	L2TxEventMonitorModel    = l2TxEventMonitor.L2TxEventMonitorModel
	L2BlockEventMonitorModel = l2BlockEventMonitor.L2BlockEventMonitorModel
	SysconfigModel           = sysconfig.SysconfigModel
	MempoolModel             = mempool.MempoolModel
	BlockModel               = block.BlockModel
	L2AssetInfoModel         = asset.AssetInfoModel
	L1TxSenderModel          = l1TxSender.L1TxSenderModel

	L2AssetInfo         = asset.AssetInfo
	Sysconfig           = sysconfig.Sysconfig
	L2TxEventMonitor    = l2TxEventMonitor.L2TxEventMonitor
	L2BlockEventMonitor = l2BlockEventMonitor.L2BlockEventMonitor
	Block               = block.Block
	L1TxSender          = l1TxSender.L1TxSender
	MempoolTx           = mempool.MempoolTx

	ZecreyLegendBlockCommit       = zecreyLegend.ZecreyLegendBlockCommit
	ZecreyLegendBlockVerification = zecreyLegend.ZecreyLegendBlockVerification
)

const (
	// zecrey event name
	EventNameNewPriorityRequest = "NewPriorityRequest"
	EventNameBlockCommit        = "BlockCommit"
	EventNameBlockVerification  = "BlockVerification"
	EventNameBlocksRevert       = "BlocksRevert"

	// tx type for l2 block event monitors
	EventTypeNewPriorityRequest = 0
	EventTypeCommittedBlock     = l2BlockEventMonitor.CommittedBlockEventType
	EventTypeVerifiedBlock      = l2BlockEventMonitor.VerifiedBlockEventType
	EventTypeRevertedBlock      = l2BlockEventMonitor.RevertedBlockEventType
	// status
	PendingStatusL2BlockEventMonitor = l2BlockEventMonitor.PendingStatus

	// governance event name
	EventNameNewAsset              = "NewAsset"
	EventNameNewGovernor           = "NewGovernor"
	EventNameNewAssetGovernance    = "NewAssetGovernance"
	EventNameValidatorStatusUpdate = "ValidatorStatusUpdate"
	EventNameAssetPausedUpdate     = "AssetPausedUpdate"

	EventTypeAddAsset              = 4
	EventTypeNewGovernor           = 5
	EventTypeNewAssetGovernance    = 6
	EventTypeValidatorStatusUpdate = 7
	EventTypeAssetPausedUpdate     = 8

	// event status
	PendingStatus = l2TxEventMonitor.PendingStatus

	// tx type
	TxTypeRegisterZns    = commonTx.TxTypeRegisterZns
	TxTypeCreatePair     = commonTx.TxTypeCreatePair
	TxTypeUpdatePairRate = commonTx.TxTypeUpdatePairRate
	TxTypeDeposit        = commonTx.TxTypeDeposit
	TxTypeDepositNft     = commonTx.TxTypeDepositNft
	TxTypeFullExit       = commonTx.TxTypeFullExit
	TxTypeFullExitNft    = commonTx.TxTypeFullExitNft

	GeneralAssetType = commonAsset.GeneralAssetType

	BlockVerifiedStatus = block.StatusVerifiedAndExecuted

	L1TxSenderPendingStatus = l1TxSender.PendingStatus
	L1TxSenderHandledStatus = l1TxSender.HandledStatus
)

var (
	// err
	ErrNotFound = sqlx.ErrNotFound

	ZecreyContractAbi, _ = abi.JSON(strings.NewReader(zecreyLegend.ZecreyLegendABI))
	// Zecrey contract logs sig
	zecreyLogNewPriorityRequestSig = []byte("NewPriorityRequest(address,uint64,uint8,bytes,uint256)")
	zecreyLogWithdrawalSig         = []byte("Withdrawal(uint16,uint128)")
	zecreyLogWithdrawalPendingSig  = []byte("WithdrawalPending(uint16,uint128)")
	zecreyLogBlockCommitSig        = []byte("BlockCommit(uint32)")
	zecreyLogBlockVerificationSig  = []byte("BlockVerification(uint32)")
	zecreyLogBlocksRevertSig       = []byte("BlocksRevert(uint32,uint32)")

	zecreyLogNewPriorityRequestSigHash = crypto.Keccak256Hash(zecreyLogNewPriorityRequestSig)
	zecreyLogWithdrawalSigHash         = crypto.Keccak256Hash(zecreyLogWithdrawalSig)
	zecreyLogWithdrawalPendingSigHash  = crypto.Keccak256Hash(zecreyLogWithdrawalPendingSig)
	zecreyLogBlockCommitSigHash        = crypto.Keccak256Hash(zecreyLogBlockCommitSig)
	zecreyLogBlockVerificationSigHash  = crypto.Keccak256Hash(zecreyLogBlockVerificationSig)
	zecreyLogBlocksRevertSigHash       = crypto.Keccak256Hash(zecreyLogBlocksRevertSig)

	GovernanceContractAbi, _ = abi.JSON(strings.NewReader(zecreyLegend.GovernanceABI))

	governanceLogNewAssetSig              = []byte("NewAsset(address,uint16)")
	governanceLogNewGovernorSig           = []byte("NewGovernor(address)")
	governanceLogNewAssetGovernanceSig    = []byte("NewAssetGovernance(address)")
	governanceLogValidatorStatusUpdateSig = []byte("ValidatorStatusUpdate(address,bool)")
	governanceLogAssetPausedUpdateSig     = []byte("AssetPausedUpdate(address,bool)")

	governanceLogNewAssetSigHash              = crypto.Keccak256Hash(governanceLogNewAssetSig)
	governanceLogNewGovernorSigHash           = crypto.Keccak256Hash(governanceLogNewGovernorSig)
	governanceLogNewAssetGovernanceSigHash    = crypto.Keccak256Hash(governanceLogNewAssetGovernanceSig)
	governanceLogValidatorStatusUpdateSigHash = crypto.Keccak256Hash(governanceLogValidatorStatusUpdateSig)
	governanceLogAssetPausedUpdateSigHash     = crypto.Keccak256Hash(governanceLogAssetPausedUpdateSig)
)

const (
	ZeroBigIntString = "0"
)
