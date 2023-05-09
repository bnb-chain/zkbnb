package config

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"strconv"
)

const (
	ProverAppId     = "zkbnb-prover"
	SystemConfigKey = "SystemConfig"
	Namespace       = "application"
)

type BlockConfig struct {
	R1CSBatchSize int
}

type Config struct {
	Postgres    apollo.Postgres
	CacheRedis  cache.CacheConf
	LogConf     logx.LogConf
	KeyPath     []string
	BlockConfig BlockConfig
}

func InitSystemConfiguration(config *Config, configFile string, proverId uint) error {
	if err := InitSystemConfigFromEnvironment(config, proverId); err != nil {
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

func InitSystemConfigFromEnvironment(c *Config, proverId uint) error {
	commonConfig, err := apollo.InitCommonConfig(apollo.CommonAppId)
	if err != nil {
		return err
	}
	c.Postgres = commonConfig.Postgres
	c.CacheRedis = commonConfig.CacheRedis

	configKey := SystemConfigKey
	if proverId > 0 {
		configKey = SystemConfigKey + "-" + strconv.Itoa(int(proverId))
	}
	systemConfigString, err := apollo.LoadApolloConfigFromEnvironment(ProverAppId, Namespace, configKey)
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
	conf.Load(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
