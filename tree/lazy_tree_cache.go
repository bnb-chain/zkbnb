package tree

import (
	"sync"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	lru "github.com/hashicorp/golang-lru"
)

type AssetTreeCache struct {
	funcMap    map[int64]func() bsmt.SparseMerkleTree
	funcLock   sync.RWMutex
	commitMap  map[int64]bool
	CommitLock sync.RWMutex
	treeCache  *lru.Cache
}

func NewLazyTreeCache(maxSize int) *AssetTreeCache {
	cache := AssetTreeCache{funcMap: make(map[int64]func() bsmt.SparseMerkleTree), commitMap: make(map[int64]bool)}
	cache.treeCache, _ = lru.NewWithEvict(maxSize, cache.OnDelete)
	return &cache
}

func (c *AssetTreeCache) AddToIndex(i int64, f func() bsmt.SparseMerkleTree) {
	c.funcLock.Lock()
	c.funcMap[i] = f
	c.funcLock.Unlock()
}

func (c *AssetTreeCache) Add(f func() bsmt.SparseMerkleTree) {
	c.funcLock.Lock()
	c.funcMap[int64(len(c.funcMap))] = f
	c.funcLock.Unlock()
}

func (c *AssetTreeCache) Get(i int64) (tree bsmt.SparseMerkleTree) {
	c.funcLock.RLock()
	c.treeCache.ContainsOrAdd(i, c.funcMap[i]())
	c.funcLock.RUnlock()
	if tmpTree, ok := c.treeCache.Get(i); ok {
		tree = tmpTree.(bsmt.SparseMerkleTree)
	}
	c.CommitLock.Lock()
	c.commitMap[i] = (tree.LatestVersion()-tree.RecentVersion() > 1)
	c.CommitLock.Unlock()
	return
}

func (c *AssetTreeCache) NeedsCommit(i int64) (ret bool) {
	if c.treeCache.Contains(i) {
		if tree, ok := c.treeCache.Peek(i); ok {
			c.CommitLock.Lock()
			c.commitMap[i] = (tree.(bsmt.SparseMerkleTree).LatestVersion()-tree.(bsmt.SparseMerkleTree).RecentVersion() > 1)
			ret = c.commitMap[i]
			c.CommitLock.Unlock()
			return
		}
	}
	c.CommitLock.RLock()
	ret = c.commitMap[i]
	c.CommitLock.RUnlock()
	return
}

func (c *AssetTreeCache) OnDelete(k, v interface{}) {
	c.CommitLock.Lock()
	c.commitMap[k.(int64)] = (v.(bsmt.SparseMerkleTree).LatestVersion()-v.(bsmt.SparseMerkleTree).RecentVersion() > 1)
	c.CommitLock.Unlock()
}

func (c *AssetTreeCache) Size() int64 {
	c.funcLock.RLock()
	defer c.funcLock.RUnlock()
	return int64(len(c.funcMap))
}
