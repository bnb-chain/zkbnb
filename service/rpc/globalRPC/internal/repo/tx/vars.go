package tx

import "github.com/zecrey-labs/zecrey-legend/pkg/zerror"

var (
	ErrNotExistInSql = zerror.New(40000, "not exist in sql ")
	ErrIllegalParam  = zerror.New(40001, "illegal param ")
)