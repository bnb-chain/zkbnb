package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	asset "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/l1BlockMonitor"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/service/cronjob/governanceMonitor/internal/config"
)

type ServiceContext struct {
	Config              config.Config
	L1BlockMonitorModel l1BlockMonitor.L1BlockMonitorModel
	L2AssetInfoModel    asset.AssetInfoModel
	SysConfigModel      sysconfig.SysconfigModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)

	return &ServiceContext{
		Config:              c,
		L1BlockMonitorModel: l1BlockMonitor.NewL1BlockMonitorModel(conn, c.CacheRedis, gormPointer),
		L2AssetInfoModel:    asset.NewAssetInfoModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:      sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
	}
}
