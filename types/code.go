package types

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// For internal errors, `Code` is not needed in current implementation.
// For external errors (app & glaobalRPC), we can define codes, however the current framework also
// does not use the codes. We can leave the codes for future enhancement.

var (
	DbErrNotFound                     = sqlx.ErrNotFound
	DbErrSqlOperation                 = errors.New("unknown sql operation error")
	DbErrFailToCreateBlock            = errors.New("fail to create block")
	DbErrFailToUpdateBlock            = errors.New("fail to update block")
	DbErrFailToCreateCompressedBlock  = errors.New("fail to create compressed block")
	DbErrFailToCreateProof            = errors.New("fail to create proof")
	DbErrFailToUpdateProof            = errors.New("fail to update proof")
	DbErrFailToCreateFailTx           = errors.New("fail to create fail tx")
	DbErrFailToCreateSysConfig        = errors.New("fail to create system config")
	DbErrFailToUpdateSysConfig        = errors.New("fail to update system config")
	DbErrFailToCreateAsset            = errors.New("fail to create asset")
	DbErrFailToUpdateAsset            = errors.New("fail to update asset")
	DbErrFailToCreateAccount          = errors.New("fail to create account")
	DbErrFailToUpdateAccount          = errors.New("fail to update account")
	DbErrFailToCreateAccountHistory   = errors.New("fail to create account history")
	DbErrFailToCreateL1RollupTx       = errors.New("fail to create l1 rollup tx")
	DbErrFailToDeleteL1RollupTx       = errors.New("fail to delete l1 rollup tx")
	DbErrFailToL1SyncedBlock          = errors.New("fail to create l1 synced block")
	DbErrFailToCreateLiquidity        = errors.New("fail to create liquidity")
	DbErrFailToUpdateLiquidity        = errors.New("fail to update liquidity")
	DbErrFailToCreateLiquidityHistory = errors.New("fail to create liquidity history")
	DbErrFailToCreatePoolTx           = errors.New("fail to create pool tx")
	DbErrFailToUpdatePoolTx           = errors.New("fail to update pool tx")
	DbErrFailToDeletePoolTx           = errors.New("fail to delete pool tx")
	DbErrFailToCreateNft              = errors.New("fail to create nft")
	DbErrFailToUpdateNft              = errors.New("fail to update nft")
	DbErrFailToCreateNftHistory       = errors.New("fail to create nft history")
	DbErrFailToCreatePriorityRequest  = errors.New("fail to create priority request")
	DbErrFailToUpdatePriorityRequest  = errors.New("fail to update priority request")

	JsonErrUnmarshal = errors.New("json.Unmarshal err")
	JsonErrMarshal   = errors.New("json.Marshal err")

	HttpErrFailToRequest = errors.New("http.NewRequest err")
	HttpErrClientDo      = errors.New("http.Client.Do err")

	IoErrFailToRead = errors.New("ioutil.ReadAll err")

	CmcNotListedErr = errors.New("cmc not listed")

	AppErrInvalidParam    = New(20001, "invalid param: ")
	AppErrInvalidTx       = New(20002, "invalid tx: cannot parse tx")
	AppErrInvalidTxType   = New(20003, "invalid tx type")
	AppErrInvalidTxField  = New(20004, "invalid tx field: ")
	AppErrInvalidGasAsset = New(25005, "invalid gas asset")
	AppErrNotFound        = New(29404, "not found")
	AppErrInternal        = New(29500, "internal server error")
)
