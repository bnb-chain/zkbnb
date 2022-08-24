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

package sender

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zeromicro/go-zero/core/logx"

	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/blockforcommit"
	"github.com/bnb-chain/zkbas/tree"
	"github.com/bnb-chain/zkbas/types"
)

const (
	EventNameBlockCommit       = "BlockCommit"
	EventNameBlockVerification = "BlockVerification"
)

var (
	ZkbasContractAbi, _ = abi.JSON(strings.NewReader(zkbas.ZkbasMetaData.ABI))

	zkbasLogBlockCommitSig       = []byte("BlockCommit(uint32)")
	zkbasLogBlockVerificationSig = []byte("BlockVerification(uint32)")
	zkbasLogBlocksRevertSig      = []byte("BlocksRevert(uint32,uint32)")

	zkbasLogBlockCommitSigHash       = crypto.Keccak256Hash(zkbasLogBlockCommitSig)
	zkbasLogBlockVerificationSigHash = crypto.Keccak256Hash(zkbasLogBlockVerificationSig)
	zkbasLogBlocksRevertSigHash      = crypto.Keccak256Hash(zkbasLogBlocksRevertSig)
)

func DefaultBlockHeader() zkbas.StorageStoredBlockInfo {
	var (
		pendingOnChainOperationsHash [32]byte
		stateRoot                    [32]byte
		commitment                   [32]byte
	)
	copy(pendingOnChainOperationsHash[:], common.FromHex(types.EmptyStringKeccak)[:])
	copy(stateRoot[:], tree.NilStateRoot[:])
	copy(commitment[:], common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000")[:])
	return zkbas.StorageStoredBlockInfo{
		BlockSize:                    0,
		BlockNumber:                  0,
		PriorityOperations:           0,
		PendingOnchainOperationsHash: pendingOnChainOperationsHash,
		Timestamp:                    big.NewInt(0),
		StateRoot:                    stateRoot,
		Commitment:                   commitment,
	}
}

func ConvertBlocksForCommitToCommitBlockInfos(oBlocks []*blockforcommit.BlockForCommit) (commitBlocks []zkbas.OldZkbasCommitBlockInfo, err error) {
	for _, oBlock := range oBlocks {
		var newStateRoot [32]byte
		var pubDataOffsets []uint32
		copy(newStateRoot[:], common.FromHex(oBlock.StateRoot)[:])
		err = json.Unmarshal([]byte(oBlock.PublicDataOffsets), &pubDataOffsets)
		if err != nil {
			logx.Errorf("[ConvertBlocksForCommitToCommitBlockInfos] unable to unmarshal: %s", err.Error())
			return nil, err
		}
		commitBlock := zkbas.OldZkbasCommitBlockInfo{
			NewStateRoot:      newStateRoot,
			PublicData:        common.FromHex(oBlock.PublicData),
			Timestamp:         big.NewInt(oBlock.Timestamp),
			PublicDataOffsets: pubDataOffsets,
			BlockNumber:       uint32(oBlock.BlockHeight),
			BlockSize:         oBlock.BlockSize,
		}
		commitBlocks = append(commitBlocks, commitBlock)
	}
	return commitBlocks, nil
}

func ConvertBlocksToVerifyAndExecuteBlockInfos(oBlocks []*block.Block) (verifyAndExecuteBlocks []zkbas.OldZkbasVerifyAndExecuteBlockInfo, err error) {
	for _, oBlock := range oBlocks {
		var pendingOnChainOpsPubData [][]byte
		if oBlock.PendingOnChainOperationsPubData != "" {
			err = json.Unmarshal([]byte(oBlock.PendingOnChainOperationsPubData), &pendingOnChainOpsPubData)
			if err != nil {
				logx.Errorf("[ConvertBlocksToVerifyAndExecuteBlockInfos] unable to unmarshal pending pub data: %s", err.Error())
				return nil, err
			}
		}
		verifyAndExecuteBlock := zkbas.OldZkbasVerifyAndExecuteBlockInfo{
			BlockHeader:              chain.ConstructStoredBlockInfo(oBlock),
			PendingOnchainOpsPubData: pendingOnChainOpsPubData,
		}
		verifyAndExecuteBlocks = append(verifyAndExecuteBlocks, verifyAndExecuteBlock)
	}
	return verifyAndExecuteBlocks, nil
}
