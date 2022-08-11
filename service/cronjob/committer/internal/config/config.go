package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"

	"github.com/bnb-chain/zkbas/common/treedb"
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	KeyPath struct {
		KeyTxCounts []int
	}
	CacheRedis cache.CacheConf
	TreeDB     struct {
		Driver        treedb.Driver
		LevelDBOption treedb.LevelDBOption `json:",optional"`
		RedisDBOption treedb.RedisDBOption `json:",optional"`
	}
}
