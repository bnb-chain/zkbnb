package errcode

import (
	"github.com/bnb-chain/zkbas/pkg/zerror"
)

var (
	// error code in [10000,20000) represent business error
	ErrExample1 = zerror.New(10000, "Example error msg")
	// error code in [20000,30000) represent code logic error
	ErrInvalidParam  = zerror.New(20000, "Invalid param")
	ErrInvalidTxType = zerror.New(20001, "txType error")
)
