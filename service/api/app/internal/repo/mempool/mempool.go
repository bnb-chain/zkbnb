package mempool

import (
	"context"
	"sort"

	table "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/errcode"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

/*
	Func: GetMempoolTxs
	Params: offset uint64, limit uint64
	Return: mempoolTx []*mempoolModel.MempoolTx, err error
	Description: query txs from db that sit in the range
*/
func (m *model) GetMempoolTxs(offset int64, limit int64) (mempoolTxs []*table.MempoolTx, err error) {
	var mempoolForeignKeyColumn = `MempoolDetails`
	dbTx := m.db.Table(m.table).Order("created_at, id").Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsList] %s", dbTx.Error)
		return nil, dbTx.Error
	}
	for _, mempoolTx := range mempoolTxs {
		err := m.db.Model(&mempoolTx).Association(mempoolForeignKeyColumn).Find(&mempoolTx.MempoolDetails)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsList] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *model) GetMempoolTxsTotalCount() (count int64, err error) {
	dbTx := m.db.Table(m.table).Where("status = ? and deleted_at is NULL", PendingTxStatus).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[txVerification.GetTxsTotalCount] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *model) GetMempoolTxByTxHash(hash string) (mempoolTx *table.MempoolTx, err error) {
	var mempoolForeignKeyColumn = `MempoolDetails`
	dbTx := m.db.Table(m.table).Where("status = ? and tx_hash = ?", PendingTxStatus, hash).Find(&mempoolTx)
	err = m.db.Model(&mempoolTx).Association(mempoolForeignKeyColumn).Find(&mempoolTx.MempoolDetails)
	if err != nil {
		logx.Errorf("[mempool.GetMempoolTxByTxHash] Get Associate MempoolDetails Error")
		return nil, err
	}
	return mempoolTx, dbTx.Error
}

func (m *model) GetMempoolTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*table.MempoolTxDetail
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

func (m *model) GetMempoolTxByTxId(ctx context.Context, txID int64) (*table.MempoolTx, error) {
	f := func() (interface{}, error) {
		tx := &table.MempoolTx{}
		dbTx := m.db.Table(m.table).Where("id = ? and deleted_at is NULL", txID).Find(&tx)
		if dbTx.Error != nil {
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			return nil, errcode.ErrDataNotExist
		}
		err := m.db.Model(&tx).Association(`MempoolDetails`).Find(&tx.MempoolDetails)
		if err != nil {
			return nil, err
		}
		sort.SliceStable(tx.MempoolDetails, func(i, j int) bool {
			return tx.MempoolDetails[i].Order < tx.MempoolDetails[j].Order
		})
		return tx, nil
	}
	tx := &table.MempoolTx{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyTxByTxId(txID), tx, 1, f)
	if err != nil {
		return nil, err
	}
	tx, _ = value.(*table.MempoolTx)
	return tx, nil
}
