package mempooldetail

import (
	mempoolModel "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
)

type Model interface {
	GetMempoolTxDetailByAccountIndex(accountIndex int64) ([]*mempoolModel.MempoolTxDetail, error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `mempool_tx_detail`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
