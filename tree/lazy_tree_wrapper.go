package tree

import (
	"sync"

	bsmt "github.com/bnb-chain/zkbnb-smt"
)

type LazyTreeWrapper struct {
	doOnce   sync.Once
	tree     bsmt.SparseMerkleTree
	initFunc func() bsmt.SparseMerkleTree
}

func NewLazyTreeWrapper(f func() bsmt.SparseMerkleTree) LazyTreeWrapper {
	return LazyTreeWrapper{initFunc: f}
}

func (tw *LazyTreeWrapper) lazyInit() {
	tw.doOnce.Do(func() {
		tw.tree = tw.initFunc()
	})
}

// SparseMerkleTree interface functions

func (tw *LazyTreeWrapper) Size() uint64 {
	tw.lazyInit()
	return tw.tree.Size()
}

func (tw *LazyTreeWrapper) Get(key uint64, version *bsmt.Version) ([]byte, error) {
	tw.lazyInit()
	return tw.tree.Get(key, version)
}

func (tw *LazyTreeWrapper) Set(key uint64, val []byte) error {
	tw.lazyInit()
	return tw.tree.Set(key, val)
}

func (tw *LazyTreeWrapper) IsEmpty() bool {
	tw.lazyInit()
	return tw.tree.IsEmpty()
}

func (tw *LazyTreeWrapper) Root() []byte {
	tw.lazyInit()
	return tw.tree.Root()
}

func (tw *LazyTreeWrapper) GetProof(key uint64) (bsmt.Proof, error) {
	tw.lazyInit()
	return tw.tree.GetProof(key)
}

func (tw *LazyTreeWrapper) VerifyProof(key uint64, proof bsmt.Proof) bool {
	tw.lazyInit()
	return tw.tree.VerifyProof(key, proof)
}

func (tw *LazyTreeWrapper) LatestVersion() bsmt.Version {
	tw.lazyInit()
	return tw.tree.LatestVersion()
}

func (tw *LazyTreeWrapper) Reset() {
	tw.lazyInit()
	tw.tree.Reset()
}

func (tw *LazyTreeWrapper) Commit(recentVersion *bsmt.Version) (bsmt.Version, error) {
	tw.lazyInit()
	return tw.tree.Commit(recentVersion)
}

func (tw *LazyTreeWrapper) Rollback(version bsmt.Version) error {
	tw.lazyInit()
	return tw.tree.Rollback(version)
}
