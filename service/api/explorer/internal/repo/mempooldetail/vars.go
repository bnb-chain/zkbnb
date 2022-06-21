package mempooldetail

import (
	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound               = sqlx.ErrNotFound
	ErrDataNotExist           = zerror.New(40000, "Data Not Exist")
	ErrInvalidMempoolTx       = zerror.New(40000, "[ErrInvalidMempoolTx] invalid mempool tx")
	ErrInvalidMempoolTxDetail = zerror.New(40001, "[ErrInvalidMempoolTxDetail] invalid mempool txDtail")
)

const (
	PendingTxStatus = iota
	HandledTxStatus
)
