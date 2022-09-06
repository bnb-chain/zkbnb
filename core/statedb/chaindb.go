package statedb

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
	"github.com/bnb-chain/zkbnb/dao/mempool"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/tx"
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
	L2NftHistoryModel     nft.L2NftHistoryModel
	MempoolModel          mempool.MempoolModel
}

func NewChainDB(db *gorm.DB) *ChainDB {
	return &ChainDB{
		BlockModel:    block.NewBlockModel(db),
		TxModel:       tx.NewTxModel(db),
		TxDetailModel: tx.NewTxDetailModel(db),

		AccountModel:          account.NewAccountModel(db),
		AccountHistoryModel:   account.NewAccountHistoryModel(db),
		L2AssetInfoModel:      asset.NewAssetModel(db),
		LiquidityModel:        liquidity.NewLiquidityModel(db),
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(db),
		L2NftModel:            nft.NewL2NftModel(db),
		L2NftHistoryModel:     nft.NewL2NftHistoryModel(db),
		MempoolModel:          mempool.NewMempoolModel(db),
	}
}
