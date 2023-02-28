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

package performexodus

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"

	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb/dao/priorityrequest"
	"github.com/bnb-chain/zkbnb/types"
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

	PendingStatus = priorityrequest.PendingStatus

	TxTypeRegisterZns = types.TxTypeRegisterZns
	TxTypeDeposit     = types.TxTypeDeposit
	TxTypeDepositNft  = types.TxTypeDepositNft
	TxTypeFullExit    = types.TxTypeFullExit
	TxTypeFullExitNft = types.TxTypeFullExitNft
)

var (
	ZkBNBContractAbi, _ = abi.JSON(strings.NewReader(zkbnb.ZkBNBMetaData.ABI))
	// ZkBNB contract logs sig
	zkbnbLogNewPriorityRequestSig = []byte("NewPriorityRequest(address,uint64,uint8,bytes,uint256)")
	zkbnbLogWithdrawalSig         = []byte("Withdrawal(uint16,uint128)")
	zkbnbLogWithdrawalPendingSig  = []byte("WithdrawalPending(uint16,uint128)")
	zkbnbLogBlockCommitSig        = []byte("BlockCommit(uint32)")
	zkbnbLogBlockVerificationSig  = []byte("BlockVerification(uint32)")
	zkbnbLogBlocksRevertSig       = []byte("BlocksRevert(uint32,uint32)")

	zkbnbLogNewPriorityRequestSigHash = crypto.Keccak256Hash(zkbnbLogNewPriorityRequestSig)
	zkbnbLogWithdrawalSigHash         = crypto.Keccak256Hash(zkbnbLogWithdrawalSig)
	zkbnbLogWithdrawalPendingSigHash  = crypto.Keccak256Hash(zkbnbLogWithdrawalPendingSig)
	zkbnbLogBlockCommitSigHash        = crypto.Keccak256Hash(zkbnbLogBlockCommitSig)
	zkbnbLogBlockVerificationSigHash  = crypto.Keccak256Hash(zkbnbLogBlockVerificationSig)
	zkbnbLogBlocksRevertSigHash       = crypto.Keccak256Hash(zkbnbLogBlocksRevertSig)

	GovernanceContractAbi, _ = abi.JSON(strings.NewReader(zkbnb.GovernanceMetaData.ABI))

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

type L1Event struct {
	// deposit / lock / committed / verified / reverted
	EventType uint8
	// tx hash
	TxHash string
	// index of the log in the block
	Index uint
}

type StorageStoredBlockInfo struct {
	BlockSize                    uint16   `json:"blockSize"`
	BlockNumber                  uint32   `json:"blockNumber"`
	PriorityOperations           uint64   `json:"priorityOperations"`
	PendingOnchainOperationsHash [32]byte `json:"pendingOnchainOperationsHash"`
	Timestamp                    *big.Int `json:"timestamp"`
	StateRoot                    [32]byte `json:"stateRoot"`
	Commitment                   [32]byte `json:"commitment"`
}
type OldZkBNBCommitBlockInfo struct {
	NewStateRoot      [32]byte `json:"newStateRoot"`
	PublicData        []byte   `json:"publicData"`
	Timestamp         *big.Int `json:"timestamp"`
	PublicDataOffsets []uint32 `json:"publicDataOffsets"`
	BlockNumber       uint32   `json:"blockNumber"`
	BlockSize         uint16   `json:"blockSize"`
}

type CommitBlocksCallData struct {
	LastCommittedBlockData *StorageStoredBlockInfo   `abi:"_lastCommittedBlockData"`
	NewBlocksData          []OldZkBNBCommitBlockInfo `abi:"_newBlocksData"`
}

type OldZkBNBVerifyAndExecuteBlockInfo struct {
	BlockHeader              *StorageStoredBlockInfo `abi:"blockHeader"`
	PendingOnchainOpsPubData [][]byte                `abi:"pendingOnchainOpsPubData"`
}

type VerifyAndExecuteBlocksCallData struct {
	Proofs                     []*big.Int                          `abi:"_proofs"`
	VerifyAndExecuteBlocksInfo []OldZkBNBVerifyAndExecuteBlockInfo `abi:"_blocks"`
}
