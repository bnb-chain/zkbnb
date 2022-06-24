package sysconf

import (
	"context"
	table "github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
)

type Sysconf interface {
	GetSysconfigByName(ctx context.Context, name string) (info *table.Sysconfig, err error)
	CreateSysconfig(ctx context.Context, config *table.Sysconfig) error
	CreateSysconfigInBatches(ctx context.Context, configs []*table.Sysconfig) (rowsAffected int64, err error)
	UpdateSysconfig(ctx context.Context, config *table.Sysconfig) error
}

func New(svcCtx *svc.ServiceContext) Sysconf {
	return &sysconf{
		table: `sys_config`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
