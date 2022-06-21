package tx

import (
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

var (
	cacheZecreyTxIdPrefix      = "cache:zecrey:tx:id:"
	cacheZecreyTxTxHashPrefix  = "cache:zecrey:tx:txHash:"
	cacheZecreyTxTxCountPrefix = "cache:zecrey:tx:txCount"
)

type tx struct {
	table      string
	db         *gorm.DB
	cachedConn sqlc.CachedConn
	redisConn  *redis.Redis
	cache      multcache.MultCache
}

/*
	Func: GetTxsTotalCount
	Params:
	Return: count int64, err error
	Description: used for counting total transactions for explorer dashboard
*/
func (m *tx) GetTxsTotalCount() (count int64, err error) {
	// result, err := m.cache.GetWithSet(cacheZecreyTxTxCountPrefix, count,
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

func (m *tx) GetTxsTotalCountByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8) (count int64, err error) {
	var (
		txDetailTable = `tx_detail`
		txIds         []int64
	)
	dbTx := m.db.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsTotalCountByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[txVerification.GetTxsTotalCountByAccountIndexAndTxTypeArray] No Txs of account index %d  and txVerification type %v in Tx Table", accountIndex, txTypeArray)
		return 0, nil
	}
	dbTx = m.db.Table(m.table).Where("id in (?) and deleted_at is NULL and tx_type in (?)", txIds, txTypeArray).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[txVerification.GetTxsTotalCountByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[txVerification.GetTxsTotalCountByAccountIndexAndTxTypeArray] no txVerification of account index %d and txVerification type = %v in mempool", accountIndex, txTypeArray)
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetTxsListByAccountIndexAndTxTypeArray
	Params: accountIndex int64, txTypeArray []uint8, limit int, offset int
	Return: txVerification []*Tx, err error
	Description: used for getTxsListByAccountIndex API, return all txVerification related to accountIndex and txTypeArray.
				Because there are many accountIndex in
				 sorted by created_time
				 Associate With TxDetail Table
*/

func (m *tx) GetTxsListByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8, limit int, offset int) (txs []*Tx, err error) {
	var (
		txDetailTable = `tx_detail`
		txIds         []int64
	)
	dbTx := m.db.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[info.GetTxsListByAccountIndexAndTxTypeArray] Get TxIds Error")
		return nil, ErrNotExistInSql
	}
	dbTx = m.db.Table(m.table).Order("created_at desc").Where("tx_type in (?)", txTypeArray).Offset(offset).Limit(limit).Find(&txs, txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[GetTxsListByAccountIndexAndTxTypeArray] Get Txs Error")
		return nil, ErrNotExistInSql
	}
	return txs, nil
}
