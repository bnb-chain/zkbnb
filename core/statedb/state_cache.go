package statedb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/zkbas/dao/nft"
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

	PendingNewNftWithdrawHistory []*nft.L2NftWithdrawHistory
	PendingNewOffer              []*nft.Offer
	PendingNewL2NftExchange      []*nft.L2NftExchange
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

		PendingNewNftWithdrawHistory: make([]*nft.L2NftWithdrawHistory, 0),
		PendingNewOffer:              make([]*nft.Offer, 0),
		PendingNewL2NftExchange:      make([]*nft.L2NftExchange, 0),

		PubData:                         make([]byte, 0),
		PriorityOperations:              0,
		PubDataOffset:                   make([]uint32, 0),
		PendingOnChainOperationsPubData: make([][]byte, 0),
		PendingOnChainOperationsHash:    common.FromHex(types.EmptyStringKeccak),
	}
}