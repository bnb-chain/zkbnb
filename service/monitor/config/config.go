package config

import (
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"os"
	"strconv"
)

const (
	MonitorAppId       = "zkbnb-monitor"
	SystemConfigKey    = "SystemConfig"
	Namespace          = "application"
	StartL1BlockHeight = "START_L1_BLOCK_HEIGHT"
)

type ChainConfig struct {
	NetworkRPCSysConfigName string
	StartL1BlockHeight      int64 `json:",optional"`
	ConfirmBlocksCount      uint64
	MaxHandledBlocksCount   int64
	KeptHistoryBlocksCount  int64 // KeptHistoryBlocksCount define the count of blocks to keep in table, old blocks will be cleaned
}

type Config struct {
	Postgres         apollo.Postgres
	CacheRedis       cache.CacheConf
	ChainConfig      ChainConfig
	LogConf          logx.LogConf
	AccountCacheSize int
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

	systemConfigString, err := apollo.LoadApolloConfigFromEnvironment(MonitorAppId, Namespace, SystemConfigKey)
	if err != nil {
		return err
	}

	systemConfig := &Config{}
	err = json.Unmarshal([]byte(systemConfigString), systemConfig)
	if err != nil {
		return err
	}

	startL1BlockHeight, err := LoadStartL1BlockHeightFromEnvironment()
	if err != nil {
		return err
	}

	c.ChainConfig = systemConfig.ChainConfig
	c.LogConf = systemConfig.LogConf
	c.AccountCacheSize = systemConfig.AccountCacheSize
	c.ChainConfig.StartL1BlockHeight = startL1BlockHeight

	return nil
}

func LoadStartL1BlockHeightFromEnvironment() (int64, error) {
	startL1BlockHeightStr := os.Getenv(StartL1BlockHeight)
	if len(startL1BlockHeightStr) == 0 {
		return 0, fmt.Errorf("environment variable START_L1_BLOCK_HEIGHT is not set")
	}

	startL1BlockHeight, err := strconv.ParseInt(startL1BlockHeightStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("environment variable START_L1_BLOCK_HEIGHT is not set correctly, %v", err)
	}
	return startL1BlockHeight, nil
}

func InitSystemConfigFromConfigFile(c *Config, configFile string) error {
	conf.MustLoad(configFile, c)
	logx.MustSetup(c.LogConf)
	c.Validate()
	logx.DisableStat()
	return nil
}

func (c *Config) Validate() {
	if c.ChainConfig.StartL1BlockHeight < 0 || c.ChainConfig.MaxHandledBlocksCount <= 0 || c.ChainConfig.KeptHistoryBlocksCount <= 0 {
		logx.Severe("invalid chain config")
		panic("invalid chain config")
	}
}
