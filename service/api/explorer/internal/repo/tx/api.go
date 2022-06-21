package tx

import (
	"context"

	table "github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
)

type Tx interface {
	GetTxsTotalCount(ctx context.Context) (count int64, err error)
	GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
	GetTxsByBlockId(blockId int64, limit, offset uint32) (txs []table.Tx, total int64, err error)
	GetTxByTxHash(txHash string) (tx *table.Tx, err error)
}

func New(svcCtx *svc.ServiceContext) Tx {
	return &tx{
		table: `tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
