package block

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
)

const (
	_ = iota
	StatusPending
	StatusCommitted
	StatusVerified
	StatusExecuted
)

var (
	ErrNotFound          = sqlx.ErrNotFound
	ErrDataNotExistInSQL = zerror.New(40000, "Err data not exist in SQL")
	ErrInvalidBlock      = zerror.New(40001, "[ErrInvalidBlock] invalid block")
	ErrInvalidMempoolTx  = zerror.New(40002, "[ErrInvalidBlock] invalid mempool tx")
)
