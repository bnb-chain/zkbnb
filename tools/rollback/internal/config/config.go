package config

import (
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm/logger"
)

type Config struct {
	Postgres struct {
		MasterDataSource string
		SlaveDataSource  string
		LogLevel         logger.LogLevel `json:",optional"`
	}
	TreeDB struct {
		Driver tree.Driver
		//nolint:staticcheck
		LevelDBOption tree.LevelDBOption `json:",optional"`
		//nolint:staticcheck
		RedisDBOption tree.RedisDBOption `json:",optional"`
		//nolint:staticcheck
		RoutinePoolSize    int `json:",optional"`
		AssetTreeCacheSize int
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
