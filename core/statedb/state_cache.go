package statedb

import (
	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/dao/block"
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
	PendingAccountMap          map[int64]*types.AccountInfo
	PendingAccountL1AddressMap map[string]int64
	PendingNftMap              map[int64]*nft.L2Nft
	PendingGasMap              map[int64]*big.Int //pending gas changes of a block

	// Record the tree states that should be updated.
	dirtyAccountsAndAssetsMap map[int64]map[int64]bool
	dirtyNftMap               map[int64]bool
}

type StateDataCopy struct {
	StateCache            *StateCache
	CurrentBlock          *block.Block
	pendingAccountSmtItem []bsmt.Item
	pendingNftSmtItem     []bsmt.Item
}

func NewStateCache(stateRoot string) *StateCache {
	return &StateCache{
		StateRoot: stateRoot,
		Txs:       make([]*tx.Tx, 0),

		PendingAccountMap:          make(map[int64]*types.AccountInfo, 0),
		PendingAccountL1AddressMap: make(map[string]int64, 0),
		PendingNftMap:              make(map[int64]*nft.L2Nft, 0),
		PendingGasMap:              make(map[int64]*big.Int, 0),

		PubData:                         make([]byte, 0),
		PriorityOperations:              0,
		PubDataOffset:                   make([]uint32, 0),
		PendingOnChainOperationsPubData: make([][]byte, 0),
		PendingOnChainOperationsHash:    common.FromHex(types.EmptyStringKeccak),

		dirtyAccountsAndAssetsMap: make(map[int64]map[int64]bool, 0),
		dirtyNftMap:               make(map[int64]bool, 0),
	}
}

func (c *StateCache) AlignPubData(blockSize int, stateCopy *StateDataCopy) {
	emptyPubData := make([]byte, (blockSize-len(stateCopy.StateCache.Txs))*cryptoTypes.PubDataBitsSizePerTx/8)
	stateCopy.StateCache.PubData = append(stateCopy.StateCache.PubData, emptyPubData...)
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

func (c *StateCache) GetDirtyAccountsAndAssetsMap() map[int64]map[int64]bool {
	return c.dirtyAccountsAndAssetsMap
}

func (c *StateCache) GetDirtyNftMap() map[int64]bool {
	return c.dirtyNftMap
}

func (c *StateCache) GetPendingAccount(accountIndex int64) (*types.AccountInfo, bool) {
	account, exist := c.PendingAccountMap[accountIndex]
	if exist {
		return account, exist
	}
	return nil, false
}

func (c *StateCache) GetPendingAccountL1AddressMap(l1Address string) (int64, bool) {
	accountIndex, exist := c.PendingAccountL1AddressMap[l1Address]
	if exist {
		return accountIndex, exist
	}
	return -1, false
}

func (c *StateCache) GetPendingNft(nftIndex int64) (*nft.L2Nft, bool) {
	nft, exist := c.PendingNftMap[nftIndex]
	if exist {
		return nft, exist
	}
	return nil, false
}

func (c *StateCache) SetPendingAccount(accountIndex int64, account *types.AccountInfo) {
	c.PendingAccountMap[accountIndex] = account
}
func (c *StateCache) SetPendingAccountL1AddressMap(l1Address string, accountIndex int64) {
	c.PendingAccountL1AddressMap[l1Address] = accountIndex
}

func (c *StateCache) SetPendingNft(nftIndex int64, nft *nft.L2Nft) {
	c.PendingNftMap[nftIndex] = nft
}

func (c *StateCache) GetPendingGas(assetId int64) *big.Int {
	if delta, ok := c.PendingGasMap[assetId]; ok {
		return delta
	}
	return types.ZeroBigInt
}

func (c *StateCache) SetPendingGas(assetId int64, balanceDelta *big.Int) {
	if _, ok := c.PendingGasMap[assetId]; !ok {
		c.PendingGasMap[assetId] = types.ZeroBigInt
	}
	c.PendingGasMap[assetId] = ffmath.Add(c.PendingGasMap[assetId], balanceDelta)
}
