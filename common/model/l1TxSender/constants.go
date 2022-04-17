package l1TxSender

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var (
	ErrNotFound = sqlx.ErrNotFound
)

const (
	TableName = "l1_tx_sender"

	// status
	PendingStatus = 1
	HandledStatus = 2

	// tx type
	CommitTxType  = 1
	VerifyTxType  = 2
	ExecuteTxType = 3
	RevertTxType  = 4
)
