package liquidityoperator

import "github.com/bnb-chain/zkbas/pkg/zerror"

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
	ErrNotExistInSql = zerror.New(40000, "not exist in sql ")
	ErrIllegalParam  = zerror.New(40001, "illegal param ")
)
