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
	"strings"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonTx"
	asset "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/l1BlockMonitor"
	"github.com/bnb-chain/zkbas/common/model/l1TxSender"
	"github.com/bnb-chain/zkbas/common/model/l2BlockEventMonitor"
	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
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

	ZkbasBlockCommit       = zkbas.ZkbasBlockCommit
	ZkbasBlockVerification = zkbas.ZkbasBlockVerification
)

const (
	// zkbas event name
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

	ZkbasContractAbi, _ = abi.JSON(strings.NewReader(zkbas.ZkbasABI))
	// Zkbas contract logs sig
	zkbasLogNewPriorityRequestSig = []byte("NewPriorityRequest(address,uint64,uint8,bytes,uint256)")
	zkbasLogWithdrawalSig         = []byte("Withdrawal(uint16,uint128)")
	zkbasLogWithdrawalPendingSig  = []byte("WithdrawalPending(uint16,uint128)")
	zkbasLogBlockCommitSig        = []byte("BlockCommit(uint32)")
	zkbasLogBlockVerificationSig  = []byte("BlockVerification(uint32)")
	zkbasLogBlocksRevertSig       = []byte("BlocksRevert(uint32,uint32)")

	zkbasLogNewPriorityRequestSigHash = crypto.Keccak256Hash(zkbasLogNewPriorityRequestSig)
	zkbasLogWithdrawalSigHash         = crypto.Keccak256Hash(zkbasLogWithdrawalSig)
	zkbasLogWithdrawalPendingSigHash  = crypto.Keccak256Hash(zkbasLogWithdrawalPendingSig)
	zkbasLogBlockCommitSigHash        = crypto.Keccak256Hash(zkbasLogBlockCommitSig)
	zkbasLogBlockVerificationSigHash  = crypto.Keccak256Hash(zkbasLogBlockVerificationSig)
	zkbasLogBlocksRevertSigHash       = crypto.Keccak256Hash(zkbasLogBlocksRevertSig)

	GovernanceContractAbi, _ = abi.JSON(strings.NewReader(zkbas.GovernanceABI))

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
