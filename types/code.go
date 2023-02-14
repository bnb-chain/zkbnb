package types

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// For internal errors, `Code` is not needed in current implementation.
// For external errors (app & globalRPC), we can define codes, however the current framework also
// does not use the codes. We can leave the codes for future enhancement.

var (
	DbErrNotFound                    = sqlx.ErrNotFound
	DbErrSqlOperation                = NewSystemError(10001, "unknown sql operation error")
	DbErrFailToCreateBlock           = NewSystemError(10002, "fail to create block")
	DbErrFailToUpdateBlock           = NewSystemError(10003, "fail to update block")
	DbErrFailToUpdateTx              = NewSystemError(10004, "fail to update tx")
	DbErrFailToCreateTx              = NewSystemError(10005, "fail to create tx")
	DbErrFailToCreateTxDetail        = NewSystemError(10006, "fail to create tx detail")
	DbErrFailToCreateCompressedBlock = NewSystemError(10007, "fail to create compressed block")
	DbErrFailToCreateProof           = NewSystemError(10008, "fail to create proof")
	DbErrFailToUpdateProof           = NewSystemError(10009, "fail to update proof")
	DbErrFailToCreateSysConfig       = NewSystemError(10010, "fail to create system config")
	DbErrFailToUpdateSysConfig       = NewSystemError(10011, "fail to update system config")
	DbErrFailToCreateAsset           = NewSystemError(10012, "fail to create asset")
	DbErrFailToUpdateAsset           = NewSystemError(10013, "fail to update asset")
	DbErrFailToCreateAccount         = NewSystemError(10014, "fail to create account")
	DbErrFailToUpdateAccount         = NewSystemError(10015, "fail to update account")
	DbErrFailToCreateAccountHistory  = NewSystemError(10016, "fail to create account history")
	DbErrFailToCreateL1RollupTx      = NewSystemError(10017, "fail to create l1 rollup tx")
	DbErrFailToDeleteL1RollupTx      = NewSystemError(10018, "fail to delete l1 rollup tx")
	DbErrFailToL1SyncedBlock         = NewSystemError(10019, "fail to create l1 synced block")
	DbErrFailToCreatePoolTx          = NewSystemError(10020, "fail to create pool tx")
	DbErrFailToUpdatePoolTx          = NewSystemError(10021, "fail to update pool tx")
	DbErrFailToDeletePoolTx          = NewSystemError(10022, "fail to delete pool tx")
	DbErrFailToCreateNft             = NewSystemError(10023, "fail to create nft")
	DbErrFailToUpdateNft             = NewSystemError(10024, "fail to update nft")
	DbErrFailToCreateNftHistory      = NewSystemError(10025, "fail to create nft history")
	DbErrFailToCreatePriorityRequest = NewSystemError(10026, "fail to create priority request")
	DbErrFailToUpdatePriorityRequest = NewSystemError(10027, "fail to update priority request")
	DbErrFailToCreateRollback        = NewSystemError(10028, "fail to create rollback")

	JsonErrUnmarshal = NewSystemError(10029, "json.Unmarshal err")
	JsonErrMarshal   = NewSystemError(10030, "json.Marshal err")

	HttpErrFailToRequest = NewSystemError(10031, "http.NewRequest err")
	HttpErrClientDo      = NewSystemError(10032, "http.Client.Do err")

	IoErrFailToRead = NewSystemError(10033, "ioutil.ReadAll err")

	CmcNotListedErr = NewSystemError(10034, "cmc not listed")

	AppErrInvalidParam   = NewBusinessError(20001, "invalid param: ")
	AppErrInvalidTxField = NewBusinessError(20002, "invalid tx field: ")

	AppErrInvalidExpireTime   = NewBusinessError(21000, "invalid expired time")
	AppErrInvalidGasFeeAmount = NewBusinessError(21001, "invalid gas fee amount")
	AppErrBalanceNotEnough    = NewBusinessError(21002, "balance is not enough")
	AppErrInvalidTreasuryRate = NewBusinessError(21003, "invalid treasury rate")
	AppErrInvalidCallDataHash = NewBusinessError(21004, "invalid calldata hash")
	AppErrInvalidToAddress    = NewBusinessError(21005, "invalid toAddress")

	// Account
	AppErrAccountNotFound              = NewBusinessError(21100, "account not found")
	AppErrAccountNonceNotFound         = NewBusinessError(21101, "account nonce not found")
	AppErrInvalidAccountIndex          = NewBusinessError(21102, "invalid account index")
	AppErrInvalidNonce                 = NewBusinessError(21103, "invalid nonce")
	AppErrInvalidGasFeeAccount         = NewBusinessError(21104, "invalid gas fee account")
	AppErrInvalidToAccountNameHash     = NewBusinessError(21105, "invalid ToAccountNameHash")
	AppErrAccountNameAlreadyRegistered = NewBusinessError(21106, "invalid account name, already registered")
	AppErrAccountInvalidToAccount      = NewBusinessError(21107, "invalid ToAccount")

	// Asset
	AppErrAssetNotFound      = NewBusinessError(21200, "asset not found")
	AppErrInvalidAssetId     = NewBusinessError(21201, "invalid asset id")
	AppErrInvalidGasFeeAsset = NewBusinessError(21202, "invalid gas fee asset")
	AppErrInvalidAssetAmount = NewBusinessError(21203, "invalid asset amount")

	// Block
	AppErrBlockNotFound      = NewBusinessError(21300, "block not found")
	AppErrInvalidBlockHeight = NewBusinessError(21301, "invalid block height")

	// Tx
	AppErrPoolTxNotFound = NewBusinessError(21400, "pool tx not found")
	AppErrInvalidTxInfo  = NewBusinessError(21401, "invalid tx info")

	// Offer
	AppErrInvalidOfferType           = NewBusinessError(21500, "invalid offer type")
	AppErrInvalidOfferState          = NewBusinessError(21501, "invalid offer state, already canceled or finalized")
	AppErrInvalidOfferId             = NewBusinessError(21502, "invalid offer id")
	AppErrInvalidListTime            = NewBusinessError(21503, "invalid listed time")
	AppErrInvalidBuyOffer            = NewBusinessError(21504, "invalid buy offer")
	AppErrInvalidSellOffer           = NewBusinessError(21505, "invalid sell offer")
	AppErrSameBuyerAndSeller         = NewBusinessError(21506, "same buyer and seller")
	AppErrBuyOfferMismatchSellOffer  = NewBusinessError(21506, "buy offer mismatches sell offer")
	AppErrInvalidBuyOfferExpireTime  = NewBusinessError(21507, "invalid BuyOffer.ExpiredAt")
	AppErrInvalidSellOfferExpireTime = NewBusinessError(21508, "invalid SellOffer.ExpiredAt")
	AppErrSellerBalanceNotEnough     = NewBusinessError(21508, "seller balance is not enough")
	AppErrBuyerBalanceNotEnough      = NewBusinessError(21509, "buyer balance is not enough")
	AppErrSellerNotOwner             = NewBusinessError(21510, "seller is not owner")
	AppErrInvalidSellOfferState      = NewBusinessError(21511, "invalid sell offer state, already canceled or finalized")
	AppErrInvalidBuyOfferState       = NewBusinessError(21512, "invalid buy offer state, already canceled or finalized")
	AppErrInvalidAssetOfOffer        = NewBusinessError(21513, "invalid asset of offer")

	// Nft
	AppErrNftAlreadyExist       = NewBusinessError(21600, "invalid nft index, already exist")
	AppErrInvalidNftContenthash = NewBusinessError(21601, "invalid nft content hash")
	AppErrNotNftOwner           = NewBusinessError(21602, "account is not owner of the nft")
	AppErrInvalidNftIndex       = NewBusinessError(21603, "invalid nft index")
	AppErrNftNotFound           = NewBusinessError(21604, "nft not found")
	AppErrInvalidToAccount      = NewBusinessError(21605, "invalid ToAccount")

	// Collection
	AppErrInvalidCollectionId   = NewBusinessError(21700, "invalid collection id")
	AppErrInvalidCollectionName = NewBusinessError(21701, "invalid collection name")
	AppErrInvalidIntroduction   = NewBusinessError(21702, "invalid introduction")

	AppErrInvalidGasAsset = NewBusinessError(25003, "invalid gas asset")
	AppErrInvalidTxType   = NewBusinessError(25004, "invalid tx type")
	AppErrTooManyTxs      = NewBusinessError(25005, "too many pending txs")
	AppErrNotFound        = NewBusinessError(29404, "not found")
	AppErrInternal        = NewBusinessError(29500, "internal server error")
)
