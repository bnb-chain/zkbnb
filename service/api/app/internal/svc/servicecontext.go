package svc

import (
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/dbcache"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/asset"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/sysConfig"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/cache"
	"github.com/bnb-chain/zkbas/service/api/app/internal/config"
	"github.com/bnb-chain/zkbas/service/api/app/internal/fetcher/price"
	"github.com/bnb-chain/zkbas/service/api/app/internal/fetcher/state"
)

type ServiceContext struct {
	Config      config.Config
	Conn        sqlx.SqlConn
	GormPointer *gorm.DB
	RedisCache  dbcache.Cache
	MemCache    *cache.MemCache

	MempoolModel          mempool.MempoolModel
	MempoolDetailModel    mempool.MempoolTxDetailModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	TxModel               tx.TxModel
	TxDetailModel         tx.TxDetailModel
	FailTxModel           tx.FailTxModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	BlockModel            block.BlockModel
	NftModel              nft.L2NftModel
	CollectionModel       nft.L2NftCollectionModel
	OfferModel            nft.OfferModel
	AssetModel            asset.AssetModel
	SysConfigModel        sysConfig.SysConfigModel

	PriceFetcher price.Fetcher
	StateFetcher state.Fetcher
}

func NewServiceContext(c config.Config) *ServiceContext {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Must(err)
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisCache := dbcache.NewRedisCache(c.CacheRedis[0].Host, c.CacheRedis[0].Pass, 15*time.Minute)

	mempoolModel := mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer)
	mempoolDetailModel := mempool.NewMempoolDetailModel(conn, c.CacheRedis, gormPointer)
	accountModel := account.NewAccountModel(conn, c.CacheRedis, gormPointer)
	liquidityModel := liquidity.NewLiquidityModel(conn, c.CacheRedis, gormPointer)
	nftModel := nft.NewL2NftModel(conn, c.CacheRedis, gormPointer)
	offerModel := nft.NewOfferModel(conn, c.CacheRedis, gormPointer)
	assetModel := asset.NewAssetModel(conn, c.CacheRedis, gormPointer)
	memCache := cache.NewMemCache(accountModel, assetModel, c.MemCache.AccountExpiration, c.MemCache.BlockExpiration,
		c.MemCache.TxExpiration, c.MemCache.AssetExpiration, c.MemCache.PriceExpiration)
	return &ServiceContext{
		Config:                c,
		Conn:                  conn,
		GormPointer:           gormPointer,
		RedisCache:            redisCache,
		MemCache:              memCache,
		MempoolModel:          mempoolModel,
		MempoolDetailModel:    mempoolDetailModel,
		AccountModel:          accountModel,
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
		TxModel:               tx.NewTxModel(conn, c.CacheRedis, gormPointer),
		TxDetailModel:         tx.NewTxDetailModel(conn, c.CacheRedis, gormPointer),
		FailTxModel:           tx.NewFailTxModel(conn, c.CacheRedis, gormPointer),
		LiquidityModel:        liquidityModel,
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(conn, c.CacheRedis, gormPointer),
		BlockModel:            block.NewBlockModel(conn, c.CacheRedis, gormPointer),
		NftModel:              nftModel,
		CollectionModel:       nft.NewL2NftCollectionModel(conn, c.CacheRedis, gormPointer),
		OfferModel:            offerModel,
		AssetModel:            assetModel,
		SysConfigModel:        sysConfig.NewSysConfigModel(conn, c.CacheRedis, gormPointer),

		PriceFetcher: price.NewFetcher(memCache, c.CoinMarketCap.Url, c.CoinMarketCap.Token),
		StateFetcher: state.NewFetcher(redisCache, accountModel, liquidityModel, nftModel),
	}
}
