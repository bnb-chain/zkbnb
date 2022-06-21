package block

import (
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
	ErrNotFound = sqlx.ErrNotFound
)
