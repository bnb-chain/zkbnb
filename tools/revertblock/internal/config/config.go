package config

import (
	"github.com/bnb-chain/zkbnb/common/apollo"
	senderConfig "github.com/bnb-chain/zkbnb/service/sender/config"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

type ChainConfig struct {
	NetworkRPCSysConfigName string
	GasLimit                uint64
	GasPrice                uint64
	MaxWaitingTime          int64
	ConfirmBlocksCount      uint64
}

type Config struct {
	Postgres    apollo.Postgres
	ChainConfig ChainConfig
	AuthConfig  senderConfig.AuthConfig
	KMSConfig   senderConfig.KMSConfig
	LogConf     logx.LogConf
}

func InitSystemConfiguration(config *Config, configFile string) error {
	if err := InitSystemConfigFromConfigFile(config, configFile); err != nil {
		return err
	}
	if config.Postgres.MasterSecretKey != "" && config.Postgres.SlaveSecretKey != "" {
		logx.Infof("replace database password by aws secret key")
		commonConfig := &apollo.CommonConfig{}
		commonConfig.Postgres = config.Postgres
		commonConfig, err := apollo.BuildCommonConfig(commonConfig)
		if err != nil {
			return err
		}
		config.Postgres = commonConfig.Postgres
	}
	return nil
}

func InitSystemConfigFromConfigFile(c *Config, configFile string) error {
	conf.Load(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
