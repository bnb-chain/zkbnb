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

	AppErrInvalidExpireTime   = New(21000, "invalid expired time")
	AppErrInvalidGasFeeAmount = New(21001, "invalid gas fee amount")
	AppErrBalanceNotEnough    = New(21002, "balance is not enough")
	AppErrInvalidTreasuryRate = New(21003, "invalid treasury rate")
	AppErrInvalidCallDataHash = New(21004, "invalid calldata hash")
	AppErrInvalidToAddress    = New(21005, "invalid toAddress")

	// Account
	AppErrAccountNotFound              = New(21100, "account not found")
	AppErrAccountNonceNotFound         = New(21101, "account nonce not found")
	AppErrInvalidAccountIndex          = New(21102, "invalid account index")
	AppErrInvalidNonce                 = New(21103, "invalid nonce")
	AppErrInvalidGasFeeAccount         = New(21104, "invalid gas fee account")
	AppErrInvalidToAccountNameHash     = New(21105, "invalid ToAccountNameHash")
	AppErrAccountNameAlreadyRegistered = New(21106, "invalid account name, already registered")

	// Asset
	AppErrAssetNotFound      = New(21200, "asset not found")
	AppErrInvalidAssetId     = New(21201, "invalid asset id")
	AppErrInvalidGasFeeAsset = New(21202, "invalid gas fee asset")
	AppErrInvalidAssetAmount = New(21203, "invalid asset amount")

	// Block
	AppErrBlockNotFound      = New(21300, "block not found")
	AppErrInvalidBlockHeight = New(21301, "invalid block height")

	// Tx
	AppErrPoolTxNotFound = New(21400, "pool tx not found")
	AppErrInvalidTxInfo  = New(21401, "invalid tx info")

	// Offer
	AppErrInvalidOfferType           = New(21500, "invalid offer type")
	AppErrInvalidOfferState          = New(21501, "invalid offer state, already canceled or finalized")
	AppErrInvalidOfferId             = New(21502, "invalid offer id")
	AppErrInvalidListTime            = New(21503, "invalid listed time")
	AppErrInvalidBuyOffer            = New(21504, "invalid buy offer")
	AppErrInvalidSellOffer           = New(21505, "invalid sell offer")
	AppErrSameBuyerAndSeller         = New(21506, "same buyer and seller")
	AppErrBuyOfferMismatchSellOffer  = New(21506, "buy offer mismatches sell offer")
	AppErrInvalidBuyOfferExpireTime  = New(21507, "invalid BuyOffer.ExpiredAt")
	AppErrInvalidSellOfferExpireTime = New(21508, "invalid SellOffer.ExpiredAt")
	AppErrSellerBalanceNotEnough     = New(21508, "seller balance is not enough")
	AppErrBuyerBalanceNotEnough      = New(21509, "buyer balance is not enough")
	AppErrSellerNotOwner             = New(21510, "seller is not owner")
	AppErrInvalidSellOfferState      = New(21511, "invalid sell offer state, already canceled or finalized")
	AppErrInvalidBuyOfferState       = New(21512, "invalid buy offer state, already canceled or finalized")
	AppErrInvalidAssetOfOffer        = New(21513, "invalid asset of offer")

	// Nft
	AppErrNftAlreadyExist       = New(21600, "invalid nft index, already exist")
	AppErrInvalidNftContenthash = New(21601, "invalid nft content hash")
	AppErrNotNftOwner           = New(21602, "account is not owner of the nft")
	AppErrInvalidNftIndex       = New(21603, "invalid nft index")
	AppErrNftNotFound           = New(21604, "nft not found")
	AppErrInvalidToAccount      = New(21605, "invalid ToAccount")

	// Collection
	AppErrInvalidCollectionId   = New(21700, "invalid collection id")
	AppErrInvalidCollectionName = New(21701, "invalid collection name")
	AppErrInvalidIntroduction   = New(21702, "invalid introduction")

	AppErrInvalidGasAsset = New(25003, "invalid gas asset")
	AppErrInvalidTxType   = New(25004, "invalid tx type")
	AppErrTooManyTxs      = New(25005, "too many pending txs")
	AppErrNotFound        = New(29404, "not found")
	AppErrInternal        = New(29500, "internal server error")
)
