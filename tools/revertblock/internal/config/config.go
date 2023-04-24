package config

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	RevertBlockAppId = "zkbnb-revertblock"
	SystemConfigKey  = "SystemConfig"
	Namespace        = "application"
)

type ChainConfig struct {
	NetworkRPCSysConfigName string
	RevertBlockSk           string
	GasLimit                uint64
	GasPrice                uint64
	MaxWaitingTime          int64
	ConfirmBlocksCount      uint64
}

type Config struct {
	Postgres    apollo.Postgres
	ChainConfig ChainConfig
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
	commonConfig, err := apollo.InitCommonConfig()
	if err != nil {
		return err
	}
	c.Postgres = commonConfig.Postgres

	systemConfigString, err := apollo.LoadApolloConfigFromEnvironment(RevertBlockAppId, Namespace, SystemConfigKey)
	if err != nil {
		return err
	}

	systemConfig := &Config{}
	err = json.Unmarshal([]byte(systemConfigString), systemConfig)
	if err != nil {
		return err
	}

	c.LogConf = systemConfig.LogConf
	c.ChainConfig = systemConfig.ChainConfig

	return nil
}

func InitSystemConfigFromConfigFile(c *Config, configFile string) error {
	conf.Load(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
