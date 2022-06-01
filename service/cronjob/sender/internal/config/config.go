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
	LogConf     logx.LogConf
	ChainConfig struct {
		L2ChainId                       int64
		NetworkRPCSysConfigName         string
		ZecreyContractAddrSysConfigName string
		MaxWaitingTime                  int64
		MaxBlockCount                   int
		MainChainIdSysConfigName        string
		Sk                              string
		GasLimit                        uint64
		L1ChainId                       string
	}
}
