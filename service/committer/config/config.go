package config

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	CommitterAppId  = "zkbnb-committer"
	SystemConfigKey = "SystemConfig"
	Namespace       = "application"
)

type BlockConfig struct {
	//second
	MaxPackedInterval      int  `json:",optional"`
	SaveBlockDataPoolSize  int  `json:",optional"`
	RollbackOnly           bool `json:",optional"`
	DisableLoadAllAccounts bool `json:",optional"`
}

type Config struct {
	core.ChainConfig

	BlockConfig    BlockConfig
	LogConf        logx.LogConf
	IpfsUrl        string
	EnableRollback bool
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
	commonConfig, err := apollo.InitCommonConfig(CommitterAppId)
	if err != nil {
		return err
	}
	c.Postgres = commonConfig.Postgres
	c.CacheRedis = commonConfig.CacheRedis
	c.EnableRollback = commonConfig.EnableRollback

	systemConfigString, err := apollo.LoadApolloConfigFromEnvironment(CommitterAppId, Namespace, SystemConfigKey)
	if err != nil {
		return err
	}

	systemConfig := &Config{}
	err = json.Unmarshal([]byte(systemConfigString), systemConfig)
	if err != nil {
		return err
	}
	c.RedisExpiration = systemConfig.RedisExpiration
	c.CacheConfig = systemConfig.CacheConfig
	c.BlockConfig = systemConfig.BlockConfig
	c.TreeDB = systemConfig.TreeDB
	c.LogConf = systemConfig.LogConf
	c.IpfsUrl = systemConfig.IpfsUrl
	c.DbBatchSize = systemConfig.DbBatchSize
	c.DbRoutineSize = systemConfig.DbRoutineSize
	return nil
}

func InitSystemConfigFromConfigFile(c *Config, configFile string) error {
	conf.Load(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
