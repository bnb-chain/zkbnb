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
	key := cacheMempoolTxListPrefix + fmt.Sprintf("_%v_%v", offset, limit)
	_, err = m.cache.GetWithSet(key, mempoolTx, multcache.SqlBatchQueryWithWhere, m.db, m.table, where, whereCondition, limit, offset, order)
	if err != nil {
		logx.Errorf("[mempool.GetMempoolTxs] %s", err)
		return nil, err
	}
	for _, mempoolTx := range mempoolTx {
		err := m.db.Model(&mempoolTx).Association(mempoolForeignKeyColumn).Find(&mempoolTx.MempoolDetails)
		if err != nil {
			return nil, err
		}
	}
	return mempoolTx, nil
}

func (m *mempool) GetMempoolTxsTotalCount() (count int64, err error) {
	where := "status = @status and deleted_at is NULL"
	whereCondition := sql.Named("status", PendingTxStatus)
	ct, err := m.cache.GetWithSet(cacheMempoolTxTotalCount, count, multcache.SqlQueryCountNamed, m.db, m.table, where, whereCondition)
	if err != nil {
		return 0, err
	}
	return ct.(int64), nil
}

func (m *mempool) GetMempoolTxByTxHash(hash string) (mempoolTx *mempoolModel.MempoolTx, err error) {
	var mempoolForeignKeyColumn = `MempoolDetails`
	dbTx := m.db.Table(m.table).Where("status = ? and tx_hash = ?", PendingTxStatus, hash).Find(&mempoolTx)
	err = m.db.Model(&mempoolTx).Association(mempoolForeignKeyColumn).Find(&mempoolTx.MempoolDetails)
	if err != nil {
		logx.Errorf("[mempool.GetMempoolTxByTxHash] Get Associate MempoolDetails Error")
		return nil, err
	}
	return mempoolTx, dbTx.Error
}

func (m *mempool) GetMempoolTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*mempoolModel.MempoolTxDetail
	dbTx := m.db.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	dbTx = m.db.Table(m.table).Where("status = ? and id in (?) and deleted_at is NULL", PendingTxStatus, mempoolIds).Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}
