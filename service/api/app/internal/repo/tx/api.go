package tx

import (
	"context"

	table "github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type Model interface {
	GetTxsTotalCount(ctx context.Context) (count int64, err error)
	GetTxsList(ctx context.Context, limit int64, offset int64) (blocks []*table.Tx, err error)
	GetTxByTxHash(ctx context.Context, txHash string) (tx *table.Tx, err error)
	GetTxByTxID(ctx context.Context, txID int64) (tx *table.Tx, err error)
	GetTxCountByTimeRange(ctx context.Context, data string) (count int64, err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
