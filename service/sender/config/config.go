package config

import (
	"github.com/zeromicro/go-zero/core/logx"
)

type Config struct {
	Postgres struct {
		MasterDataSource string
		SlaveDataSource  string
	}
	ChainConfig struct {
		NetworkRPCSysConfigName string
		MaxWaitingTime          int64
		MaxBlockCount           int
		ConfirmBlocksCount      uint64
		CommitBlockSk           string
		VerifyBlockSk           string
		GasLimit                uint64
		GasPrice                uint64
	}
	LogConf logx.LogConf
}
