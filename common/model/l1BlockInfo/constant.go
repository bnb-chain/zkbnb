package l1BlockInfo

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound = sqlx.ErrNotFound
)

const (
	TableName = "l1_block_info"
)
