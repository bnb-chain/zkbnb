package config

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

const (
	ProverAppId     = "zkbnb-prover"
	SystemConfigKey = "SystemConfig"
	Namespace       = "application"
)

type BlockConfig struct {
	OptionalBlockSizes []int
	R1CSBatchSize      int
}

type Config struct {
	Postgres    apollo.Postgres
	CacheRedis  cache.CacheConf
	LogConf     logx.LogConf
	KeyPath     []string
	BlockConfig BlockConfig
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
	c.CacheRedis = commonConfig.CacheRedis

	systemConfigString, err := apollo.LoadApolloConfigFromEnvironment(ProverAppId, Namespace, SystemConfigKey)
	if err != nil {
		return err
	}

	systemConfig := &Config{}
	err = json.Unmarshal([]byte(systemConfigString), systemConfig)
	if err != nil {
		return err
	}

	c.LogConf = systemConfig.LogConf
	c.KeyPath = systemConfig.KeyPath
	c.BlockConfig = systemConfig.BlockConfig

	return nil
}

func InitSystemConfigFromConfigFile(c *Config, configFile string) error {
	conf.MustLoad(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
