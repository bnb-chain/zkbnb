package util

import (
	"math/big"

	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/zkbas/common/model/block"
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
