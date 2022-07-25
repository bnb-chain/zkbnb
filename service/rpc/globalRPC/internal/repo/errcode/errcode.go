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
	ErrDataNotExist           = zerror.New(30000, "Data not exist")
	ErrSqlOperation           = zerror.New(30001, "Sql operation")
	ErrToFormatAccountInfo    = zerror.New(30002, "Err ToFormatAccountInfo")
	ErrFromFormatAccountInfo  = zerror.New(30003, "Err FromFormatAccountInfo")
	ErrComputeNewBalance      = zerror.New(30004, "Err ComputeNewBalance")
	ErrParseAccountAsset      = zerror.New(30005, "Err ParseAccountAsset")
	ErrParseInt               = zerror.New(30006, "Err ParseInt")
	ErrInvalidAssetType       = zerror.New(30007, "Invalid asset type")
	ErrConstructLiquidityInfo = zerror.New(30008, "Err ConstructLiquidityInfo")
	ErrParseLiquidityInfo     = zerror.New(30009, "Err ParseLiquidityInfo")
	ErrParseNftInfo           = zerror.New(30010, "Err ParseNftInfo")
	ErrInvalidFailTx          = zerror.New(30011, "Err invalid fail txVerification")
)
