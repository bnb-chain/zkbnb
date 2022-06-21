package tx

import (
	"context"

	"github.com/bnb-chain/zkbas/pkg/multcache"
	"gorm.io/gorm"
)

type tx struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

/*
	Func: GetTxsTotalCount
	Params:
	Return: count int64, err error
	Description: used for counting total transactions for explorer dashboard
*/
func (m *tx) GetTxsTotalCount(ctx context.Context) (count int64, err error) {
	dbTx := m.db.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error == ErrNotFound {
		return 0, nil
	}
	if dbTx.Error != nil {
		return 0, err
	}
	return count, nil
}

func (m *tx) GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error) {
	var (
		txDetailTable = `tx_detail`
	)
	dbTx := m.db.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Count(&count)
	return count, dbTx.Error
}
