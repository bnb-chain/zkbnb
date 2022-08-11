package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	CacheRedis  cache.CacheConf
	ChainConfig struct {
		NetworkRPCSysConfigName string
		StartL1BlockHeight      int64
		PendingBlocksCount      uint64
		MaxHandledBlocksCount   int64
	}
}
