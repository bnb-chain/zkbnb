package treedb

import (
	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/bas-smt/database"
)

const (
	defaultBatchReloadSize = 1000
)

type Context struct {
	Name          string
	Driver        Driver
	LevelDBOption *LevelDBOption
	RedisDBOption *RedisDBOption

	TreeDB          database.TreeDB
	defaultOptions  []bsmt.Option
	Reload          bool
	batchReloadSize int
}

func (ctx *Context) IsLoad() bool {
	if ctx.Reload {
		return true
	}
	return ctx.Driver == MemoryDB
}

func (ctx *Context) Options(blockHeight int64) []bsmt.Option {
	var opts []bsmt.Option
	for i := range ctx.defaultOptions {
		opts = append(opts, ctx.defaultOptions[i])
	}
	if ctx.Driver == MemoryDB {
		opts = append(opts, bsmt.InitializeVersion(bsmt.Version(blockHeight)))
	}
	return opts
}

func (ctx *Context) SetOptions(opts ...bsmt.Option) {
	ctx.defaultOptions = append(ctx.defaultOptions, opts...)
}

func (ctx *Context) BatchReloadSize() int {
	if ctx.batchReloadSize <= 0 {
		return defaultBatchReloadSize // default
	}

	return ctx.batchReloadSize
}

func (ctx *Context) SetBatchReloadSize(size int) {
	ctx.batchReloadSize = size
}
