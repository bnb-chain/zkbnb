package tx

import (
	"context"
	"fmt"
	"sort"

	table "github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/errcode"
	"gorm.io/gorm"
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
func (m *model) GetTxsTotalCount(ctx context.Context) (count int64, err error) {
	dbTx := m.db.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error == errcode.ErrNotFound {
		return 0, nil
	}
	if dbTx.Error != nil {
		return 0, err
	}
	return count, nil
}

func (m *model) GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error) {
	var (
		txDetailTable = `tx_detail`
	)
	dbTx := m.db.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Count(&count)
	return count, dbTx.Error
}

func (m *model) GetTxByTxHash(txHash string) (tx *table.Tx, err error) {
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

func (m *model) GetTxsByBlockId(blockId int64, limit, offset uint32) (txs []table.Tx, total int64, err error) {
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

func (m *model) GetTxs(limit, offset uint32) (txs []*table.Tx, err error) {
	// dbTx := m.db.Table(m.table).Where("block_id = ?", blockId).Offset(int(offset)).Limit(int(limit)).Find(&txs)
	// if dbTx.Error != nil {
	// 	err = fmt.Errorf("[txVerification.GetTxsByBlockId] %s", dbTx.Error)
	// 	return
	// } else if dbTx.RowsAffected == 0 {
	// 	err = fmt.Errorf("[txVerification.GetTxsByBlockId] No such Tx with blockId: %v", blockId)
	// 	return
	// }
	return
}

func (m *model) GetTxByTxID(txID int64) (*table.Tx, error) {
	tx := &table.Tx{}
	dbTx := m.db.Table(m.table).Where("id = ? and deleted_at is NULL", txID).Find(&tx)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, errcode.ErrDataNotExist
	}
	return tx, nil
}
