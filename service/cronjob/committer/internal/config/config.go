package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"

	"github.com/bnb-chain/zkbas/pkg/treedb"
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
	LogConf logx.LogConf
}
