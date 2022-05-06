package svc

import (
	"github.com/zecrey-labs/zecrey-core/common/general/model/liquidityPair"
	"github.com/zecrey-labs/zecrey-core/common/general/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/asset"
	"github.com/zecrey-labs/zecrey-legend/common/model/assetHistory"
	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2asset"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config              config.Config
	MempoolModel        mempool.MempoolModel
	MempoolDetailModel  mempool.MempoolTxDetailModel
	AccountModel        account.AccountModel
	AccountHistoryModel account.AccountHistoryModel
	AssetModel          asset.AccountAssetModel
	AssetHistoryModel   assetHistory.AccountAssetHistoryModel
	LiquidityAssetModel asset.AccountLiquidityModel
	TxModel             tx.TxModel
	TxDetailModel       tx.TxDetailModel
	FailTxModel         tx.FailTxModel
	LiquidityPairModel  liquidityPair.LiquidityPairModel
	BlockModel          block.BlockModel

	L2AssetModel l2asset.L2AssetInfoModel

	SysConfigModel sysconfig.SysconfigModel

	RedisConnection *redis.Redis

	DbEngine *gorm.DB
}

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func NewServiceContext(c config.Config) *ServiceContext {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	return &ServiceContext{
		Config:              c,
		MempoolModel:        mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		MempoolDetailModel:  mempool.NewMempoolDetailModel(conn, c.CacheRedis, gormPointer),
		AccountModel:        account.NewAccountModel(conn, c.CacheRedis, gormPointer),
		AccountHistoryModel: account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
		AssetModel:          asset.NewAccountAssetModel(conn, c.CacheRedis, gormPointer),
		AssetHistoryModel:   assetHistory.NewAccountAssetHistoryModel(conn, c.CacheRedis, gormPointer),
		LiquidityAssetModel: asset.NewAccountLiquidityModel(conn, c.CacheRedis, gormPointer),
		TxModel:             tx.NewTxModel(conn, c.CacheRedis, gormPointer, redisConn),
		TxDetailModel:       tx.NewTxDetailModel(conn, c.CacheRedis, gormPointer),
		FailTxModel:         tx.NewFailTxModel(conn, c.CacheRedis, gormPointer),
		LiquidityPairModel:  liquidityPair.NewLiquidityPairModel(conn, c.CacheRedis, gormPointer),
		BlockModel:          block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		L2AssetModel:        l2asset.NewL2AssetInfoModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:      sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
		RedisConnection:     redisConn,
	}
}
