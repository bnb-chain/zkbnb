package l2asset

import (
	"errors"

	"github.com/bnb-chain/zkbas/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound            = sqlx.ErrNotFound
	ErrInvalidL2AssetInput = errors.New("[ErrInvalidL2AssetInput] Invalid L2AssetInfo input")
)

var (
	ErrNotExistInSql = zerror.New(40000, "not exist in sql")
)
