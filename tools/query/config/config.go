package config

import (
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"

	"github.com/bnb-chain/zkbnb/tree"
)

type Config struct {
	Postgres   apollo.Postgres
	CacheRedis cache.CacheConf
	TreeDB     struct {
		Driver tree.Driver
		//nolint:staticcheck
		LevelDBOption tree.LevelDBOption `json:",optional"`
		//nolint:staticcheck
		RedisDBOption tree.RedisDBOption `json:",optional"`
		//nolint:staticcheck
		RoutinePoolSize    int `json:",optional"`
		AssetTreeCacheSize int
	}
	EnableRollback bool
	DbRoutineSize  int `json:",optional"`
	DbBatchSize    int
	LogConf        logx.LogConf
}
