package l1amount

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound        = sqlx.ErrNotFound
	ErrInvalidL1Amount = errors.New("[ErrInvalidAmount] invalid system l1amount")
)

const (
	TableName = `l1_amount`

	TotalAmountColumn = "total_amount"
)
