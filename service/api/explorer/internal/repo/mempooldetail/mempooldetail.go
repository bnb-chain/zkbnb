package mempooldetail

import (
	mempoolModel "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"gorm.io/gorm"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) GetMempoolTxDetailByAccountIndex(accountIndex int64) ([]*mempoolModel.MempoolTxDetail, error) {
	result := make([]*mempoolModel.MempoolTxDetail, 0)
	dbTx := m.db.Table(m.table).Where("account_index = ?", accountIndex).Find(&result)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrDataNotExist
	}
	return result, nil
}
