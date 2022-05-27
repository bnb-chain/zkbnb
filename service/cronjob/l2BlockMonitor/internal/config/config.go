package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Postgres struct {
		DataSource string
	}
	CacheRedis  cache.CacheConf
	LogConf     logx.LogConf
	ChainConfig struct {
		// bsc
		BSCNetworkRPCSysConfigName string
		BSCPendingBlocksCount      uint64
	}
}
