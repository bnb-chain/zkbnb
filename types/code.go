package types

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// For internal errors, `Code` is not needed in current implementation.
// For external errors (app & glaobalRPC), we can define codes, however the current framework also
// does not use the codes. We can leave the codes for future enhancement.

var (
	DbErrNotFound                    = sqlx.ErrNotFound
	DbErrSqlOperation                = errors.New("unknown sql operation error")
	DbErrFailToCreateBlock           = errors.New("fail to create block")
	DbErrFailToUpdateBlock           = errors.New("fail to update block")
	DbErrFailToUpdateTx              = errors.New("fail to update tx")
	DbErrFailToCreateCompressedBlock = errors.New("fail to create compressed block")
	DbErrFailToCreateProof           = errors.New("fail to create proof")
	DbErrFailToUpdateProof           = errors.New("fail to update proof")
	DbErrFailToCreateSysConfig       = errors.New("fail to create system config")
	DbErrFailToUpdateSysConfig       = errors.New("fail to update system config")
	DbErrFailToCreateAsset           = errors.New("fail to create asset")
	DbErrFailToUpdateAsset           = errors.New("fail to update asset")
	DbErrFailToCreateAccount         = errors.New("fail to create account")
	DbErrFailToUpdateAccount         = errors.New("fail to update account")
	DbErrFailToCreateAccountHistory  = errors.New("fail to create account history")
	DbErrFailToCreateL1RollupTx      = errors.New("fail to create l1 rollup tx")
	DbErrFailToDeleteL1RollupTx      = errors.New("fail to delete l1 rollup tx")
	DbErrFailToL1SyncedBlock         = errors.New("fail to create l1 synced block")
	DbErrFailToCreatePoolTx          = errors.New("fail to create pool tx")
	DbErrFailToUpdatePoolTx          = errors.New("fail to update pool tx")
	DbErrFailToDeletePoolTx          = errors.New("fail to delete pool tx")
	DbErrFailToCreateNft             = errors.New("fail to create nft")
	DbErrFailToUpdateNft             = errors.New("fail to update nft")
	DbErrFailToCreateNftHistory      = errors.New("fail to create nft history")
	DbErrFailToCreatePriorityRequest = errors.New("fail to create priority request")
	DbErrFailToUpdatePriorityRequest = errors.New("fail to update priority request")

	JsonErrUnmarshal = errors.New("json.Unmarshal err")
	JsonErrMarshal   = errors.New("json.Marshal err")

	HttpErrFailToRequest = errors.New("http.NewRequest err")
	HttpErrClientDo      = errors.New("http.Client.Do err")

	IoErrFailToRead = errors.New("ioutil.ReadAll err")

	CmcNotListedErr = errors.New("cmc not listed")

	AppErrInvalidParam   = New(20001, "invalid param: ")
	AppErrInvalidTxField = New(20002, "invalid tx field: ")

	AppErrAccountNotFound      = New(21000, "account not found")
	AppErrAccountNonceNotFound = New(21001, "account nonce not found")
	AppErrInvalidAccountIndex  = New(21002, "invalid account index")
	AppErrAssetNotFound        = New(21100, "asset not found")
	AppErrInvalidAssetId       = New(21101, "invalid asset id")
	AppErrBlockNotFound        = New(21200, "block not found")
	AppErrInvalidBlockHeight   = New(21201, "invalid block height")
	AppErrPoolTxNotFound       = New(21300, "pool tx not found")

	AppErrTxInvalidExpireTime            = New(24000, "tx: invalid expired time")
	AppErrTxInvalidNonce                 = New(24001, "tx: invalid nonce")
	AppErrTxInvalidGasFeeAccount         = New(24002, "tx: invalid gas fee account")
	AppErrTxInvalidGasFeeAsset           = New(24003, "tx: invalid gas fee asset")
	AppErrTxInvalidGasFeeAmount          = New(24004, "tx: invalid gas fee amount")
	AppErrTxInvalidOfferType             = New(24005, "tx: invalid offer type")
	AppErrTxInvalidOfferState            = New(24005, "tx: invalid offer state, already canceled or finalized")
	AppErrTxBalanceNotEnough             = New(24006, "tx: balance is not enough")
	AppErrTxInvalidAssetAmount           = New(24007, "tx: invalid asset amount")
	AppErrTxNftAlreadyExist              = New(24008, "tx: invalid nft index, already exist")
	AppErrTxInvalidTxInfo                = New(24009, "tx: invalid tx info")
	AppErrTxInvalidNftContenthash        = New(24010, "tx: invalid nft content hash")
	AppErrTxInvalidCollectionId          = New(24011, "tx: invalid collection id")
	AppErrTxInvalidToAccountNameHash     = New(24012, "tx: invalid ToAccountNameHash")
	AppErrTxAccountNameAlreadyRegistered = New(24013, "tx: invalid account name, already registered")
	AppErrTxInvalidAccountIndex          = New(24014, "tx: invalid account index")
	AppErrTxNotNftOwner                  = New(24015, "tx: account is not owner of the nft")
	AppErrTxInvalidOfferId               = New(24016, "tx: invalid offer id")
	AppErrTxInvalidNftIndex              = New(24017, "tx: invalid nft index")
	AppErrTxInvalidAssetId               = New(24018, "tx: invalid asset id")
	AppErrTxInvalidListTime              = New(24019, "tx: invalid listed time")
	AppErrTxInvalidTreasuryRate          = New(24020, "tx: invalid treasury rate")
	AppErrTxInvalidCollectionName        = New(24021, "tx: invalid collection name")
	AppErrTxInvalidIntroduction          = New(24022, "tx: invalid introduction")
	AppErrTxInvalidPairIndex             = New(24023, "tx: invalid pair index")
	AppErrTxInvalidLpAmount              = New(24024, "tx: invalid lp amount")
	AppErrTxInvalidCallDataHash          = New(24025, "tx: invalid calldata hash")
	AppErrTxInvalidAssetMinAmount        = New(24026, "tx: invalid asset min amount")
	AppErrTxInvalidToAddress             = New(24027, "tx: invalid toAddress")
	AppErrTxInvalidBuyOffer              = New(24028, "tx: invalid buy offer")
	AppErrTxInvalidSellOffer             = New(24029, "tx: invalid sell offer")

	AppErrTxSameBuyerAndSeller         = New(24100, "tx: same buyer and seller")
	AppErrTxBuyOfferMismatchSellOffer  = New(24101, "tx: buy offer mismatches sell offer")
	AppErrTxInvalidBuyOfferExpireTime  = New(24102, "tx: invalid BuyOffer.ExpiredAt")
	AppErrTxInvalidSellOfferExpireTime = New(24103, "tx: invalid SellOffer.ExpiredAt")
	AppErrTxSellerBalanceNotEnough     = New(24104, "tx: seller balance is not enough")
	AppErrTxBuyerBalanceNotEnough      = New(24105, "tx: buyer balance is not enough")
	AppErrTxSellerNotOwner             = New(24106, "tx: seller is not owner")
	AppErrTxInvalidSellOfferState      = New(24107, "tx: invalid sell offer state, already canceled or finalized")
	AppErrTxInvalidBuyOfferState       = New(24108, "tx: invalid buy offer state, already canceled or finalized")

	AppErrInvalidGasAsset = New(25003, "invalid gas asset")
	AppErrInvalidTxType   = New(25004, "invalid tx type")
	AppErrTooManyTxs      = New(25005, "too many pending txs")
	AppErrNotFound        = New(29404, "not found")
	AppErrInternal        = New(29500, "internal server error")
)
