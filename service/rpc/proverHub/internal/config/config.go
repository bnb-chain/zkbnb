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
	KeyPath struct {
		VerifyingKeyPath []string
		VerifyingKeyTxsCount []int
	}
	CacheRedis cache.CacheConf
	LogConf    logx.LogConf
}
