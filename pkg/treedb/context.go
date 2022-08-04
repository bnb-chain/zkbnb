package treedb

import (
	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/bas-smt/database"
)

type Context struct {
	Name          string
	Driver        Driver
	LevelDBOption *LevelDBOption
	RedisDBOption *RedisDBOption

	TreeDB database.TreeDB
}

func (ctx *Context) IsLoad() bool {
	return ctx.Driver == MemoryDB
}

func (ctx *Context) Options(blockHeight int64) []bsmt.Option {
	var opts []bsmt.Option
	if ctx.Driver == MemoryDB {
		opts = append(opts, bsmt.InitializeVersion(bsmt.Version(blockHeight)))
	}
	return opts
}
