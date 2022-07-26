package mempooltxdetail

import (
	"context"

	table "github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type Model interface {
	GetMemPoolTxDetailByAccountIndex(ctx context.Context, accountIndex int64) (mempoolTx []*table.MempoolTxDetail, err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `mempool_tx_detail`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
