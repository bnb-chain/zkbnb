package tx

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type Tx interface {
	GetTxsTotalCount(ctx context.Context) (count int64, err error)
	GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
}

func New(svcCtx *svc.ServiceContext) Tx {
	return &tx{
		table: `tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
