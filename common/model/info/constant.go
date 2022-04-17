package info

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound      = sqlx.ErrNotFound
	ErrInvalidVolume = errors.New("[ErrInvalidVolume] invalid system volume")
	ErrInvalidTVL    = errors.New("[ErrInvalidVolume] invalid system tvl")
)
