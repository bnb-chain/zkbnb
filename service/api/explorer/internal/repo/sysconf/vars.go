package sysconf

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound         = sqlx.ErrNotFound
	ErrInvalidSysconfig = errors.New("[ErrInvalidSysconfig] invalid system config")

	NameColumn      = "name"
	ValueColumn     = "value"
	ValueTypeColumn = "value_type"
	CommentColumn   = "comment"
)
