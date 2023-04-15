package config

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

const (
	ApiServerAppId  = "zkbnb-apiserver"
	SystemConfigKey = "SystemConfig"
	Namespace       = "application"
)

type TxPool struct {
	MaxPendingTxCount int
}

type CoinMarketCap struct {
	Url   string
	Token string
}

type BinanceOracle struct {
	Url       string
	Apikey    string
	ApiSecret string
}

type MemCache struct {
	AccountExpiration   int
	AssetExpiration     int
	BlockExpiration     int
	TxExpiration        int
	PriceExpiration     int
	TxPendingExpiration int `json:",optional"`
	// Number of 4-bit access counters to keep for admission and eviction
	// Setting this to 10x the number of items you expect to keep in the cache when full
	MaxCounterNum int64
	MaxKeyNum     int64
}

type Config struct {
	rest.RestConf
	Postgres      apollo.Postgres
	TxPool        TxPool
	CacheRedis    cache.CacheConf
	LogConf       logx.LogConf
	CoinMarketCap CoinMarketCap
	BinanceOracle BinanceOracle
	IpfsUrl       string
	MemCache      MemCache
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

	systemConfigString, err := apollo.LoadApolloConfigFromEnvironment(ApiServerAppId, Namespace, SystemConfigKey)
	if err != nil {
		return err
	}

	systemConfig := &Config{}
	err = json.Unmarshal([]byte(systemConfigString), systemConfig)
	if err != nil {
		return err
	}
	c.RestConf = systemConfig.RestConf
	c.TxPool = systemConfig.TxPool
	c.LogConf = systemConfig.LogConf
	c.CoinMarketCap = systemConfig.CoinMarketCap
	c.BinanceOracle = systemConfig.BinanceOracle
	c.IpfsUrl = systemConfig.IpfsUrl
	c.MemCache = systemConfig.MemCache

	return nil
}

func InitSystemConfigFromConfigFile(c *Config, configFile string) error {
	conf.MustLoad(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
