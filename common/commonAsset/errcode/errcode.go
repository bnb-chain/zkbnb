package errcode

import (
	"github.com/bnb-chain/zkbas/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// err := error.New(10000, "Example error msg")
// fmt.Println("err:", err.Sprintf())
// error code in [10000,20000) represent business error
// error code in [20000,30000) represent logic layer error
// error code in [30000,40000) represent repo layer error
// error code in [40000,50000) represent common error

var (
	ErrNotFound = sqlx.ErrNotFound
)

var (
	ErrUnmarshal = zerror.New(50000, "Unmarshal err:")
)
