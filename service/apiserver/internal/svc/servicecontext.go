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
	"github.com/bnb-chain/zkbnb/dao/mempool"
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

	MempoolModel          mempool.MempoolModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	TxModel               tx.TxModel
	TxDetailModel         tx.TxDetailModel
	FailTxModel           tx.FailTxModel
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
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Must(err)
	}
	redisCache := dbcache.NewRedisCache(c.CacheRedis[0].Host, c.CacheRedis[0].Pass, 15*time.Minute)

	mempoolModel := mempool.NewMempoolModel(gormPointer)
	accountModel := account.NewAccountModel(gormPointer)
	liquidityModel := liquidity.NewLiquidityModel(gormPointer)
	nftModel := nft.NewL2NftModel(gormPointer)
	assetModel := asset.NewAssetModel(gormPointer)
	memCache := cache.NewMemCache(accountModel, assetModel, c.MemCache.AccountExpiration, c.MemCache.BlockExpiration,
		c.MemCache.TxExpiration, c.MemCache.AssetExpiration, c.MemCache.PriceExpiration)
	return &ServiceContext{
		Config:                c,
		RedisCache:            redisCache,
		MemCache:              memCache,
		MempoolModel:          mempoolModel,
		AccountModel:          accountModel,
		AccountHistoryModel:   account.NewAccountHistoryModel(gormPointer),
		TxModel:               tx.NewTxModel(gormPointer),
		TxDetailModel:         tx.NewTxDetailModel(gormPointer),
		FailTxModel:           tx.NewFailTxModel(gormPointer),
		LiquidityModel:        liquidityModel,
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(gormPointer),
		BlockModel:            block.NewBlockModel(gormPointer),
		NftModel:              nftModel,
		AssetModel:            assetModel,
		SysConfigModel:        sysconfig.NewSysConfigModel(gormPointer),

		PriceFetcher: price.NewFetcher(memCache, c.CoinMarketCap.Url, c.CoinMarketCap.Token),
		StateFetcher: state.NewFetcher(redisCache, accountModel, liquidityModel, nftModel),
	}
}
