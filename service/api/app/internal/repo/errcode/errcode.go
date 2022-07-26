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

var (
	ErrNotFound = sqlx.ErrNotFound
)

var (
	ErrDataNotExist     = zerror.New(30000, "Data not exist")
	ErrSqlOperation     = zerror.New(30001, "Sql operation")
	ErrInvalidSysconfig = zerror.New(30002, "Invalid system config")
	ErrTypeAsset        = zerror.New(30003, "Err TypeAsset")
	ErrQuoteNotExist    = zerror.New(30004, "Err QuoteNotExist")
	ErrNewHttpRequest   = zerror.New(30005, "http.NewRequest err")
	ErrHttpClientDo     = zerror.New(30006, "http.Client.Do err")
	ErrIoutilReadAll    = zerror.New(30007, "ioutil.ReadAll err")
	ErrJsonUnmarshal    = zerror.New(30008, "json.Unmarshal err")
	ErrJsonMarshal      = zerror.New(30009, "json.Marshal err")
	ErrTypeAssertion    = zerror.New(30010, "type assertion err")
	ErrSetCache         = zerror.New(30011, "set cache err")
)
