/*
 * Copyright © 2021 ZkBAS Protocol
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
	"strings"

	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/priorityRequest"
)

const (
	EventNameNewPriorityRequest = "NewPriorityRequest"
	EventNameBlockCommit        = "BlockCommit"
	EventNameBlockVerification  = "BlockVerification"

	EventTypeNewPriorityRequest = 0
	EventTypeCommittedBlock     = 1
	EventTypeVerifiedBlock      = 2
	EventTypeRevertedBlock      = 3

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

	PendingStatus = priorityRequest.PendingStatus

	TxTypeRegisterZns    = commonTx.TxTypeRegisterZns
	TxTypeCreatePair     = commonTx.TxTypeCreatePair
	TxTypeUpdatePairRate = commonTx.TxTypeUpdatePairRate
	TxTypeDeposit        = commonTx.TxTypeDeposit
	TxTypeDepositNft     = commonTx.TxTypeDepositNft
	TxTypeFullExit       = commonTx.TxTypeFullExit
	TxTypeFullExitNft    = commonTx.TxTypeFullExitNft

	BlockVerifiedStatus = block.StatusVerifiedAndExecuted
)

var (
	ZkbasContractAbi, _ = abi.JSON(strings.NewReader(zkbas.ZkbasMetaData.ABI))
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

	GovernanceContractAbi, _ = abi.JSON(strings.NewReader(zkbas.GovernanceMetaData.ABI))

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

type L1EventInfo struct {
	// deposit / lock / committed / verified / reverted
	EventType uint8
	// tx hash
	TxHash string
}
