package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Postgres struct {
		DataSource string
	}
	CacheRedis    cache.CacheConf
	LogConf       logx.LogConf
	CoinMarketCap struct {
		Url   string
		Token string
	}
	MaxPendingTxCount int
	MemCache          struct {
		AccountExpiration int
		AssetExpiration   int
		BlockExpiration   int
		TxExpiration      int
		PriceExpiration   int
	}
}
