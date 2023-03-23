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
	DbErrFailToL1Signature           = NewSystemError(10029, "fail to l1 signature")

	JsonErrUnmarshal = NewSystemError(10029, "json.Unmarshal err")
	JsonErrMarshal   = NewSystemError(10030, "json.Marshal err")

	HttpErrFailToRequest = NewSystemError(10031, "http.NewRequest err")
	HttpErrClientDo      = NewSystemError(10032, "http.Client.Do err")

	IoErrFailToRead = NewSystemError(10033, "ioutil.ReadAll err")

	CmcNotListedErr = NewSystemError(10034, "cmc not listed")

	TreeErrUnsupportedDriver = NewSystemError(11001, "unsupported db driver")

	AppErrInvalidParam                  = NewBusinessError(20001, "invalid param: ")
	AppErrInvalidTxField                = NewBusinessError(20002, "invalid tx field: ")
	AppErrDepositPubDataInvalidSize     = NewBusinessError(20003, "[ParseDepositPubData] invalid size")
	AppErrDepositNFTPubDataInvalidSize  = NewBusinessError(20004, "[ParseDepositNftPubData] invalid size")
	AppErrFullExitPubDataInvalidSize    = NewBusinessError(20005, "[ParseFullExitPubData] invalid size")
	AppErrFullExitNftPubDataInvalidSize = NewBusinessError(20006, "[ParseFullExitNftPubData] invalid size")

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
	AppErrAssetNotFound               = NewBusinessError(21200, "asset not found")
	AppErrInvalidAssetId              = NewBusinessError(21201, "invalid asset id")
	AppErrInvalidGasFeeAsset          = NewBusinessError(21202, "invalid gas fee asset")
	AppErrInvalidAssetAmount          = NewBusinessError(21203, "invalid asset amount")
	AppErrCompilationNotDeterministic = NewBusinessError(21204, "compilation is not deterministic")
	AppErrInvalidWitnessSolvedCS      = NewBusinessError(21205, "invalid witness solved the constraint system")
	AppErrInvalidWitnessVerified      = NewBusinessError(21206, "invalid witness resulted in a valid proof")
	AppErrInvalidAssetType            = NewBusinessError(21207, "invalid asset type")

	// Block
	AppErrBlockNotFound         = NewBusinessError(21300, "block not found")
	AppErrInvalidBlockHeight    = NewBusinessError(21301, "invalid block height")
	AppErrFailUpdateBlockStatus = NewBusinessError(21302, "update block status failed")

	// Tx
	AppErrPoolTxNotFound            = NewBusinessError(21400, "pool tx not found")
	AppErrInvalidTxInfo             = NewBusinessError(21401, "invalid tx info")
	AppErrMarshalTxFailed           = NewBusinessError(21402, "marshal tx failed")
	AppErrInsufficientGasFeeBalance = NewBusinessError(21403, "insufficient gas fee balance")
	AppErrInsufficientAssetBalance  = NewBusinessError(21404, "insufficient asset a balance")
	AppErrOfferIndexAlreadyInUse    = NewBusinessError(21405, "account offer index is already in use")
	AppErrBothOfferNotExist         = NewBusinessError(21406, "both buyOffer and sellOffer does not exist")
	AppErrTxSignatureError          = NewBusinessError(21407, "tx Signature Error")
	AppErrNoFetcherForTxType        = NewBusinessError(21408, "can not find fetcher function for tx type")
	AppErrNoSignFunctionForTxType   = NewBusinessError(21409, "can not find signature function for tx type")
	AppErrAccountNotNftOwner        = NewBusinessError(21410, "account is not owner of the nft")
	AppErrUnsupportedTxType         = NewBusinessError(21411, "unsupported tx type")
	AppErrPrepareNftFailed          = NewBusinessError(21412, "prepare nft failed")

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
	AppErrInvalidPlatformRate        = NewBusinessError(21514, "invalid platform rate")
	AppErrInvalidPlatformAmount      = NewBusinessError(21515, "invalid platform amount")
	AppErrInvalidSellPlatformAmount  = NewBusinessError(21516, "sell platform amount should be 0")
	AppErrInvalidSellPlatformRate    = NewBusinessError(21517, "sell platform rate should be 0")

	// Nft
	AppErrNftAlreadyExist          = NewBusinessError(21600, "invalid nft index, already exist")
	AppErrInvalidNftContenthash    = NewBusinessError(21601, "invalid nft content hash")
	AppErrNotNftOwner              = NewBusinessError(21602, "account is not owner of the nft")
	AppErrInvalidNftIndex          = NewBusinessError(21603, "invalid nft index")
	AppErrNftNotFound              = NewBusinessError(21604, "nft not found")
	AppErrInvalidToAccount         = NewBusinessError(21605, "invalid ToAccount")
	AppErrInvalidNft               = NewBusinessError(21606, "mutableAttributes is synchronizing")
	AppErrInvalidMetaData          = NewBusinessError(21607, "metaData should not be larger than ")
	AppErrInvalidMutableAttributes = NewBusinessError(21608, "mutableAttributes should not be larger than ")
	AppErrInvalidNftNonce          = NewBusinessError(21609, "invalid nft nonce")
	AppErrInvalidRoyaltyRate       = NewBusinessError(21610, "invalid royalty rate")
	// Collection
	AppErrInvalidCollectionId   = NewBusinessError(21700, "invalid collection id")
	AppErrInvalidCollectionName = NewBusinessError(21701, "invalid collection name")
	AppErrInvalidIntroduction   = NewBusinessError(21702, "invalid introduction")
	AppErrNotExistCollectionId  = NewBusinessError(21703, "collection id not exist")

	// Proof
	AppErrRelatedProofsNotReady = NewBusinessError(21800, "related proofs not ready")
	AppErrProofNumberNotMatch   = NewBusinessError(21801, "proof number not match")

	// Committer
	AppErrNilOptionalBlockSize = NewBusinessError(21802, "nil optional block sizes")

	// Witness
	AppErrInvalidBalanceString = NewBusinessError(21900, "invalid balance string")
	AppErrStateRootNotMatch    = NewBusinessError(21901, "state root doesn't match")

	// StateDB
	AppErrNotFindGasAccountConfig   = NewBusinessError(22000, "cannot find config for gas account index")
	AppErrInvalidGasAccountIndex    = NewBusinessError(22001, "invalid account index for gas account")
	AppErrFailUnmarshalGasFeeConfig = NewBusinessError(22002, "fail to unmarshal gas fee config")

	// RateLimit
	AppErrTooManyRequest         = NewBusinessError(23000, "Too Many Request!")
	AppErrUnknownRatelimitStatus = NewBusinessError(23001, "Unknown Rate Limit Status Error!")

	// PermissionControl
	AppErrForbiddenByBlackList    = NewBusinessError(23100, "Your account is blacklisted and not authorized,please contact our customer service!")
	AppErrNotPermittedByWhiteList = NewBusinessError(23101, "Your operation has been prohibited, please contact our customer service!")

	AppErrInvalidGasAsset     = NewBusinessError(25003, "invalid gas asset")
	AppErrInvalidTxType       = NewBusinessError(25004, "invalid tx type")
	AppErrInvalidAddress      = NewBusinessError(25005, "invalid address")
	AppErrInvalidSize         = NewBusinessError(25006, "invalid size")
	AppErrTooManyTxs          = NewBusinessError(25007, "too many pending txs")
	AppErrLockUsed            = NewBusinessError(29508, "the lock has been used, re-try later")
	AppErrCircuitMethodDefErr = NewBusinessError(29509, "frontend.Circuit methods must be defined on pointer receiver")
	AppErrNotFound            = NewBusinessError(29404, "not found")
	AppErrInternal            = NewBusinessError(29500, "internal server error")
)
