package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	CacheRedis  cache.CacheConf
	ChainConfig struct {
		NetworkRPCSysConfigName string
		MaxWaitingTime          int64
		MaxBlockCount           int
		ConfirmBlocksCount      uint64
		Sk                      string
		GasLimit                uint64
	}
	LogConf logx.LogConf
}
