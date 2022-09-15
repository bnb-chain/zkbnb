package config

import (
	"github.com/zeromicro/go-zero/core/logx"
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	ChainConfig struct {
		NetworkRPCSysConfigName string
		StartL1BlockHeight      int64
		ConfirmBlocksCount      uint64
		MaxHandledBlocksCount   int64
		KeptHistoryBlocksCount  int64 // KeptHistoryBlocksCount define the count of blocks to keep in table, old blocks will be cleaned
	}
	LogConf logx.LogConf
}
