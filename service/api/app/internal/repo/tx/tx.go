package tx

import (
	"log"

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
	sqlc.CachedConn
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
