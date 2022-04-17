package tx

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	TxDetailTableName = `tx_detail`
	TxTableName       = `tx`
)

const (
	_ = iota
	StatusPending
	StatusSuccess
	StatusFail
)

const maxBlocks = 1000

var (
	ErrNotFound      = sqlx.ErrNotFound
	ErrInvalidFailTx = errors.New("[ErrInvalidTxFail] invalid fail tx")
)
