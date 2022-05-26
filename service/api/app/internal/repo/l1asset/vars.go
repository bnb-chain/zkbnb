package l1asset

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound             = sqlx.ErrNotFound
	ErrUpdateNoRowsAffected = errors.New("[ErrUpdateNoRowsAffected] Update no rows affected")
	ErrAssetInvalid         = errors.New("[ErrInvalidAsset] Invalid asset input")
	ErrL1AssetExist         = errors.New("[ErrL1AssetExist] L1Asset already exists")
)
