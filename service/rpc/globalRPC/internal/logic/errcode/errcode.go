package errcode

import (
	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
)

var (
	// error code in [10000,20000) represent business error
	ErrExample1 = zerror.New(10000, "Example error msg")
	// error code in [20000,30000) represent code logic error
	ErrInvalidParam   = zerror.New(20000, "Invalid param")
	ErrInvalidTxType  = zerror.New(20001, "txType error")
	ErrInvalidGasFee  = zerror.New(20002, "Invalid Gas Fee")
	ErrInvalidAmount  = zerror.New(20003, "Invalid Amount")
	ErrInvalidAsset   = zerror.New(20004, "AssetA or AssetB is 0 in the liquidity Table")
	ErrInvalidAssetID = zerror.New(-1, "invalid pair assetId")
)
