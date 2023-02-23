package config

import (
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/tree"
)

type Config struct {
	Postgres struct {
		MasterDataSource string
	}
	ChainConfig struct {
		StartL1BlockHeight     int64
		ConfirmBlocksCount     uint64
		MaxHandledBlocksCount  int64
		KeptHistoryBlocksCount int64 // KeptHistoryBlocksCount define the count of blocks to keep in table, old blocks will be cleaned
		BscTestNetRpc          string
		ZkBnbContractAddress   string
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
	LogConf logx.LogConf
}
