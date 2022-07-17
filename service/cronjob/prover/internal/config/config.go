package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
	KeyPath struct {
		ProvingKeyPath   []string
		VerifyingKeyPath []string
		KeyTxCounts      []int
	}
	Postgres struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
}
