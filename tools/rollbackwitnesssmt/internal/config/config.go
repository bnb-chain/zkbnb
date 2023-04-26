package config

import (
	"github.com/bnb-chain/zkbnb/common/apollo"
	witnessConfig "github.com/bnb-chain/zkbnb/service/witness/config"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

func InitSystemConfiguration(config *witnessConfig.Config, configFile string) error {
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

func InitSystemConfigFromConfigFile(c *witnessConfig.Config, configFile string) error {
	conf.Load(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
