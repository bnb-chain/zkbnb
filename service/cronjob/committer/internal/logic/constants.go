package logic

import (
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/assetInfo"
	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/model/blockForCommit"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"math/big"
)

var (
	ErrNotFound = sqlx.ErrNotFound
)

type (
	// tx
	Tx       = tx.Tx
	TxDetail = tx.TxDetail
	// block
	Block          = block.Block
	BlockForCommit = blockForCommit.BlockForCommit
	// mempool
	MempoolTx       = mempool.MempoolTx
	MempoolTxDetail = mempool.MempoolTxDetail
	// assets
	L2Nft = nft.L2Nft
	// assets history
	L2NftHistory = nft.L2NftHistory
	// account history
	Account        = account.Account
	AccountHistory = account.AccountHistory

	FormatAccountInfo        = commonAsset.AccountInfo
	FormatAccountHistoryInfo = commonAsset.FormatAccountHistoryInfo

	Liquidity        = liquidity.Liquidity
	LiquidityHistory = liquidity.LiquidityHistory

	SysconfigModel = sysconfig.SysconfigModel
	MempoolModel   = mempool.MempoolModel
	BlockModel     = block.BlockModel
	AssetInfoModel = assetInfo.AssetInfoModel
	AssetInfo      = assetInfo.AssetInfo

	L2NftModel        = nft.L2NftModel
	L2NftHistoryModel = nft.L2NftHistoryModel

	PoolInfo = commonAsset.LiquidityInfo
)

const (
	// tx status
	TxStatusPending = tx.StatusPending
	// mempool status
	MempoolTxHandledTxStatus = mempool.SuccessTxStatus
	// block status
	BlockStatusPending = block.StatusPending
	// asset type
	GeneralAssetType         = commonAsset.GeneralAssetType
	LiquidityAssetType       = commonAsset.LiquidityAssetType
	NftAssetType             = commonAsset.NftAssetType
	CollectionNonceAssetType = commonAsset.CollectionNonceAssetType
	//MaxTxsAmountPerBlock = transactions.TxsCountPerBlock
	MaxTxsAmountPerBlock = 1

	TxTypeRegisterZns      = commonTx.TxTypeRegisterZns
	TxTypeCreatePair       = commonTx.TxTypeCreatePair
	TxTypeUpdatePairRate   = commonTx.TxTypeUpdatePairRate
	TxTypeDeposit          = commonTx.TxTypeDeposit
	TxTypeTransfer         = commonTx.TxTypeTransfer
	TxTypeSwap             = commonTx.TxTypeSwap
	TxTypeAddLiquidity     = commonTx.TxTypeAddLiquidity
	TxTypeRemoveLiquidity  = commonTx.TxTypeRemoveLiquidity
	TxTypeMintNft          = commonTx.TxTypeMintNft
	TxTypeCreateCollection = commonTx.TxTypeCreateCollection
	TxTypeTransferNft      = commonTx.TxTypeTransferNft
	TxTypeAtomicMatch      = commonTx.TxTypeAtomicMatch
	TxTypeCancelOffer      = commonTx.TxTypeCancelOffer
	TxTypeDepositNft       = commonTx.TxTypeDepositNft
	TxTypeWithdraw         = commonTx.TxTypeWithdraw
	TxTypeWithdrawNft      = commonTx.TxTypeWithdrawNft
	TxTypeFullExit         = commonTx.TxTypeFullExit
	TxTypeFullExitNft      = commonTx.TxTypeFullExitNft
)

const (
	// 15 minutes
	MaxCommitterInterval = 60 * 15
)

var (
	ZeroBigIntString = big.NewInt(0).String()
	ZeroBigInt       = big.NewInt(0)
)
