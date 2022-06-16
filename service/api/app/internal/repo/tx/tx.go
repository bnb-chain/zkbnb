package tx

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
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
	// result, err := m.cache.GetWithSet(ctx, cacheZecreyTxTxCountPrefix, count, 1,
	// 	multcache.SqlQueryCount, m.db, m.table,
	// 	"deleted_at is NULL")
	// if err != nil {
	// 	return 0, err
	// }
	// count, ok := result.(int64)
	// if !ok {
	// 	log.Fatal("Error type!")
	// }
	return count, nil
}

func (m *tx) GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error) {
	var (
		txDetailTable = `tx_detail`
	)
	dbTx := m.db.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Count(&count)
	return count, dbTx.Error
}
