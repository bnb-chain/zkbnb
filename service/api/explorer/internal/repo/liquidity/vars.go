package liquidity

import "github.com/bnb-chain/zkbas/pkg/zerror"

var (
	ErrNotExistInSql = zerror.New(40000, "not exist in sql ")
)
