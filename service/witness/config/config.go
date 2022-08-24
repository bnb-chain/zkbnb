package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"

	"github.com/bnb-chain/zkbas/tree"
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
	TreeDB     struct {
		Driver        tree.Driver
		LevelDBOption tree.LevelDBOption `json:",optional"`
		RedisDBOption tree.RedisDBOption `json:",optional"`
	}
	LogConf logx.LogConf
}