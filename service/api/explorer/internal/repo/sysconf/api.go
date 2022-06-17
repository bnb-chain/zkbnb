package sysconf

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
)

type Sysconf interface {
	GetSysconfigByName(name string) (info *table.Sysconfig, err error)
	CreateSysconfig(config *table.Sysconfig) error
	CreateSysconfigInBatches(configs []*table.Sysconfig) (rowsAffected int64, err error)
	UpdateSysconfig(config *table.Sysconfig) error
}

func New(svcCtx *svc.ServiceContext) Sysconf {
	return &sysconf{
		table: `sys_config`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
