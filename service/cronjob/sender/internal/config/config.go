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
		NetworkRPCSysConfigName         string
		ZecreyContractAddrSysConfigName string
		MaxWaitingTime                  int64
		MaxBlockCount                   int
		Sk                              string
		GasLimit                        uint64
		L1ChainId                       string
	}
}
