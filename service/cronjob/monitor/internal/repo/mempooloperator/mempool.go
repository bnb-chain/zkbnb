package mempooloperator

import (
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) CreateMempoolTxs(pendingNewMempoolTxs []*mempool.MempoolTx) (err error) {
	dbTx := m.db.Table(mempool.MempoolTableName).CreateInBatches(pendingNewMempoolTxs, len(pendingNewMempoolTxs))
	if dbTx.Error != nil {
		logx.Errorf("[CreateInBatches] unable to create pending new mempool txs: %s", dbTx.Error.Error())
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(pendingNewMempoolTxs)) {
		logx.Errorf("[CreateInBatches] invalid new mempool txs")
		return errors.New("[CreateInBatches] invalid new mempool txs")
	}
	return nil

}

func (m *model) DeleteMempoolTxs(pendingUpdateMempoolTxs []*mempool.MempoolTx) (err error) {
	for _, pendingDeleteMempoolTx := range pendingUpdateMempoolTxs {
		for _, detail := range pendingDeleteMempoolTx.MempoolDetails {
			dbTx := m.db.Table(mempool.DetailTableName).Where("id = ?", detail.ID).Delete(&detail)
			if dbTx.Error != nil {
				logx.Errorf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Delete Invalid Mempool Tx")
				return errors.New("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Delete Invalid Mempool Tx")
			}
		}
		dbTx := m.db.Table(mempool.MempoolTableName).Where("id = ?", pendingDeleteMempoolTx.ID).Delete(&pendingDeleteMempoolTx)
		if dbTx.Error != nil {
			logx.Errorf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Error("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Delete Invalid Mempool Tx")
			return errors.New("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Delete Invalid Mempool Tx")
		}
	}
	return nil
}
