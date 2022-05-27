package mempool

import "github.com/zeromicro/go-zero/core/stores/sqlx"
import "errors"

var (
	ErrNotFound               = sqlx.ErrNotFound
	ErrInvalidMempoolTx       = errors.New("[ErrInvalidMempoolTx] invalid mempool tx")
	ErrInvalidMempoolTxDetail = errors.New("[ErrInvalidMempoolTxDetail] invalid mempool txDtail")
	MempoolTableName          = `mempool_tx`
)

const (
	PendingTxStatus = iota
	HandledTxStatus
)
