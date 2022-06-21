package accounthistory

import (
	"github.com/bnb-chain/zkbas/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound = sqlx.ErrNotFound
)

var (
	ErrNotExistInSql = zerror.New(40000, "not exist in sql")
)
