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
		ConfirmBlocksCount      uint64
		SendSignatureMode       string
		GasLimit                uint64
		GasPrice                uint64
		//((MaxGasPrice-GasPrice)/GasPrice)*100
		MaxGasPriceIncreasePercentage uint64 `json:",optional"`
	}
	Apollo struct {
		AppID          string
		Cluster        string
		ApolloIp       string
		Namespace      string
		IsBackupConfig bool
	}
	AuthConfig struct {
		CommitBlockSk string `json:",optional"`
		VerifyBlockSk string `json:",optional"`
	}
	KMSConfig struct {
		CommitKeyId string `json:",optional"`
		VerifyKeyId string `json:",optional"`
	}
	LogConf logx.LogConf
}
