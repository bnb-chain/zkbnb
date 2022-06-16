package tx

import (
	"fmt"
	"log"
	"sort"

	table "github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
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
	result, err := m.cache.GetWithSet(cacheZecreyTxTxCountPrefix, count,
		multcache.SqlQueryCount, m.db, m.table,
		"deleted_at is NULL")
	if err != nil {
		return 0, err
	}
	count, ok := result.(int64)
	if !ok {
		log.Fatal("Error type!")
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

func (m *tx) GetTxByTxHash(txHash string) (tx *table.Tx, err error) {
	var txForeignKeyColumn = `TxDetails`

	dbTx := m.db.Table(m.table).Where("tx_hash = ?", txHash).Find(&tx)
	if dbTx.Error != nil {
		err = fmt.Errorf("[txVerification.GetTxByTxHash] %s", dbTx.Error)
		return nil, err
	} else if dbTx.RowsAffected == 0 {
		err = fmt.Errorf("[txVerification.GetTxByTxHash] No such Tx with txHash: %s", txHash)
		return nil, err
	}
	err = m.db.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
	if err != nil {
		err = fmt.Errorf("[txVerification.GetTxByTxHash] Get Associate TxDetails Error")
		return nil, err
	}
	// re-order tx details
	sort.SliceStable(tx.TxDetails, func(i, j int) bool {
		return tx.TxDetails[i].Order < tx.TxDetails[j].Order
	})

	return tx, nil
}

func (m *tx) GetTxsByBlockId(blockId int64, limit, offset uint32) (txs []table.Tx, total int64, err error) {
	query := m.db.Table(m.table).Where("block_id = ?", blockId)
	if err = query.Count(&total).Error; err != nil {
		err = fmt.Errorf("[txVerification.GetTxsByBlockId] %s", err)
		return
	}
	dbTx := query.Offset(int(offset)).Limit(int(limit)).Find(&txs)
	if dbTx.Error != nil {
		err = fmt.Errorf("[txVerification.GetTxsByBlockId] %s", dbTx.Error)
		return
	} else if dbTx.RowsAffected == 0 {
		err = fmt.Errorf("[txVerification.GetTxsByBlockId] No such Tx with blockId: %v", blockId)
		return
	}
	return
}
