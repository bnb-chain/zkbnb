package mempool

import (
	"errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	MempoolTableName = `mempool_tx`
	DetailTableName  = `mempool_tx_detail`
)

var (
	ErrNotFound               = sqlx.ErrNotFound
	ErrInvalidMempoolTx       = errors.New("[ErrInvalidMempoolTx] invalid mempool tx")
	ErrInvalidMempoolTxDetail = errors.New("[ErrInvalidMempoolTxDetail] invalid mempool txDtail")
)

const (
	PendingTxStatus = iota
	HandledTxStatus
)
