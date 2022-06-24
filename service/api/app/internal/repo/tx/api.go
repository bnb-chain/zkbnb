package tx

import (
	"context"

	table "github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
)

type Model interface {
	GetTxsTotalCount(ctx context.Context) (count int64, err error)
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
