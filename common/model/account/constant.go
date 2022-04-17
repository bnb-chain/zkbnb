package account

import (
	"errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	AccountTableName        = `account`
	AccountHistoryTableName = `account_history`
)

const (
	AccountHistoryPending = iota
	AccountHistoryConfirmed
)

var (
	ErrNotFound               = sqlx.ErrNotFound
	ErrInvalidKeyPair         = errors.New("[ErrInvalidKeyPair] invalid key pair")
	ErrDuplicatedAccountName  = errors.New("duplicated account name, fatal error")
	ErrDuplicatedAccountIndex = errors.New("duplicated account index, fatal error")
)
