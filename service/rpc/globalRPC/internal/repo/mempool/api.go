package mempool

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
)

type Mempool interface {
	GetMempoolTxs(offset int64, limit int64) (mempoolTx []*table.MempoolTx, err error)
	GetMempoolTxsTotalCount() (count int64, err error)
	GetMempoolTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
	GetMempoolTxsListByAccountIndex(accountIndex int64, limit int64, offset int64) (mempoolTxs []*table.MempoolTx, err error)

	//GetMempoolTxsTotalCountByPublicKey(pk string) (mempoolTx []*types.Tx, err error)
	GetMempoolTxByTxHash(hash string) (mempoolTxs *table.MempoolTx, err error)
	GetAccountAssetMempoolDetails(accountIndex int64, assetId int64, assetType int64) (mempoolTxDetails []*table.MempoolTxDetail, err error)
	GetMempoolTxsTotalCountByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8) (count int64, err error)
	GetMempoolTxsListByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8, limit int64, offset int64) (mempoolTxs []*table.MempoolTx, err error)
}

func New(svcCtx *svc.ServiceContext) Mempool {
	return &mempool{
		table: `mempool_tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
