package sysConfig

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound         = sqlx.ErrNotFound
	ErrInvalidSysconfig = errors.New("[ErrInvalidSysconfig] invalid system config")
	MaxChainId          = "Max_Chain_Id"
	MaxAssetId          = "Max_Asset_Id"

	NameColumn      = "name"
	ValueColumn     = "value"
	ValueTypeColumn = "value_type"
	CommentColumn   = "comment"
)
