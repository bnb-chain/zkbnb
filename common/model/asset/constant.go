package asset

import (
	"errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	GeneralAssetTable   = `account_asset`
	LiquidityAssetTable = `account_liquidity`
)

var (
	ErrNotFound                     = sqlx.ErrNotFound
	ErrAccountExist                 = errors.New("[ErrAccountExist] Account already exists")
	ErrInvalidAccountAssetInput     = errors.New("[ErrInvalidAccountAssetInput] Invalid accountAsset input")
	ErrInvalidAccountAssetLockInput = errors.New("[ErrInvalidAccountAssetLockInput]Invalid AccountAssetLock input")
	ErrInvalidAccountLiquidityInput = errors.New("[ErrInvalidAccountLiquidityInput] invalid account liquidity input")
)
