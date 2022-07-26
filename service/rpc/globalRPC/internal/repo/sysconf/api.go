package sysconf

import (
	"context"

	table "github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
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
