package config

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	KmsCommitKeyId = "KMS_COMMIT_KEY_ID"
	KmsVerifyKeyId = "KMS_VERIFY_KEY_ID"

	AuthCommitBlockSk = "AUTH_COMMIT_BLOCK_SK"
	AuthVerifyBlockSk = "AUTH_COMMIT_BLOCK_SK"
)

const (
	SenderAppId     = "zkbnb-sender"
	SystemConfigKey = "SystemConfig"
	Namespace       = "application"
)

type ChainConfig struct {
	NetworkRPCSysConfigName string
	MaxWaitingTime          int64
	ConfirmBlocksCount      uint64
	GasLimit                uint64
	GasPrice                uint64
	//((MaxGasPrice-GasPrice)/GasPrice)*100
	MaxGasPriceIncreasePercentage uint64 `json:",optional"`
	DisableCommitBlock            bool   `json:",optional"`
	DisableVerifyBlock            bool   `json:",optional"`
}

type AuthConfig struct {
	CommitBlockSk string `json:",optional"`
	VerifyBlockSk string `json:",optional"`
}

type KMSConfig struct {
	CommitKeyId string `json:",optional"`
	VerifyKeyId string `json:",optional"`
}

type Config struct {
	Postgres    apollo.Postgres
	ChainConfig ChainConfig
	AuthConfig  AuthConfig
	KMSConfig   KMSConfig
	LogConf     logx.LogConf
}

func InitSystemConfiguration(config *Config, configFile string) error {
	if err := InitSystemConfigFromEnvironment(config); err != nil {
		logx.Errorf("Init system configuration from environment raise error: %v", err)
	} else {
		logx.Infof("Init system configuration from environment Successfully")
		return nil
	}
	if err := InitSystemConfigFromConfigFile(config, configFile); err != nil {
		logx.Errorf("Init system configuration from config file raise error: %v", err)
		panic("Init system configuration from config file raise error:" + err.Error())
	} else {
		logx.Infof("Init system configuration from config file Successfully")
		return nil
	}
}

func InitSystemConfigFromEnvironment(c *Config) error {
	commonConfig, err := apollo.InitCommonConfig(SenderAppId)
	if err != nil {
		return err
	}
	c.Postgres = commonConfig.Postgres

	systemConfigString, err := apollo.LoadApolloConfigFromEnvironment(SenderAppId, Namespace, SystemConfigKey)
	if err != nil {
		return err
	}

	systemConfig := &Config{}
	err = json.Unmarshal([]byte(systemConfigString), systemConfig)
	if err != nil {
		return err
	}

	c.ChainConfig = systemConfig.ChainConfig
	c.KMSConfig = systemConfig.KMSConfig
	c.AuthConfig = systemConfig.AuthConfig
	c.LogConf = systemConfig.LogConf

	return nil
}

func InitSystemConfigFromConfigFile(c *Config, configFile string) error {
	conf.Load(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
