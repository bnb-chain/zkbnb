package mempool

import (
	"github.com/bnb-chain/zkbas/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound               = sqlx.ErrNotFound
	ErrInvalidMempoolTx       = zerror.New(30000, "[ErrInvalidMempoolTx] invalid mempool tx")
	ErrInvalidMempoolTxDetail = zerror.New(30001, "[ErrInvalidMempoolTxDetail] invalid mempool txDtail")
	MempoolTableName          = `mempool_tx`
)

const (
	PendingTxStatus = iota
	HandledTxStatus
)

var (
	ErrNotExistInSql = zerror.New(40000, "not exist in sql ")
	ErrIllegalParam  = zerror.New(40001, "illegal param ")
)
