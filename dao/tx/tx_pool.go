/*
 * Copyright © 2021 ZkBNB Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package tx

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	PoolTxTableName = `pool_tx`
)

type (
	TxPoolModel interface {
		CreatePoolTxTable() error
		DropPoolTxTable() error
		GetTxs(limit int64, offset int64, options ...GetTxOptionFunc) (txs []*Tx, err error)
		GetTxsTotalCount(options ...GetTxOptionFunc) (count int64, err error)
		GetTxByTxHash(hash string) (txs *Tx, err error)
		GetTxsByStatus(status int) (txs []*Tx, err error)
		GetTxsByStatusAndMaxId(status int, maxId uint, limit int64) (txs []*Tx, err error)
		CreateTxs(txs []*Tx) error
		GetPendingTxsByAccountIndex(accountIndex int64, options ...GetTxOptionFunc) (txs []*Tx, err error)
		GetMaxNonceByAccountIndex(accountIndex int64) (nonce int64, err error)
		CreateTxsInTransact(tx *gorm.DB, txs []*Tx) error
		UpdateTxsInTransact(tx *gorm.DB, txs []*Tx) error
		DeleteTxsInTransact(tx *gorm.DB, txs []*Tx) error
		DeleteTxsBatchInTransact(tx *gorm.DB, txs []*Tx) error
		GetLatestTx(txTypes []int64, statuses []int) (tx *Tx, err error)
		GetFirstTxByStatus(status int) (tx *Tx, err error)
		UpdateTxsToPending() error
		GetLatestExecutedTx() (tx *Tx, err error)
		GetTxsPageByStatus(status int, limit int64) (txs []*Tx, err error)
		UpdateTxsStatusByIds(ids []uint, status int) error
		UpdateTxsStatusAndHeightByIds(ids []uint, status int, blockHeight int64) error
		DeleteTxsBatch(txs []*Tx) error
	}

	defaultTxPoolModel struct {
		table string
		DB    *gorm.DB
	}

	PoolTx struct {
		Tx
	}
)

func NewTxPoolModel(db *gorm.DB) TxPoolModel {
	return &defaultTxPoolModel{
		table: PoolTxTableName,
		DB:    db,
	}
}

func (*PoolTx) TableName() string {
	return PoolTxTableName
}

func (m *defaultTxPoolModel) CreatePoolTxTable() error {
	return m.DB.AutoMigrate(PoolTx{})
}

func (m *defaultTxPoolModel) DropPoolTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultTxPoolModel) GetTxs(limit int64, offset int64, options ...GetTxOptionFunc) (txs []*Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	dbTx := m.DB.Table(m.table)
	subTx := m.DB.Table(m.table)

	if opt.WithDeleted {
		dbTx = dbTx.Unscoped()
		subTx = subTx.Unscoped()
	}
	if len(opt.Statuses) > 0 {
		dbTx = dbTx.Where("tx_status IN ?", opt.Statuses)
	}
	if len(opt.FromHash) > 0 {
		subTx = subTx.Select("id").Where("tx_hash = ?", opt.FromHash).Limit(1)
		dbTx = dbTx.Where("id > (?)", subTx)
	}

	dbTx = dbTx.Limit(int(limit)).Offset(int(offset)).Order("created_at desc, id desc").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetTxsByStatus(status int) (txs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ?", status).Order("created_at, id").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetTxsPageByStatus(status int, limit int64) (txs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Where("tx_status = ?", status).Order("created_at, id").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetTxsByStatusAndMaxId(status int, maxId uint, limit int64) (txs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Where("tx_status = ? and id > ?", status, maxId).Order("id asc").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetTxsTotalCount(options ...GetTxOptionFunc) (count int64, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	dbTx := m.DB.Table(m.table)
	subTx := m.DB.Table(m.table)

	if opt.WithDeleted {
		dbTx = dbTx.Unscoped()
		subTx = subTx.Unscoped()
	}
	if len(opt.Statuses) > 0 {
		dbTx = dbTx.Where("tx_status IN ?", opt.Statuses)
	}
	if len(opt.FromHash) > 0 {
		subTx = subTx.Select("id").Where("tx_hash = ?", opt.FromHash).Limit(1)
		dbTx = dbTx.Where("id > (?)", subTx)
	}

	dbTx = dbTx.Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultTxPoolModel) GetTxByTxHash(hash string) (tx *Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_hash = ?", hash).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultTxPoolModel) CreateTxs(txs []*Tx) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(txs)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToCreatePoolTx
		}
		return nil
	})
}

func (m *defaultTxPoolModel) GetPendingTxsByAccountIndex(accountIndex int64, options ...GetTxOptionFunc) (txs []*Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	dbTx := m.DB.Table(m.table).Where("tx_status = ? AND account_index = ?", StatusPending, accountIndex)
	if len(opt.Types) > 0 {
		dbTx = dbTx.Where("tx_type IN ?", opt.Types)
	}

	dbTx = dbTx.Order("created_at, id").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetMaxNonceByAccountIndex(accountIndex int64) (nonce int64, err error) {
	dbTx := m.DB.Table(m.table).Select("nonce").Where("deleted_at is null and account_index = ?", accountIndex).Order("nonce desc").Limit(1).Find(&nonce)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return nonce, nil
}

func (m *defaultTxPoolModel) CreateTxsInTransact(tx *gorm.DB, txs []*Tx) error {
	dbTx := tx.Table(m.table).CreateInBatches(txs, len(txs))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreatePoolTx
	}
	return nil
}

func (m *defaultTxPoolModel) UpdateTxsInTransact(tx *gorm.DB, txs []*Tx) error {
	for _, poolTx := range txs {
		// Don't write tx details when update tx pool.
		txDetails := poolTx.TxDetails
		poolTx.TxDetails = nil
		dbTx := tx.Scopes().Table(m.table).Where("id = ?", poolTx.ID).
			Select("*").
			Updates(&poolTx)
		poolTx.TxDetails = txDetails
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToUpdatePoolTx
		}
	}
	return nil
}

func (m *defaultTxPoolModel) UpdateTxsStatusByIds(ids []uint, status int) error {
	dbTx := m.DB.Model(&PoolTx{}).Where("id in ?", ids).Update("tx_status", status)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}

func (m *defaultTxPoolModel) UpdateTxsStatusAndHeightByIds(ids []uint, status int, blockHeight int64) error {
	dbTx := m.DB.Model(&PoolTx{}).Select("TxStatus", "BlockHeight").Where("id in ?", ids).Updates(map[string]interface{}{
		"status":       status,
		"block_height": blockHeight,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}

	return nil
}

func (m *defaultTxPoolModel) UpdateTxsToPending() error {
	var statuses = []int{StatusProcessing, StatusExecuted}
	dbTx := m.DB.Model(&PoolTx{}).Where("tx_status in ? ", statuses).Update("tx_status", StatusPending)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}

func (m *defaultTxPoolModel) DeleteTxsInTransact(tx *gorm.DB, txs []*Tx) error {
	for _, poolTx := range txs {
		dbTx := tx.Table(m.table).Where("id = ?", poolTx.ID).Delete(&poolTx)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToDeletePoolTx
		}
	}
	return nil
}

func (m *defaultTxPoolModel) DeleteTxsBatchInTransact(tx *gorm.DB, txs []*Tx) error {
	if len(txs) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(txs))
	for _, poolTx := range txs {
		ids = append(ids, poolTx.ID)
	}
	dbTx := tx.Table(m.table).Where("id in ?", ids).Delete(&Tx{})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToDeletePoolTx
	}
	return nil
}

func (m *defaultTxPoolModel) DeleteTxsBatch(txs []*Tx) error {
	if len(txs) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(txs))
	for _, poolTx := range txs {
		ids = append(ids, poolTx.ID)
	}
	dbTx := m.DB.Table(m.table).Where("id in ?", ids).Delete(&Tx{})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToDeletePoolTx
	}
	return nil
}

func (m *defaultTxPoolModel) GetLatestTx(txTypes []int64, statuses []int) (tx *Tx, err error) {

	dbTx := m.DB.Table(m.table).Where("tx_status IN ? AND tx_type IN ?", statuses, txTypes).Order("id DESC").Limit(1).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}

	return tx, nil
}

func (m *defaultTxPoolModel) GetFirstTxByStatus(status int) (txs *Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ?", status).Order("id asc").Limit(1).Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, nil
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetProposingBlockHeight() (ids []int64, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ? and deleted_at is null", StatusExecuted).Select("id").Order("id asc").Find(&ids)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return ids, nil
}

func (m *defaultTxPoolModel) GetLatestExecutedTx() (tx *Tx, err error) {
	var statuses = []int{StatusFailed, StatusExecuted}
	dbTx := m.DB.Table(m.table).Unscoped().Where("tx_status IN ?", statuses).Order("id DESC").Limit(1).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, nil
	}
	return tx, nil
}
