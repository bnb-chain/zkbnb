package config

import (
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm/logger"

	"github.com/bnb-chain/zkbnb/tree"
)

type Config struct {
	Postgres struct {
		MasterDataSource string
		LogLevel         logger.LogLevel `json:",optional"`
	}
	ChainConfig struct {
		StartL1BlockHeight               int64
		ConfirmBlocksCount               uint64
		MaxWaitingTime                   int64
		MaxHandledBlocksCount            int64
		MaxCancelOutstandingDepositCount int64
		KeptHistoryBlocksCount           int64 // KeptHistoryBlocksCount define the count of blocks to keep in table, old blocks will be cleaned
		BscTestNetRpc                    string
		ZkBnbContractAddress             string
		GovernanceContractAddress        string
		GasLimit                         uint64
		PrivateKey                       string `json:",optional"`
	}
	CacheConfig statedb.CacheConfig `json:",optional"`

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
	LogConf       logx.LogConf
	KeyPath       string `json:",optional"`
	R1CSBatchSize int    `json:",optional"`

	Address     string `json:",optional"`
	Token       string `json:",optional"`
	NftIndex    int64  `json:",optional"`
	ProofFolder string `json:",optional"`
}
