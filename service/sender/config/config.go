package config

import (
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	PrivateKeySignMode = "PrivateKeySignMode"
	KeyManageSignMode  = "KeyManageSignMode"
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
		SendSignatureMode       string
		GasLimit                uint64
		GasPrice                uint64
		//((MaxGasPrice-GasPrice)/GasPrice)*100
		MaxGasPriceIncreasePercentage uint64 `json:",optional"`
	}
	AuthConfig struct {
		CommitBlockSk string
		VerifyBlockSk string
	}
	KMSConfig struct {
		CommitKeyId string
		VerifyKeyId string
	}
	LogConf logx.LogConf
}
