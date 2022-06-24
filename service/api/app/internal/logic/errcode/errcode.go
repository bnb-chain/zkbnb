package errcode

import (
	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
)

// error Custom error type in zecrey
// using method:
// err := error.New(10000, "Example error msg")
// fmt.Println("err:", err.Sprintf())
// error code in [10000,20000) represent business error
// error code in [20000,30000) represent logic layer error
// error code in [30000,40000) represent repo layer error

var (
	ErrInvalidParam = zerror.New(20000, "Invalid param")
)
