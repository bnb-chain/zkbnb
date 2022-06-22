package txdetail

import (
	"context"

	table "github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
)

type Model interface {
	GetTxDetailByAccountIndex(ctx context.Context, accountIndex int64) ([]*table.TxDetail, error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `tx_detail`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
