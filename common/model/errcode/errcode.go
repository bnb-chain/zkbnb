package errcode

import (
	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// err := error.New(10000, "Example error msg")
// fmt.Println("err:", err.Sprintf())
// error code in [10000,20000) represent business error
// error code in [20000,30000) represent logic layer error
// error code in [30000,40000) represent repo layer error

var (
	ErrNotFound = sqlx.ErrNotFound
)

var (
	ErrDataNotExist = zerror.New(40000, "Data not exist")
	ErrSQLErr       = zerror.New(40001, "SQL err:")
)
