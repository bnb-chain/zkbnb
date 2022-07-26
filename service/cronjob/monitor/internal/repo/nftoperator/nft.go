package nftoperator

import (
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) CreateNfts(pendingNewNfts []*nft.L2Nft) (err error) {
	if len(pendingNewNfts) != 0 {
		return nil
	}
	dbTx := m.db.Table(nft.L2NftTableName).CreateInBatches(pendingNewNfts, len(pendingNewNfts))
	if dbTx.Error != nil {
		logx.Errorf("[CreateMempoolAndActiveAccount] unable to create pending new nft infos: %s", dbTx.Error.Error())
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(pendingNewNfts)) {
		logx.Errorf("[CreateMempoolAndActiveAccount] invalid new nft infos")
		return errors.New("[CreateMempoolAndActiveAccount] invalid new nft infos")
	}
	return nil
}
