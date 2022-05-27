package mempool

import (
	"database/sql"
	"fmt"
	mempoolModel "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

var (
	cacheMempoolTxListPrefix = "cache:mempool:txList"
	cacheMempoolTxTotalCount = "cache:mempool:totalCount"
	//"cache:AccountsHistoryList_%v_%v", limit, offset
)

type mempool struct {
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	cache      multcache.MultCache
	redisConn  *redis.Redis
}

/*
	Func: GetMempoolTxs
	Params: offset uint64, limit uint64
	Return: mempoolTx []*mempoolModel.MempoolTx, err error
	Description: query txs from db that sit in the range
*/
func (m *mempool) GetMempoolTxs(offset int64, limit int64) (mempoolTx []*mempoolModel.MempoolTx, err error) {
	var mempoolForeignKeyColumn = `MempoolDetails`
	where := "status = @status"
	whereCondition := sql.Named("status", PendingTxStatus)
	order := "created_at desc, id desc"
	_, dberr := m.cache.GetWithSet(cacheMempoolTxListPrefix+fmt.Sprintf("_%v_%v", offset, limit), mempoolTx, multcache.SqlBatchQueryWithWhere, m.db, m.table, where, whereCondition, limit, offset, order)
	if err != nil {
		logx.Errorf("[mempool.GetMempoolTxs] %s", dberr)
		return nil, dberr
	}
	for _, mempoolTx := range mempoolTx {
		err := m.db.Model(&mempoolTx).Association(mempoolForeignKeyColumn).Find(&mempoolTx.MempoolDetails)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsList] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTx, nil
}

func (m *mempool) GetMempoolTxsTotalCount() (count int64, err error) {
	where := "status = @status and deleted_at is NULL"
	whereCondition := sql.Named("status", PendingTxStatus)
	ct, dberr := m.cache.GetWithSet(cacheMempoolTxTotalCount, count, multcache.SqlQueryCountNamed, m.db, m.table, where, whereCondition)
	if dberr != nil {
		logx.Errorf("[tx.GetTxsTotalCount] %s", dberr)
		return 0, dberr
	}
	return ct.(int64), nil
}
