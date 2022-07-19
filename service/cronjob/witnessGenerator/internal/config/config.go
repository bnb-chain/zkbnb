package config

import (
	"github.com/bnb-chain/zkbas/pkg/treedb"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
	TreeDB     struct {
		Driver        treedb.Driver
		LevelDBOption treedb.LevelDBOptions `yaml:",optional"`
		RedisDBOption treedb.RedisDBOptions `yaml:",optional"`
	}
}
