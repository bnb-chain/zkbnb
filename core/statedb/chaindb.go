package statedb

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/dao/asset"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/tx"
)

type ChainDB struct {
	// Block Chain data
	BlockModel    block.BlockModel
	TxModel       tx.TxModel
	TxDetailModel tx.TxDetailModel

	// State DB
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	L2AssetInfoModel      asset.AssetModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	L2NftModel            nft.L2NftModel
	OfferModel            nft.OfferModel
	L2NftExchangeModel    nft.L2NftExchangeModel
	L2NftHistoryModel     nft.L2NftHistoryModel
	MempoolModel          mempool.MempoolModel
}

func NewChainDB(conn sqlx.SqlConn, config cache.CacheConf, gormPointer *gorm.DB) *ChainDB {
	return &ChainDB{
		BlockModel:    block.NewBlockModel(conn, config, gormPointer),
		TxModel:       tx.NewTxModel(conn, config, gormPointer),
		TxDetailModel: tx.NewTxDetailModel(conn, config, gormPointer),

		AccountModel:          account.NewAccountModel(conn, config, gormPointer),
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, config, gormPointer),
		L2AssetInfoModel:      asset.NewAssetModel(conn, config, gormPointer),
		LiquidityModel:        liquidity.NewLiquidityModel(conn, config, gormPointer),
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(conn, config, gormPointer),
		L2NftModel:            nft.NewL2NftModel(conn, config, gormPointer),
		OfferModel:            nft.NewOfferModel(conn, config, gormPointer),
		L2NftExchangeModel:    nft.NewL2NftExchangeModel(conn, config, gormPointer),
		L2NftHistoryModel:     nft.NewL2NftHistoryModel(conn, config, gormPointer),
		MempoolModel:          mempool.NewMempoolModel(conn, config, gormPointer),
	}
}
