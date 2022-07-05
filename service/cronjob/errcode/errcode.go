package errcode

import (
	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound = sqlx.ErrNotFound
)

var (
	// error code in [10000,20000) represent business error
	ErrExample1 = zerror.New(10000, "Example error msg")
	// error code in [20000,30000) represent code logic error
	ErrNotEnoughTransactions     = zerror.New(20000, "[CommitterTask] not enough transactions")
	ErrNotInvalidCollectionNonce = zerror.New(20001, "[CommitterTask] invalid collection nonce")
)
