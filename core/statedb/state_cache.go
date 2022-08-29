package statedb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/types"
)

const (
	_ = iota
	StateCachePending
	StateCacheCached
)

type StateCache struct {
	StateRoot string
	// Updated in executor's GeneratePubData method.
	PubData                         []byte
	PriorityOperations              int64
	PubDataOffset                   []uint32
	PendingOnChainOperationsPubData [][]byte
	PendingOnChainOperationsHash    []byte
	Txs                             []*tx.Tx

	// Updated in executor's ApplyTransaction method.
	PendingNewAccountIndexMap      map[int64]int
	PendingNewLiquidityIndexMap    map[int64]int
	PendingNewNftIndexMap          map[int64]int
	PendingUpdateAccountIndexMap   map[int64]int
	PendingUpdateLiquidityIndexMap map[int64]int
	PendingUpdateNftIndexMap       map[int64]int
}

func NewStateCache(stateRoot string) *StateCache {
	return &StateCache{
		StateRoot: stateRoot,
		Txs:       make([]*tx.Tx, 0),

		PendingNewAccountIndexMap:      make(map[int64]int, 0),
		PendingNewLiquidityIndexMap:    make(map[int64]int, 0),
		PendingNewNftIndexMap:          make(map[int64]int, 0),
		PendingUpdateAccountIndexMap:   make(map[int64]int, 0),
		PendingUpdateLiquidityIndexMap: make(map[int64]int, 0),
		PendingUpdateNftIndexMap:       make(map[int64]int, 0),

		PubData:                         make([]byte, 0),
		PriorityOperations:              0,
		PubDataOffset:                   make([]uint32, 0),
		PendingOnChainOperationsPubData: make([][]byte, 0),
		PendingOnChainOperationsHash:    common.FromHex(types.EmptyStringKeccak),
	}
}

func (c *StateCache) AlignPubData(blockSize int) {
	emptyPubdata := make([]byte, (blockSize-len(c.Txs))*32*std.PubDataSizePerTx)
	c.PubData = append(c.PubData, emptyPubdata...)
}
