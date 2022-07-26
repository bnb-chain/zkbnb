package txdetail

import (
	"context"

	table "github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type Model interface {
	GetTxsTotalCountByAccountIndex(ctx context.Context, accountIndex int64) (count int64, err error)
	GetTxDetailByAccountIndex(ctx context.Context, accountIndex int64) ([]*table.TxDetail, error)
	GetTxIdsByAccountIndex(ctx context.Context, accountIndex int64) ([]int64, error)
	GetDauInTxDetail(ctx context.Context, data string) (count int64, err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `tx_detail`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
