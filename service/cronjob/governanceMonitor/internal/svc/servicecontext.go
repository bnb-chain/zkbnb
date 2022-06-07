package svc

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/l1BlockMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2asset"
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/governanceMonitor/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config              config.Config
	L1BlockMonitorModel l1BlockMonitor.L1BlockMonitorModel
	L2AssetInfoModel    l2asset.L2AssetInfoModel
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
		L2AssetInfoModel:    l2asset.NewL2AssetInfoModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:      sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
	}
}
