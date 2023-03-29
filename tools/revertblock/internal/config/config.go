package config

import (
	"github.com/zeromicro/go-zero/core/logx"
)

type Config struct {
	Postgres struct {
		MasterDataSource string
	}
	ChainConfig struct {
		NetworkRPCSysConfigName string
		RevertBlockSk           string
		GasLimit                uint64
		GasPrice                uint64
		MaxWaitingTime          int64
		ConfirmBlocksCount      uint64
	}
	LogConf logx.LogConf
}
