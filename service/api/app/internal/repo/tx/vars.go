package tx

import (
	"github.com/bnb-chain/zkbas/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	Fail = iota
	NoRegisteredNoActive
	RegisteredNoActive
	Active
)

const (
	AccountHistoryPending = iota
	AccountHistoryConfirmed
)

var (
	ErrNotFound      = sqlx.ErrNotFound
	ErrNotExistInSql = zerror.New(40000, "not exist in sql ")
	ErrIllegalParam  = zerror.New(40001, "illegal param ")
)
