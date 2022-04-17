package block

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	_ = iota
	StatusPending
	StatusCommitted
	StatusVerified
)

var (
	ErrNotFound         = sqlx.ErrNotFound
	ErrInvalidBlock     = errors.New("[ErrInvalidBlock] invalid block")
	ErrInvalidMempoolTx = errors.New("[ErrInvalidBlock] invalid mempool tx")
)

const (
	BlockTableName = `block`
)
