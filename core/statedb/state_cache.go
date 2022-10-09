package statedb

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	cryptoTypes "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
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

	// Record the flat data that should be updated.
	PendingNewAccountMap    map[int64]*types.AccountInfo
	PendingNewNftMap        map[int64]*nft.L2Nft
	PendingUpdateAccountMap map[int64]*types.AccountInfo
	PendingUpdateNftMap     map[int64]*nft.L2Nft
	PendingGasMap           map[int64]*big.Int //pending gas changes of a block

	// Record the tree states that should be updated.
	dirtyAccountsAndAssetsMap map[int64]map[int64]bool
	dirtyNftMap               map[int64]bool
}

func NewStateCache(stateRoot string) *StateCache {
	return &StateCache{
		StateRoot: stateRoot,
		Txs:       make([]*tx.Tx, 0),

		PendingNewAccountMap:    make(map[int64]*types.AccountInfo, 0),
		PendingNewNftMap:        make(map[int64]*nft.L2Nft, 0),
		PendingUpdateAccountMap: make(map[int64]*types.AccountInfo, 0),
		PendingUpdateNftMap:     make(map[int64]*nft.L2Nft, 0),
		PendingGasMap:           make(map[int64]*big.Int, 0),

		PubData:                         make([]byte, 0),
		PriorityOperations:              0,
		PubDataOffset:                   make([]uint32, 0),
		PendingOnChainOperationsPubData: make([][]byte, 0),
		PendingOnChainOperationsHash:    common.FromHex(types.EmptyStringKeccak),

		dirtyAccountsAndAssetsMap: make(map[int64]map[int64]bool, 0),
		dirtyNftMap:               make(map[int64]bool, 0),
	}
}

func (c *StateCache) AlignPubData(blockSize int) {
	emptyPubdata := make([]byte, (blockSize-len(c.Txs))*32*cryptoTypes.PubDataSizePerTx)
	c.PubData = append(c.PubData, emptyPubdata...)
}

func (c *StateCache) MarkAccountAssetsDirty(accountIndex int64, assets []int64) {
	if accountIndex < 0 {
		return
	}

	if _, ok := c.dirtyAccountsAndAssetsMap[accountIndex]; !ok {
		c.dirtyAccountsAndAssetsMap[accountIndex] = make(map[int64]bool, 0)
	}

	for _, assetIndex := range assets {
		// Should never happen, but protect here.
		if assetIndex < 0 {
			continue
		}
		c.dirtyAccountsAndAssetsMap[accountIndex][assetIndex] = true
	}
}

func (c *StateCache) MarkNftDirty(nftIndex int64) {
	c.dirtyNftMap[nftIndex] = true
}

func (c *StateCache) GetPendingAccount(accountIndex int64) (*types.AccountInfo, bool) {
	account, exist := c.PendingNewAccountMap[accountIndex]
	if exist {
		return account, exist
	}
	account, exist = c.PendingUpdateAccountMap[accountIndex]
	if exist {
		return account, exist
	}
	return nil, false
}

func (c *StateCache) GetPendingNft(nftIndex int64) (*nft.L2Nft, bool) {
	nft, exist := c.PendingNewNftMap[nftIndex]
	if exist {
		return nft, exist
	}
	nft, exist = c.PendingUpdateNftMap[nftIndex]
	if exist {
		return nft, exist
	}
	return nil, false
}

func (c *StateCache) SetPendingNewAccount(accountIndex int64, account *types.AccountInfo) {
	c.PendingNewAccountMap[accountIndex] = account
}

func (c *StateCache) SetPendingUpdateAccount(accountIndex int64, account *types.AccountInfo) {
	c.PendingUpdateAccountMap[accountIndex] = account
}

func (c *StateCache) SetPendingNewNft(nftIndex int64, nft *nft.L2Nft) {
	c.PendingNewNftMap[nftIndex] = nft
}

func (c *StateCache) SetPendingUpdateNft(nftIndex int64, nft *nft.L2Nft) {
	c.PendingUpdateNftMap[nftIndex] = nft
}

func (c *StateCache) GetPendingUpdateGas(assetId int64) *big.Int {
	if delta, ok := c.PendingGasMap[assetId]; ok {
		return delta
	}
	return types.ZeroBigInt
}

func (c *StateCache) SetPendingUpdateGas(assetId int64, balanceDelta *big.Int) {
	if _, ok := c.PendingGasMap[assetId]; !ok {
		c.PendingGasMap[assetId] = types.ZeroBigInt
	}
	c.PendingGasMap[assetId] = ffmath.Add(c.PendingGasMap[assetId], balanceDelta)
}
