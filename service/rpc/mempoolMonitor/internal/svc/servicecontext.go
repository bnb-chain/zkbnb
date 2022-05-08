package svc

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/asset"
	"github.com/zecrey-labs/zecrey-legend/common/model/assetHistory"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2asset"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/mempoolMonitor/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config                   config.Config
	L2TxEventMonitorModel    l2TxEventMonitor.L2TxEventMonitorModel
	L2assetInfoModel         l2asset.L2AssetInfoModel
	AccountModel             account.AccountModel
	AccountHistoryModel      account.AccountHistoryModel
	AccountAssetModel        asset.AccountAssetModel
	AccountAssetHistoryModel assetHistory.AccountAssetHistoryModel
	MempoolModel             mempool.MempoolModel
	MempoolTxDetailModel     mempool.MempoolTxDetailModel
	RedisConnection          *redis.Redis
	DbEngine                 *gorm.DB
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
		Config:                   c,
		L2TxEventMonitorModel:    l2TxEventMonitor.NewL2TxEventMonitorModel(conn, c.CacheRedis, gormPointer),
		L2assetInfoModel:         l2asset.NewL2AssetInfoModel(conn, c.CacheRedis, gormPointer),
		AccountModel:             account.NewAccountModel(conn, c.CacheRedis, gormPointer),
		AccountHistoryModel:      account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
		AccountAssetModel:        asset.NewAccountAssetModel(conn, c.CacheRedis, gormPointer),
		AccountAssetHistoryModel: assetHistory.NewAccountAssetHistoryModel(conn, c.CacheRedis, gormPointer),
		MempoolModel:             mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		MempoolTxDetailModel:     mempool.NewMempoolDetailModel(conn, c.CacheRedis, gormPointer),
		RedisConnection:          redisConn,
	}
}
