package mempool

import (
	"database/sql"
	"fmt"

	mempoolModel "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	table "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
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
	_, err = m.cache.GetWithSet(key, mempoolTx, multcache.SqlBatchQueryWithWhere, m.db, m.table, where, whereCondition, int(limit), int(offset), order)
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

func (m *mempool) GetAccountAssetMempoolDetails(accountIndex int64, assetId int64, assetType int64) (mempoolTxDetails []*table.MempoolTxDetail, err error) {
	var dbTx *gorm.DB
	dbTx = m.db.Table(m.table).Where("account_index = ? and asset_id = ? and asset_type = ? ", accountIndex, assetId, assetType).
		Order("created_at, id").Find(&mempoolTxDetails)
	if dbTx.Error != nil {
		logx.Errorf("[mempoolDetail.GetAccountAssetMempoolDetails] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempoolDetail.GetAccountAssetMempoolDetails] Get MempoolTxDetails Error")
		return nil, ErrNotFound
	}
	return mempoolTxDetails, nil
}

func (m *mempool) GetMempoolTxsTotalCountByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8) (count int64, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*table.MempoolTxDetail
	dbTx := m.db.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsTotalCountByAccountIndexAndTxType] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	dbTx = m.db.Table(m.table).Where("status = ? and id in (?) and deleted_at is NULL  and tx_type in (?)", PendingTxStatus, mempoolIds, txTypeArray).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsTotalCountByAccountIndexAndTxType] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[mempool.GetMempoolTxsTotalCountByAccountIndexAndTxType] no txVerification of account index %d and txVerification type = %v in mempool", accountIndex, txTypeArray)
		return 0, nil
	}
	return count, nil
}

func (m *mempool) GetMempoolTxsListByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8, limit int64, offset int64) (mempoolTxs []*MempoolTx, err error) {
	var (
		mempoolDetailTable      = `mempool_tx_detail`
		mempoolIds              []int64
		mempoolForeignKeyColumn = `MempoolDetails`
	)
	var mempoolTxDetails []*table.MempoolTxDetail
	dbTx := m.db.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] Get MempoolIds Error")
		return nil, ErrNotFound
	}
	dbTx = m.db.Table(m.table).Where("status = ? and tx_type in (?)", PendingTxStatus, txTypeArray).Order("created_at desc").Offset(int(offset)).Limit(int(limit)).Find(&mempoolTxs, mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] Get MempoolTxs Error")
		return nil, ErrNotFound
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.db.Model(&mempoolTx).Association(mempoolForeignKeyColumn).Find(&mempoolTx.MempoolDetails)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}
