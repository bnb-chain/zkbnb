package txdetail

import (
	"context"

	table "github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
)

type Model interface {
	GetTxDetailByAccountIndex(ctx context.Context, accountIndex int64) ([]*table.TxDetail, error)
	GetDauInTxDetail(ctx context.Context, data string) (count int64, err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `tx_detail`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
