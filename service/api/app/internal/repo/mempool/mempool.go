package mempool

import (
	mempoolModel "github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type mempool struct {
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
func (m *mempool) GetMempoolTxs(offset int64, limit int64) (mempoolTxs []*mempoolModel.MempoolTx, err error) {
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

func (m *mempool) GetMempoolTxsTotalCount() (count int64, err error) {
	dbTx := m.db.Table(m.table).Where("status = ? and deleted_at is NULL", PendingTxStatus).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[txVerification.GetTxsTotalCount] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
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
