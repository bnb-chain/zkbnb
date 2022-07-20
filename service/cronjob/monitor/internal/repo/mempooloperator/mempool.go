package mempooloperator

import (
	"errors"

	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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
