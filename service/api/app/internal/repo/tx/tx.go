package tx

import (
	"context"
	"sort"
	"time"

	"gorm.io/gorm"

	table "github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
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

func (m *model) GetTxsTotalCount(ctx context.Context) (int64, error) {
	f := func() (interface{}, error) {
		var count int64
		dbTx := m.db.Table(m.table).Where("deleted_at is NULL").Count(&count)
		if dbTx.Error != nil {
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, nil
		}
		return &count, nil
	}
	var countType int64
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyTxsCount(), &countType, 5, f)
	if err != nil {
		return 0, err
	}
	count, _ := value.(*int64)
	return *count, nil
}

func (m *model) GetTxByTxHash(ctx context.Context, txHash string) (*table.Tx, error) {
	f := func() (interface{}, error) {
		tx := &table.Tx{}
		dbTx := m.db.Table(m.table).Where("tx_hash = ?", txHash).Find(&tx)
		if dbTx.Error != nil {
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		err := m.db.Model(&tx).Association(`TxDetails`).Find(&tx.TxDetails)
		if err != nil {
			return nil, err
		}
		sort.SliceStable(tx.TxDetails, func(i, j int) bool {
			return tx.TxDetails[i].Order < tx.TxDetails[j].Order
		})
		return tx, nil
	}
	tx := &table.Tx{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyTxByTxHash(txHash), tx, 1, f)
	if err != nil {
		return nil, err
	}
	tx, _ = value.(*table.Tx)
	return tx, nil
}

func (m *model) GetTxByTxID(ctx context.Context, txID int64) (*table.Tx, error) {
	f := func() (interface{}, error) {
		tx := &table.Tx{}
		dbTx := m.db.Table(m.table).Where("id = ? and deleted_at is NULL", txID).Find(&tx)
		if dbTx.Error != nil {
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		err := m.db.Model(&tx).Association(`TxDetails`).Find(&tx.TxDetails)
		if err != nil {
			return nil, err
		}
		sort.SliceStable(tx.TxDetails, func(i, j int) bool {
			return tx.TxDetails[i].Order < tx.TxDetails[j].Order
		})
		return tx, nil
	}
	tx := &table.Tx{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyTxByTxId(txID), tx, 1, f)
	if err != nil {
		return nil, err
	}
	tx, _ = value.(*table.Tx)
	return tx, nil
}

func (m *model) GetTxCountByTimeRange(ctx context.Context, data string) (int64, error) {
	var (
		from time.Time
		to   time.Time
	)
	now := time.Now()
	today := now.Round(24 * time.Hour).Add(-8 * time.Hour)
	switch data {
	case "yesterday":
		from = today.Add(-24 * time.Hour)
		to = today
	case "today":
		from = today
		to = now
	}
	f := func() (interface{}, error) {
		var count int64
		dbTx := m.db.Table(m.table).Where("created_at BETWEEN ? AND ?", from, to).Count(&count)
		if dbTx.Error != nil {
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, nil
		}
		return &count, nil
	}
	var countType int64
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyTxCountByTimeRange(data), &countType, 5, f)
	if err != nil {
		return 0, err
	}
	count, _ := value.(*int64)
	return *count, nil
}
