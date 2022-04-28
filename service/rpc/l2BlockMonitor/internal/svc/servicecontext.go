package svc

import (
	"github.com/zecrey-labs/zecrey-core/common/general/model/l1TxSender"
	"github.com/zecrey-labs/zecrey-core/common/general/model/l2BlockEventMonitor"
	"github.com/zecrey-labs/zecrey-core/common/general/model/nft"
	"github.com/zecrey-labs/zecrey-core/common/general/model/sysconfig"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/asset"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/assetHistory"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/block"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config                       config.Config
	Mempool                      mempool.MempoolModel
	Block                        block.BlockModel
	L2BlockEventMonitor          l2BlockEventMonitor.L2BlockEventMonitorModel
	L1TxSender                   l1TxSender.L1TxSenderModel
	SysConfig                    sysconfig.SysconfigModel
	AccountAssetModel            asset.AccountAssetModel
	AccountLiquidityModel        asset.AccountLiquidityModel
	NftModel                     nft.L2NftModel
	AccountAssetHistoryModel     assetHistory.AccountAssetHistoryModel
	AccountLiquidityHistoryModel assetHistory.AccountLiquidityHistoryModel
	NftHistoryModel              nft.L2NftHistoryModel
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
		Config:                       c,
		Mempool:                      mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		Block:                        block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		L2BlockEventMonitor:          l2BlockEventMonitor.NewL2BlockEventMonitorModel(conn, c.CacheRedis, gormPointer),
		L1TxSender:                   l1TxSender.NewL1TxSenderModel(conn, c.CacheRedis, gormPointer),
		SysConfig:                    sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
		AccountAssetModel:            asset.NewAccountAssetModel(conn, c.CacheRedis, gormPointer),
		AccountLiquidityModel:        asset.NewAccountLiquidityModel(conn, c.CacheRedis, gormPointer),
		AccountAssetHistoryModel:     assetHistory.NewAccountAssetHistoryModel(conn, c.CacheRedis, gormPointer),
		AccountLiquidityHistoryModel: assetHistory.NewAccountLiquidityHistoryModel(conn, c.CacheRedis, gormPointer),
	}
}
