package svc

import (
	"github.com/zecrey-labs/zecrey/common/model/account"
	"github.com/zecrey-labs/zecrey/common/model/asset"
	"github.com/zecrey-labs/zecrey/common/model/block"
	"github.com/zecrey-labs/zecrey/common/model/l1amount"
	"github.com/zecrey-labs/zecrey/common/model/l1asset"
	"github.com/zecrey-labs/zecrey/common/model/l2asset"
	"github.com/zecrey-labs/zecrey/common/model/liquidityPair"
	"github.com/zecrey-labs/zecrey/common/model/mempool"
	"github.com/zecrey-labs/zecrey/common/model/tx"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/config"
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
	KeyPairModel        account.KeyPairModel
	AssetModel          asset.AccountAssetModel
	LiquidityAssetModel asset.AccountLiquidityModel
	LockAssetModel      asset.AccountAssetLockModel
	TxModel             tx.TxModel
	TxDetailModel       tx.TxDetailModel
	FailTxModel         tx.FailTxModel
	LiquidityPairModel  liquidityPair.LiquidityPairModel
	BlockModel          block.BlockModel

	L1AssetModel l1asset.L1AssetInfoModel
	L2AssetModel l2asset.L2AssetInfoModel

	L1AmountModel l1amount.L1AmountModel

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
		KeyPairModel:        account.NewKeyPairModel(conn, c.CacheRedis, gormPointer),
		TxModel:             tx.NewTxModel(conn, c.CacheRedis, gormPointer, redisConn),
		TxDetailModel:       tx.NewTxDetailModel(conn, c.CacheRedis, gormPointer),
		FailTxModel:         tx.NewFailTxModel(conn, c.CacheRedis, gormPointer),
		LiquidityAssetModel: asset.NewAccountLiquidityModel(conn, c.CacheRedis, gormPointer),
		AssetModel:          asset.NewAccountAssetModel(conn, c.CacheRedis, gormPointer),
		LockAssetModel:      asset.NewAccountAssetLockModel(conn, c.CacheRedis, gormPointer),
		LiquidityPairModel:  liquidityPair.NewLiquidityPairModel(conn, c.CacheRedis, gormPointer),
		L1AssetModel:        l1asset.NewL1AssetInfoModel(conn, c.CacheRedis, gormPointer),
		L2AssetModel:        l2asset.NewL2AssetInfoModel(conn, c.CacheRedis, gormPointer),
		BlockModel:          block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		L1AmountModel:       l1amount.NewL1AmountModel(conn, c.CacheRedis, gormPointer),
	}
}
