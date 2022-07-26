package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/account"
	asset "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/l1BlockMonitor"
	"github.com/bnb-chain/zkbas/common/model/l1TxSender"
	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/config"
)

type ServiceContext struct {
	NftModel              nft.L2NftModel
	BlockModel            block.BlockModel
	AccountModel          account.AccountModel
	MempoolModel          mempool.MempoolModel
	LiquidityModel        liquidity.LiquidityModel
	SysConfigModel        sysconfig.SysconfigModel
	L1TxSenderModel       l1TxSender.L1TxSenderModel
	L2AssetInfoModel      asset.AssetInfoModel
	L2TxEventMonitorModel l2TxEventMonitor.L2TxEventMonitorModel
	L1BlockMonitorModel   l1BlockMonitor.L1BlockMonitorModel

	RedisConnection *redis.Redis
	GormPointer     *gorm.DB
	Cache           multcache.MultCache
	Conn            sqlx.SqlConn
	Config          config.Config
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
		L2TxEventMonitorModel: l2TxEventMonitor.NewL2TxEventMonitorModel(conn, c.CacheRedis, gormPointer),
		AccountModel:          account.NewAccountModel(conn, c.CacheRedis, gormPointer),
		MempoolModel:          mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		LiquidityModel:        liquidity.NewLiquidityModel(conn, c.CacheRedis, gormPointer),
		NftModel:              nft.NewL2NftModel(conn, c.CacheRedis, gormPointer),
		BlockModel:            block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		L1TxSenderModel:       l1TxSender.NewL1TxSenderModel(conn, c.CacheRedis, gormPointer),
		L1BlockMonitorModel:   l1BlockMonitor.NewL1BlockMonitorModel(conn, c.CacheRedis, gormPointer),
		L2AssetInfoModel:      asset.NewAssetInfoModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:        sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
		RedisConnection:       redisConn,
		GormPointer:           gormPointer,
		Cache:                 multcache.NewGoCache(100, 10),
		Conn:                  conn,
		Config:                c,
	}
}
