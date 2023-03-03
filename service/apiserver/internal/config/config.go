package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Postgres struct {
		MasterDataSource string
		SlaveDataSource  string
		MaxIdle          int
		MaxConn          int
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
	IpfsUrl  string
	MemCache struct {
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
	Apollo struct {
		AppID          string
		Cluster        string
		ApolloIp       string
		Namespace      string
		IsBackupConfig bool
	}
}
