package l2asset

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	L2AssetInfoTableName = `l2_asset_info`
)

var (
	ErrNotFound            = sqlx.ErrNotFound
	ErrInvalidL2AssetInput = errors.New("[ErrInvalidL2AssetInput] Invalid L2AssetInfo input")
)
