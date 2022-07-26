package failtx

import (
	table "github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

//go:generate mockgen -source api.go -destination api_mock.go -package failtx

type Model interface {
	CreateFailTx(failTx *table.FailTx) error
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `fail_tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
