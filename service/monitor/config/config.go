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
	}
	LogConf logx.LogConf
}
