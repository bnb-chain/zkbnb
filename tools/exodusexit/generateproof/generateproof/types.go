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
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
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

type ZkBNBVerifyAndExecuteBlockInfo struct {
	BlockHeader              *StorageStoredBlockInfo `abi:"blockHeader"`
	PendingOnchainOpsPubData [][]byte                `abi:"pendingOnchainOpsPubData"`
}

type VerifyAndExecuteBlocksCallData struct {
	Proofs                     []*big.Int                       `abi:"_proofs"`
	VerifyAndExecuteBlocksInfo []ZkBNBVerifyAndExecuteBlockInfo `abi:"_blocks"`
}

type PerformDesertData struct {
	NftRoot            [32]byte
	ExitData           zkbnb.ExodusVerifierExitData
	AssetMerkleProof   [16][32]byte
	AccountMerkleProof [32][32]byte
}

type PerformDesertNftData struct {
	OwnerAccountIndex *big.Int
	AccountRoot       [32]byte
	ExitNfts          []zkbnb.ExodusVerifierExitNftData
	NftMerkleProofs   [][40][32]byte
}
