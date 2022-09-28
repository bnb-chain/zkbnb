package svc

import (
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
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

	DB                    *gorm.DB
	TxPoolModel           tx.TxPoolModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	TxModel               tx.TxModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	BlockModel            block.BlockModel
	NftModel              nft.L2NftModel
	AssetModel            asset.AssetModel
	SysConfigModel        sysconfig.SysConfigModel

	PriceFetcher price.Fetcher
	StateFetcher state.Fetcher
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Must(err)
	}
	redisCache := dbcache.NewRedisCache(c.CacheRedis[0].Host, c.CacheRedis[0].Pass, 15*time.Minute)

	txPoolModel := tx.NewTxPoolModel(db)
	accountModel := account.NewAccountModel(db)
	liquidityModel := liquidity.NewLiquidityModel(db)
	nftModel := nft.NewL2NftModel(db)
	assetModel := asset.NewAssetModel(db)
	memCache := cache.NewMemCache(accountModel, assetModel, c.MemCache.CacheSize)
	return &ServiceContext{
		Config:                c,
		RedisCache:            redisCache,
		MemCache:              memCache,
		DB:                    db,
		TxPoolModel:           txPoolModel,
		AccountModel:          accountModel,
		AccountHistoryModel:   account.NewAccountHistoryModel(db),
		TxModel:               tx.NewTxModel(db),
		LiquidityModel:        liquidityModel,
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(db),
		BlockModel:            block.NewBlockModel(db),
		NftModel:              nftModel,
		AssetModel:            assetModel,
		SysConfigModel:        sysconfig.NewSysConfigModel(db),

		PriceFetcher: price.NewFetcher(memCache, assetModel, c.CoinMarketCap.Url, c.CoinMarketCap.Token),
		StateFetcher: state.NewFetcher(redisCache, accountModel, liquidityModel, nftModel),
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
