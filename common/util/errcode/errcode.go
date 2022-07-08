package errcode

import (
	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
)

var (
	// error code in [10000,20000) represent business error
	ErrInvalidAmount = zerror.New(10000, "Invalid Amount")
)
