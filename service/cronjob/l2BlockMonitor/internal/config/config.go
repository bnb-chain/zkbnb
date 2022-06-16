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
		// bsc
		BSCNetworkRPCSysConfigName string
		BSCPendingBlocksCount      uint64
	}
}
