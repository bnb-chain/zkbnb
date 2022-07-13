package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	KeyPath struct {
		KeyTxCounts []int
	}
	CacheRedis cache.CacheConf
}
