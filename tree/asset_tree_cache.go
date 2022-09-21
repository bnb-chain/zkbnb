package tree

import (
	"sync"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	lru "github.com/hashicorp/golang-lru"
)

// Lazy init cache for asset trees
type AssetTreeCache struct {
	initFunction      func(index int64) bsmt.SparseMerkleTree
	nextAccountNumber int64
	mainLock          sync.RWMutex
	changes           map[int64]bool
	changesLock       sync.RWMutex
	treeCache         *lru.Cache
}

// Creates new AssetTreeCache
// maxSize defines the maximum size of currently initialized trees
// accountNumber defines the number of accounts to create/or next index for new account
func NewLazyTreeCache(maxSize int, accountNumber int64, f func(index int64) bsmt.SparseMerkleTree) *AssetTreeCache {
	cache := AssetTreeCache{initFunction: f, nextAccountNumber: accountNumber, changes: make(map[int64]bool, maxSize*10)}
	cache.treeCache, _ = lru.NewWithEvict(maxSize, cache.onDelete)
	return &cache
}

// Updates current cache with new(updated) init function and with latest account index
func (c *AssetTreeCache) UpdateCache(accountNumber int64, f func(index int64) bsmt.SparseMerkleTree) {
	c.mainLock.Lock()
	c.initFunction = f
	if c.nextAccountNumber < accountNumber {
		c.nextAccountNumber = accountNumber
	}
	c.mainLock.Unlock()
}

// Returns index of next account
func (c *AssetTreeCache) GetNextAccountIndex() int64 {
	c.mainLock.RLock()
	defer c.mainLock.RUnlock()
	return c.nextAccountNumber + 1
}

// Returns asset tree based on account index
func (c *AssetTreeCache) Get(i int64) (tree bsmt.SparseMerkleTree) {
	c.mainLock.RLock()
	c.treeCache.ContainsOrAdd(i, c.initFunction(i))
	c.mainLock.RUnlock()
	if tmpTree, ok := c.treeCache.Get(i); ok {
		tree = tmpTree.(bsmt.SparseMerkleTree)
	}
	return
}

// Returns slice of indexes of asset trees that were changned
// and resets cache of the marked changes of asset trees.
func (c *AssetTreeCache) GetChanges() []int64 {
	c.mainLock.Lock()
	c.changesLock.Lock()
	defer c.mainLock.Unlock()
	defer c.changesLock.Unlock()
	for _, key := range c.treeCache.Keys() {
		tree, _ := c.treeCache.Peek(key)
		if tree.(bsmt.SparseMerkleTree).LatestVersion()-tree.(bsmt.SparseMerkleTree).RecentVersion() > 1 {
			c.changes[key.(int64)] = true
		}
	}
	ret := make([]int64, 0, len(c.changes))
	for key := range c.changes {
		ret = append(ret, key)
	}
	// Cleans map
	c.changes = make(map[int64]bool, len(c.changes))
	return ret
}

// Internal method to that marks if changes happend to tree eviced from LRU
func (c *AssetTreeCache) onDelete(k, v interface{}) {
	c.changesLock.Lock()
	if v.(bsmt.SparseMerkleTree).LatestVersion()-v.(bsmt.SparseMerkleTree).RecentVersion() > 1 {
		c.changes[k.(int64)] = true
	}
	c.changesLock.Unlock()
}
