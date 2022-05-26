package mempool

import "github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

type Mempool interface {
	GetMempoolTxs(offset uint64, limit uint64) (mempoolTx []*types.Tx, err error)
	GetMempoolTxsTotalCount() (count int64, err error)
	GetMempoolTxsListByAccountIndex(accountIndex int64, limit int64, offset int64) (mempoolTx []*types.Tx, err error)
	GetMempoolTxsTotalCountByPublicKey(pk string) (mempoolTx []*types.Tx, err error)
	GetMempoolTxByTxHash(hash string) (mempoolTxs *types.Tx, err error)
}
