package tx

import (
	"encoding/json"
	mempoolModel "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
	"log"
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

func (m *tx) UpdateTxCache(tx *mempoolModel.MempoolTx) error {
	key := cacheZecreyTxTxHashPrefix + tx.TxHash
	txJson, err := json.Marshal(*tx)
	if err != nil {
		return zerror.New(30002, "Serializing tx error!")
	}
	m.cache.Set(key, txJson)
	return nil
}

func (m *tx) GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error) {
	var (
		txDetailTable = `tx_detail`
	)
	dbTx := m.db.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Count(&count)
	return count, dbTx.Error
}
