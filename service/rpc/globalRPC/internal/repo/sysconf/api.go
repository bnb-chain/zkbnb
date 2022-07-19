package sysconf

import (
	"context"

	table "github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
)

type Model interface {
	GetSysconfigByName(ctx context.Context, name string) (info *table.Sysconfig, err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `sys_config`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
