package tx

import "github.com/bnb-chain/zkbas/pkg/zerror"

var (
	ErrNotExistInSql = zerror.New(40000, "not exist in sql ")
	ErrIllegalParam  = zerror.New(40001, "illegal param ")
)
