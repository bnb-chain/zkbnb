package l2BlockEventMonitor

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var (
	ErrNotFound = sqlx.ErrNotFound
)

const (
	TableName = "l2_block_event_monitor"

	// status
	PendingStatus = 1
	HandledStatus = 2

	// block event type
	CommittedBlockEventType = 1
	VerifiedBlockEventType  = 2
	ExecutedBlockEventType  = 3
	RevertedBlockEventType  = 4
)
