package config

import (
	"github.com/bnb-chain/zkbnb/common/apollo"
	senderConfig "github.com/bnb-chain/zkbnb/service/sender/config"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

type Config struct {
	Postgres apollo.Postgres
	TreeDB   struct {
		Driver tree.Driver
		//nolint:staticcheck
		LevelDBOption tree.LevelDBOption `json:",optional"`
		//nolint:staticcheck
		RedisDBOption tree.RedisDBOption `json:",optional"`
		//nolint:staticcheck
		RoutinePoolSize    int `json:",optional"`
		AssetTreeCacheSize int
	}
	ChainConfig struct {
		NetworkRPCSysConfigName string
		GasLimit                uint64
		GasPrice                uint64
		MaxWaitingTime          int64
		ConfirmBlocksCount      uint64
	}
	AuthConfig senderConfig.AuthConfig
	KMSConfig  senderConfig.KMSConfig
	LogConf    logx.LogConf
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
