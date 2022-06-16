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
	StatusExecuted
)

var (
	ErrNotFound         = sqlx.ErrNotFound
	ErrInvalidBlock     = errors.New("[ErrInvalidBlock] invalid block")
	ErrInvalidMempoolTx = errors.New("[ErrInvalidBlock] invalid mempool tx")
)

const (
	DetailTableName = `block_detail`

	BlockStatusColumn = "block_status"
	CommittedAtColumn = "committed_at"
	VerifiedAtColumn  = "verified_at"
	ExecutedAtColumn  = "executed_at"
)
