package tx

import (
	"context"

	table "github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
)

type Model interface {
	GetTxsTotalCount(ctx context.Context) (count int64, err error)
	GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
	GetTxsByBlockId(blockId int64, limit, offset uint32) (txs []table.Tx, total int64, err error)
	GetTxByTxHash(txHash string) (tx *table.Tx, err error)
	GetTxs(limit, offset uint32) (txs []*table.Tx, err error)
	GetTxByTxID(txID int64) (tx *table.Tx, err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
