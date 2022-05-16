package logic

import (
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2asset"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/util"
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
	Block = block.Block
	// mempool
	MempoolTx       = mempool.MempoolTx
	MempoolTxDetail = mempool.MempoolTxDetail
	// assets
	L2Nft            = nft.L2Nft
	// assets history
	L2NftHistory            = nft.L2NftHistory
	// account history
	Account        = account.Account
	AccountHistory = account.AccountHistory

	SysconfigModel   = sysconfig.SysconfigModel
	MempoolModel     = mempool.MempoolModel
	BlockModel       = block.BlockModel
	L2AssetInfoModel = l2asset.L2AssetInfoModel
	L2AssetInfo      = l2asset.L2AssetInfo

	L2NftModel                   = nft.L2NftModel
	L2NftHistoryModel            = nft.L2NftHistoryModel

	PoolInfo = util.PoolInfo
)

const (
	// tx status
	TxStatusPending = tx.StatusPending
	// mempool status
	MempoolTxHandledTxStatus = mempool.SuccessTxStatus
	// block status
	BlockStatusPending = block.StatusPending
	// asset type
	GeneralAssetType     = commonAsset.GeneralAssetType
	LiquidityAssetType   = commonAsset.LiquidityAssetType
	LiquidityLpAssetType = commonAsset.LiquidityLpAssetType
	NftAssetType         = commonAsset.NftAssetType
	//MaxTxsAmountPerBlock = transactions.TxsCountPerBlock
	MaxTxsAmountPerBlock = 1

	TxTypeRegisterZns     = commonTx.TxTypeRegisterZns
	TxTypeDeposit         = commonTx.TxTypeDeposit
	TxTypeTransfer        = commonTx.TxTypeTransfer
	TxTypeSwap            = commonTx.TxTypeSwap
	TxTypeAddLiquidity    = commonTx.TxTypeAddLiquidity
	TxTypeRemoveLiquidity = commonTx.TxTypeRemoveLiquidity
	TxTypeMintNft         = commonTx.TxTypeMintNft
	TxTypeTransferNft     = commonTx.TxTypeTransferNft
	TxTypeSetNftPrice     = commonTx.TxTypeSetNftPrice
	TxTypeBuyNft          = commonTx.TxTypeBuyNft
	TxTypeDepositNft      = commonTx.TxTypeDepositNft
	TxTypeWithdraw        = commonTx.TxTypeWithdraw
	TxTypeWithdrawNft     = commonTx.TxTypeWithdrawNft
	TxTypeFullExit        = commonTx.TxTypeFullExit
	TxTypeFullExitNft     = commonTx.TxTypeFullExitNft
)

const (
	// 15 minutes
	MaxCommitterInterval = 60 * 15
)

var (
	ZeroBigIntString = big.NewInt(0).String()
)
