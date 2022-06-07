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
	ChainConfig struct {
		NetworkRPCSysConfigName             string
		GovernanceContractAddrSysConfigName string
		StartL1BlockHeight                  int64
		PendingBlocksCount                  uint64
		MaxHandledBlocksCount               int64
	}
	LogConf logx.LogConf
}
