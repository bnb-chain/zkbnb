package errorcode

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// For internal errors, `Code` is not needed in current implementation.
// For external errors (app & glaobalRPC), we can define codes, however the current framework also
// does not use the codes. We can leave the codes for future enhancement.

var (
	DbErrNotFound                  = sqlx.ErrNotFound
	DbErrSqlOperation              = errors.New("unknown sql operation error")
	DbErrDuplicatedAccountName     = errors.New("duplicated account name")
	DbErrDuplicatedAccountIndex    = errors.New("duplicated account index")
	DbErrDuplicatedCollectionIndex = errors.New("duplicated collection index")
	DbErrFailToCreateBlock         = errors.New("fail to create block")
	DbErrFailToCreateAssetInfo     = errors.New("fail to create asset info")
	DbErrFailToCreateVolume        = errors.New("fail to create volume")
	DbErrFailToCreateTVL           = errors.New("fail to create tvl")
	DbErrFailToCreateLiquidity     = errors.New("fail to create liquidity")
	DbErrFailToCreateMempoolTx     = errors.New("fail to create mempool tx")
	DbErrFailToCreateProof         = errors.New("fail to create proof")
	DbErrFailToCreateFailTx        = errors.New("fail to create fail tx")
	DbErrFailToCreateSysconfig     = errors.New("fail to create system config")

	JsonErrUnmarshal = errors.New("json.Unmarshal err")
	JsonErrMarshal   = errors.New("json.Marshal err")

	HttpErrFailToRequest = errors.New("http.NewRequest err")
	HttpErrClientDo      = errors.New("http.Client.Do err")

	IoErrFailToRead = errors.New("ioutil.ReadAll err")

	//TODO: more error code, parameter check, transaction check

	//global rpc

	RpcErrInvalidTx      = New(20001, "invalid tx: cannot parse tx")
	RpcErrInvalidTxType  = New(20002, "invalid tx type ")
	RpcErrInvalidTxField = New(20003, "invalid tx field: ")
	RpcErrVerification   = New(20004, "tx verification failed: ")

	RpcErrInvalidParam                = New(20005, "invalid param: ")
	RpcErrLiquidityInvalidAssetAmount = New(20006, "invalid liquidity asset amount")
	RpcErrLiquidityInvalidAssetID     = New(20007, "invalid liquidity asset id")

	RpcErrNotFound = New(24404, "not found")
	RpcErrInternal = New(24500, "internal server error")

	//app service

	AppErrInvalidParam    = New(25001, "invalid param:")
	AppErrQuoteNotExist   = New(25002, "quote asset does not exist")
	AppErrInvalidGasAsset = New(25003, "invalid gas asset")

	AppErrNotFound = New(29404, "not found")
	AppErrInternal = New(29500, "internal server error")
)
