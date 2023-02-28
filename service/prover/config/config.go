package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
	Postgres struct {
		MasterDataSource string
		SlaveDataSource  string
	}
	CacheRedis  cache.CacheConf
	LogConf     logx.LogConf
	KeyPath     []string
	BlockConfig struct {
		OptionalBlockSizes []int
	}
}
