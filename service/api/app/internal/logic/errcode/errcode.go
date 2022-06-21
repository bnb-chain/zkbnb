package errcode

import (
	"github.com/bnb-chain/zkbas/pkg/zerror"
)

var (
	// error code in [20000,30000) represent code logic error
	ErrInvalidParam = zerror.New(20000, "Invalid param")

	ErrNoLiquidity = zerror.New(20001, "no liquidity")
)
