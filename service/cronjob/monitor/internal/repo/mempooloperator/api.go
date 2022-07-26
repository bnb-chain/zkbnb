package mempooloperator

import (
	table "github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/svc"
)

type Model interface {
	CreateMempoolTxs(pendingNewMempoolTxs []*table.MempoolTx) (err error)
	DeleteMempoolTxs(pendingUpdateMempoolTxs []*table.MempoolTx) (err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `mempool_tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
