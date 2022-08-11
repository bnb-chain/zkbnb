package logic

import (
	"math/big"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/model/tx"
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

	SysconfigModel = sysconfig.SysConfigModel
	MempoolModel   = mempool.MempoolModel
	BlockModel     = block.BlockModel

	L2NftModel        = nft.L2NftModel
	L2NftHistoryModel = nft.L2NftHistoryModel

	PoolInfo = commonAsset.LiquidityInfo
)

const (
	// tx status
	TxStatusPending = tx.StatusPending
	// asset type
	GeneralAssetType         = commonAsset.GeneralAssetType
	LiquidityAssetType       = commonAsset.LiquidityAssetType
	NftAssetType             = commonAsset.NftAssetType
	CollectionNonceAssetType = commonAsset.CollectionNonceAssetType

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
	MaxCommitterInterval = 60 * 1
)

var (
	ZeroBigInt        = big.NewInt(0)
	TxsAmountPerBlock []int
)
