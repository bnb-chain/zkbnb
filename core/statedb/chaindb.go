package statedb

import (
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/priorityrequest"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/dao/tx"
)

type ChainDB struct {
	DB *gorm.DB
	// Block Chain data
	BlockModel           block.BlockModel
	CompressedBlockModel compressedblock.CompressedBlockModel
	TxModel              tx.TxModel
	TxDetailModel        tx.TxDetailModel
	PriorityRequestModel priorityrequest.PriorityRequestModel

	// State DB
	AccountModel        account.AccountModel
	AccountHistoryModel account.AccountHistoryModel
	L2AssetInfoModel    asset.AssetModel
	L2NftModel          nft.L2NftModel
	L2NftHistoryModel   nft.L2NftHistoryModel
	TxPoolModel         tx.TxPoolModel

	// Sys config
	SysConfigModel sysconfig.SysConfigModel
}

func NewChainDB(db *gorm.DB) *ChainDB {
	return &ChainDB{
		DB:                   db,
		BlockModel:           block.NewBlockModel(db),
		CompressedBlockModel: compressedblock.NewCompressedBlockModel(db),
		TxModel:              tx.NewTxModel(db),
		TxDetailModel:        tx.NewTxDetailModel(db),
		PriorityRequestModel: priorityrequest.NewPriorityRequestModel(db),

		AccountModel:        account.NewAccountModel(db),
		AccountHistoryModel: account.NewAccountHistoryModel(db),
		L2AssetInfoModel:    asset.NewAssetModel(db),
		L2NftModel:          nft.NewL2NftModel(db),
		L2NftHistoryModel:   nft.NewL2NftHistoryModel(db),
		TxPoolModel:         tx.NewTxPoolModel(db),

		SysConfigModel: sysconfig.NewSysConfigModel(db),
	}
}

func (c *ChainDB) Close() {
	sqlDB, err := c.DB.DB()
	if err == nil && sqlDB != nil {
		err = sqlDB.Close()
	}
	if err != nil {
		logx.Errorf("close db error: %s", err.Error())
	}
}
