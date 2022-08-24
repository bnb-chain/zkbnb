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
 *
 */

package chain

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	curve "github.com/bnb-chain/zkbas-crypto/ecc/ztwistededwards/tebn254"
	"github.com/bnb-chain/zkbas-crypto/ffmath"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/dao/block"
)

func CreateBlockCommitment(
	currentBlockHeight int64,
	createdAt int64,
	oldStateRoot []byte,
	newStateRoot []byte,
	pubData []byte,
	onChainOpsCount int64,
) string {
	var buf bytes.Buffer
	common2.PaddingInt64IntoBuf(&buf, currentBlockHeight)
	common2.PaddingInt64IntoBuf(&buf, createdAt)
	buf.Write(CleanAndPaddingByteByModulus(oldStateRoot))
	buf.Write(CleanAndPaddingByteByModulus(newStateRoot))
	buf.Write(CleanAndPaddingByteByModulus(pubData))
	common2.PaddingInt64IntoBuf(&buf, onChainOpsCount)
	// TODO Keccak256
	//hFunc := mimc.NewMiMC()
	//hFunc.Write(buf.Bytes())
	//commitment := hFunc.Sum(nil)
	commitment := common2.KeccakHash(buf.Bytes())
	return common.Bytes2Hex(commitment)
}

func ConstructStoredBlockInfo(oBlock *block.Block) zkbas.StorageStoredBlockInfo {
	var (
		PendingOnchainOperationsHash [32]byte
		StateRoot                    [32]byte
		Commitment                   [32]byte
	)
	copy(PendingOnchainOperationsHash[:], common.FromHex(oBlock.PendingOnChainOperationsHash)[:])
	copy(StateRoot[:], common.FromHex(oBlock.StateRoot)[:])
	copy(Commitment[:], common.FromHex(oBlock.BlockCommitment)[:])
	return zkbas.StorageStoredBlockInfo{
		BlockNumber:                  uint32(oBlock.BlockHeight),
		PriorityOperations:           uint64(oBlock.PriorityOperations),
		PendingOnchainOperationsHash: PendingOnchainOperationsHash,
		Timestamp:                    big.NewInt(oBlock.CreatedAt.UnixMilli()),
		StateRoot:                    StateRoot,
		Commitment:                   Commitment,
		BlockSize:                    oBlock.BlockSize,
	}
}

func CleanAndPaddingByteByModulus(buf []byte) []byte {
	if len(buf) <= 32 {
		return ffmath.Mod(new(big.Int).SetBytes(buf), curve.Modulus).FillBytes(make([]byte, 32))
	}
	offset := 32
	var pendingBuf bytes.Buffer
	for offset <= len(buf) {
		pendingBuf.Write(ffmath.Mod(new(big.Int).SetBytes(buf[offset-32:offset]), curve.Modulus).FillBytes(make([]byte, 32)))
		offset += 32
	}
	return pendingBuf.Bytes()
}
