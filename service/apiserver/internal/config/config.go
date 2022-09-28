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
	TxPool struct {
		MaxPendingTxCount int
	}
	CacheRedis    cache.CacheConf
	LogConf       logx.LogConf
	CoinMarketCap struct {
		Url   string
		Token string
	}
	MemCache struct {
		AccountExpiration int
		AssetExpiration   int
		BlockExpiration   int
		TxExpiration      int
		PriceExpiration   int
	}
}
