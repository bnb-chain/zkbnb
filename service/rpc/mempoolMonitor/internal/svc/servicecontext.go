package svc

import (
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/account"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/asset"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/l2asset"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/mempoolMonitor/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config                config.Config
	L2TxEventMonitorModel l2TxEventMonitor.L2TxEventMonitorModel
	L2assetInfoModel      l2asset.L2AssetInfoModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	AccountAssetModel     asset.AccountAssetModel
	MempoolModel          mempool.MempoolModel
	DbEngine              *gorm.DB
	//GlobalRPC             globalrpc.GlobalRPC
}

func NewServiceContext(c config.Config) *ServiceContext {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)

	return &ServiceContext{
		Config: c,

		L2TxEventMonitorModel: l2TxEventMonitor.NewL2TxEventMonitorModel(conn, c.CacheRedis, gormPointer),
		L2assetInfoModel:      l2asset.NewL2AssetInfoModel(conn, c.CacheRedis, gormPointer),
		AccountModel:          account.NewAccountModel(conn, c.CacheRedis, gormPointer),
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
		AccountAssetModel:     asset.NewAccountAssetModel(conn, c.CacheRedis, gormPointer),
		MempoolModel:          mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),

		//GlobalRPC: globalrpc.NewGlobalRPC(zrpc.MustNewClient(c.GlobalRpc)),
	}
}
