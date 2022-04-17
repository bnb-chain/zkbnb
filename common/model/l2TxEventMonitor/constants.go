package l2TxEventMonitor

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var (
	ErrNotFound = sqlx.ErrNotFound
)

const (
	TableName = "l2_tx_event_monitor"
)
