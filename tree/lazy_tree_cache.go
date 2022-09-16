package tree

import (
	bsmt "github.com/bnb-chain/zkbnb-smt"
	lru "github.com/hashicorp/golang-lru"
)

type LazyTreeCache struct {
	funcMap   map[int64]func() bsmt.SparseMerkleTree
	treeCache *lru.Cache
}

func NewLazyTreeCache(maxSize int) *LazyTreeCache {
	l, _ := lru.New(maxSize)
	m := make(map[int64]func() bsmt.SparseMerkleTree)
	return &LazyTreeCache{funcMap: m, treeCache: l}
}

func (c *LazyTreeCache) AddToIndex(i int64, f func() bsmt.SparseMerkleTree) {
	c.funcMap[i] = f
}

func (c *LazyTreeCache) Add(f func() bsmt.SparseMerkleTree) {
	c.funcMap[int64(len(c.funcMap))] = f
}

func (c *LazyTreeCache) Get(i int64) bsmt.SparseMerkleTree {
	c.treeCache.ContainsOrAdd(i, c.funcMap[i]())
	tree, _ := c.treeCache.Get(i)
	return tree.(bsmt.SparseMerkleTree)
}

func (c *LazyTreeCache) Size() int64 {
	return int64(len(c.funcMap))
}
