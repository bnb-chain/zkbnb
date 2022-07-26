package util

import (
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

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
