package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/l1RollupTx"

	"github.com/bnb-chain/zkbas/common/model/l1Block"
	"github.com/bnb-chain/zkbas/common/model/priorityRequest"

	asset "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/config"
)

type ServiceContext struct {
	BlockModel            block.BlockModel
	MempoolModel          mempool.MempoolModel
	SysConfigModel        sysconfig.SysconfigModel
	L1RollupTxModel       l1RollupTx.L1RollupTxModel
	L2AssetInfoModel      asset.AssetInfoModel
	L2TxEventMonitorModel priorityRequest.PriorityRequestModel
	L1BlockMonitorModel   l1Block.L1BlockModel
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
		logx.Errorf("gorm connect db error, err: %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	return &ServiceContext{
		L2TxEventMonitorModel: priorityRequest.NewPriorityRequestModel(conn, c.CacheRedis, gormPointer),
		MempoolModel:          mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		BlockModel:            block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		L1RollupTxModel:       l1RollupTx.NewL1RollupTxModel(conn, c.CacheRedis, gormPointer),
		L1BlockMonitorModel:   l1Block.NewL1BlockModel(conn, c.CacheRedis, gormPointer),
		L2AssetInfoModel:      asset.NewAssetInfoModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:        sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
	}
}
