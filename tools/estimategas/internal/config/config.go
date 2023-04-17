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
		CommitBlockSk           string
		GasLimit                uint64
	}
	LogConf logx.LogConf
}
