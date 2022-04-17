package liquidityPair

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound                  = sqlx.ErrNotFound
	ErrInvalidLiquidityPairInput = errors.New("[ErrInvalidLiquidityPairInput] invalid liquidity pair input")
)
