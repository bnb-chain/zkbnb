package svc

import (
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/plugin/dbresolver"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/cache"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/fetcher/price"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/fetcher/state"
)

type ServiceContext struct {
	Config     config.Config
	RedisCache dbcache.Cache
	MemCache   *cache.MemCache

	DB                  *gorm.DB
	TxPoolModel         tx.TxPoolModel
	AccountModel        account.AccountModel
	AccountHistoryModel account.AccountHistoryModel
	TxModel             tx.TxModel
	BlockModel          block.BlockModel
	NftModel            nft.L2NftModel
	AssetModel          asset.AssetModel
	SysConfigModel      sysconfig.SysConfigModel

	PriceFetcher       price.Fetcher
	StateFetcher       state.Fetcher
	SendTxMetrics      prometheus.Counter
	SendTxTotalMetrics prometheus.Counter
}

func NewServiceContext(c config.Config) *ServiceContext {

	masterDataSource := c.Postgres.MasterDataSource
	slaveDataSource := c.Postgres.SlaveDataSource
	db, err := gorm.Open(postgres.Open(masterDataSource))
	if err != nil {
		logx.Must(err)
	}

	db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDataSource)},
		Replicas: []gorm.Dialector{postgres.Open(slaveDataSource)},
	}))

	rawDB, err := db.DB()
	if err != nil {
		logx.Must(err)
	}
	rawDB.SetMaxOpenConns(c.Postgres.MaxConn)
	rawDB.SetMaxIdleConns(c.Postgres.MaxIdle)

	redisCache := dbcache.NewRedisCache(c.CacheRedis[0].Host, c.CacheRedis[0].Pass, 15*time.Minute)

	txPoolModel := tx.NewTxPoolModel(db)
	accountModel := account.NewAccountModel(db)
	nftModel := nft.NewL2NftModel(db)
	assetModel := asset.NewAssetModel(db)
	if c.MemCache.TxPendingExpiration == 0 {
		c.MemCache.TxPendingExpiration = 60000
	}
	memCache := cache.MustNewMemCache(accountModel, assetModel, c.MemCache.AccountExpiration, c.MemCache.BlockExpiration,
		c.MemCache.TxExpiration, c.MemCache.AssetExpiration, c.MemCache.TxPendingExpiration, c.MemCache.PriceExpiration, c.MemCache.MaxCounterNum, c.MemCache.MaxKeyNum)

	sendTxMetrics := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "sent_tx_count",
		Help:      "sent tx count",
	})

	sendTxTotalMetrics := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "sent_tx_total_count",
		Help:      "sent tx total count",
	})

	prometheus.Register(sendTxMetrics)
	prometheus.Register(sendTxTotalMetrics)

	return &ServiceContext{
		Config:              c,
		RedisCache:          redisCache,
		MemCache:            memCache,
		DB:                  db,
		TxPoolModel:         txPoolModel,
		AccountModel:        accountModel,
		AccountHistoryModel: account.NewAccountHistoryModel(db),
		TxModel:             tx.NewTxModel(db),
		BlockModel:          block.NewBlockModel(db),
		NftModel:            nftModel,
		AssetModel:          assetModel,
		SysConfigModel:      sysconfig.NewSysConfigModel(db),

		PriceFetcher:       price.NewFetcher(memCache, assetModel, c.CoinMarketCap.Url, c.CoinMarketCap.Token),
		StateFetcher:       state.NewFetcher(redisCache, accountModel, nftModel),
		SendTxMetrics:      sendTxMetrics,
		SendTxTotalMetrics: sendTxTotalMetrics,
	}
}

func (s *ServiceContext) Shutdown() {
	sqlDB, err := s.DB.DB()
	if err != nil {
		_ = sqlDB.Close()
	}
	_ = s.RedisCache.Close()
	s.PriceFetcher.Stop()
}
