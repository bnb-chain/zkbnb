package config

import (
	"github.com/zeromicro/go-zero/core/logx"
)

type Config struct {
	ChainConfig struct {
		StartL1BlockHeight               int64
		EndL1BlockHeight                 int64
		ConfirmBlocksCount               uint64
		MaxHandledBlocksCount            int64
		MaxCancelOutstandingDepositCount int64
		KeptHistoryBlocksCount           int64 // KeptHistoryBlocksCount define the count of blocks to keep in table, old blocks will be cleaned
		BscTestNetRpc                    string
		ZkBnbContractAddress             string
		GasLimit                         uint64
		PrivateKey                       string `json:",optional"`
	}
	LogConf logx.LogConf
}
