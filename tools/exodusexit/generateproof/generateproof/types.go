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

package generateproof

import (
	"math/big"
)

type StorageStoredBlockInfo struct {
	BlockSize                    uint16   `json:"blockSize"`
	BlockNumber                  uint32   `json:"blockNumber"`
	PriorityOperations           uint64   `json:"priorityOperations"`
	PendingOnchainOperationsHash [32]byte `json:"pendingOnchainOperationsHash"`
	Timestamp                    *big.Int `json:"timestamp"`
	StateRoot                    [32]byte `json:"stateRoot"`
	Commitment                   [32]byte `json:"commitment"`
}
type ZkBNBCommitBlockInfo struct {
	NewStateRoot      [32]byte                    `json:"newStateRoot"`
	PublicData        []byte                      `json:"publicData"`
	Timestamp         *big.Int                    `json:"timestamp"`
	OnchainOperations []ZkBNBOnchainOperationData `json:"zkBNBOnchainOperationData"`
	BlockNumber       uint32                      `json:"blockNumber"`
	BlockSize         uint16                      `json:"blockSize"`
}

type ZkBNBOnchainOperationData struct {
	EthWitness       []byte `json:"ethWitness"`
	PublicDataOffset uint32 `json:"publicDataOffset"`
}

type CommitBlocksCallData struct {
	LastCommittedBlockData *StorageStoredBlockInfo `abi:"_lastCommittedBlockData"`
	NewBlocksData          []ZkBNBCommitBlockInfo  `abi:"_newBlocksData"`
}

type ZkBNBVerifyAndExecuteBlockInfo struct {
	BlockHeader              *StorageStoredBlockInfo `abi:"blockHeader"`
	PendingOnchainOpsPubData [][]byte                `abi:"pendingOnchainOpsPubData"`
}

type VerifyAndExecuteBlocksCallData struct {
	Proofs                     []*big.Int                       `abi:"_proofs"`
	VerifyAndExecuteBlocksInfo []ZkBNBVerifyAndExecuteBlockInfo `abi:"_blocks"`
}

type PerformDesertAssetData struct {
	StoredBlockInfo    StoredBlockInfo
	AssetExitData      ExodusVerifierAssetExitData
	AccountExitData    ExodusVerifierAccountExitData
	AssetMerkleProof   [16]string
	AccountMerkleProof [32]string
	NftRoot            string
}

type PerformDesertNftData struct {
	StoredBlockInfo    StoredBlockInfo
	AccountExitData    ExodusVerifierAccountExitData
	ExitNfts           []ExodusVerifierNftExitData
	AssetRoot          string
	NftMerkleProofs    [][40]string
	AccountMerkleProof [32]string
}

type ExodusVerifierAssetExitData struct {
	AssetId                  uint32
	Amount                   int64
	OfferCanceledOrFinalized int64
}

type ExodusVerifierAccountExitData struct {
	AccountId       uint32
	L1Address       string
	PubKeyX         string
	PubKeyY         string
	Nonce           int64
	CollectionNonce int64
}

type ExodusVerifierNftExitData struct {
	NftIndex            uint64
	OwnerAccountIndex   int64
	CreatorAccountIndex int64
	NftContentHash      string
	NftContentType      uint8
	CreatorTreasuryRate int64
	CollectionId        int64
}

type StoredBlockInfo struct {
	BlockSize                    uint16
	BlockNumber                  uint32
	PriorityOperations           uint64
	PendingOnchainOperationsHash string
	Timestamp                    int64
	StateRoot                    string
	Commitment                   string
}
