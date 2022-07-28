package l2asset

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound            = sqlx.ErrNotFound
	ErrInvalidL2AssetInput = errors.New("[ErrInvalidL2AssetInput] Invalid L2AssetInfo input")
)
