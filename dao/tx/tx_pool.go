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
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"

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
		GetTxUnscopedByTxHash(hash string) (txs *Tx, err error)
		GetTxsByStatus(status int) (txs []*Tx, err error)
		GetTxsByAccountIndex(accountIndex int64, limit int64, offset int64, options ...GetTxOptionFunc) (txs []*Tx, err error)
		GetTxsCountByAccountIndex(accountIndex int64, options ...GetTxOptionFunc) (count int64, err error)
		GetTxsByStatusAndMaxId(status int, maxId uint, limit int64) (txs []*Tx, err error)
		GetTxsByStatusAndIdRange(status int, fromId uint, toId uint) (txs []*Tx, err error)
		GetTxsByStatusAndCreateTime(status int, fromCreatedAt time.Time, toId uint) (txs []*Tx, err error)
		CreateTxs(txs []*PoolTx) error
		GetPendingTxsByAccountIndex(accountIndex int64, options ...GetTxOptionFunc) (txs []*Tx, err error)
		GetMaxNonceByAccountIndex(accountIndex int64) (nonce int64, err error)
		CreateTxsInTransact(tx *gorm.DB, txs []*PoolTx) error
		DeleteTxsInTransact(tx *gorm.DB, txs []*Tx) error
		DeleteTxsBatchInTransact(tx *gorm.DB, txs []*Tx) error
		GetLatestTx(txTypes []int64, statuses []int) (tx *Tx, err error)
		GetFirstTxByStatus(status int) (tx *Tx, err error)
		UpdateTxsToPending(tx *gorm.DB) error
		GetLatestExecutedTx() (tx *Tx, err error)
		GetTxsPageByStatus(status int, limit int64) (txs []*Tx, err error)
		UpdateTxsStatusByIds(ids []uint, status int) error
		UpdateTxsStatusAndHeightByIds(ids []uint, status int, blockHeight int64) error
		DeleteTxsBatch(poolTxIds []uint, status int, blockHeight int64) error
		DeleteTxIdsBatchInTransact(tx *gorm.DB, ids []uint) error
		UpdateTxsToPendingByHeights(tx *gorm.DB, blockHeight []int64) error
		UpdateTxsToPendingByMaxId(tx *gorm.DB, maxPoolTxId uint) error
		BatchUpdateNftIndexOrCollectionId(txs []*PoolTx) (err error)
		GetLatestMintNft() (tx *Tx, err error)
		GetTxsUnscopedByHeights(blockHeights []int64) (txs []*Tx, err error)
		GetLatestRollback(status int, rollback bool) (tx *PoolTx, err error)
		GetCountByGreaterHeight(blockHeight int64) (count int64, err error)
	}

	defaultTxPoolModel struct {
		table string
		DB    *gorm.DB
	}

	PoolTx struct {
		BaseTx
		Rollback bool
		// l1 request id
		L1RequestId int64
	}

	BaseTx struct {
		gorm.Model

		// Assigned when created in the tx pool.
		TxHash       string `gorm:"uniqueIndex"`
		TxType       int64
		TxInfo       string
		AccountIndex int64 `gorm:"index:idx_pool_tx_account_index_nonce,priority:1"`
		Nonce        int64 `gorm:"index:idx_pool_tx_account_index_nonce,priority:2"`
		ExpiredAt    int64

		// Assigned after executed.
		GasFee        string
		GasFeeAssetId int64
		NftIndex      int64
		CollectionId  int64
		AssetId       int64
		TxAmount      string
		Memo          string
		ExtraInfo     string
		NativeAddress string // a. Priority tx, assigned when created b. Other tx, assigned after executed.

		TxIndex     int64
		BlockHeight int64 `gorm:"index"`
		BlockId     uint  `gorm:"index"`
		TxStatus    int   `gorm:"index"`

		TxDetails []*TxDetail `gorm:"-"`
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

func (m *defaultTxPoolModel) GetTxsUnscopedByHeights(blockHeights []int64) (txs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Unscoped().Select("id,tx_hash").Where("tx_status = ? and block_height in ?", StatusExecuted, blockHeights).Order("id asc").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
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
	var poolTxs []*PoolTx
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Where("tx_status = ? and id > ?", status, maxId).Order("id asc").Find(&poolTxs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	for _, poolTx := range poolTxs {
		txs = append(txs, &Tx{BaseTx: poolTx.BaseTx, Rollback: poolTx.Rollback, L1RequestId: poolTx.L1RequestId})
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetTxsByStatusAndIdRange(status int, fromId uint, toId uint) (txs []*Tx, err error) {
	var poolTxs []*PoolTx
	dbTx := m.DB.Table(m.table).Where("tx_status = ? and id >= ? and id <= ?", status, fromId, toId).Order("id asc").Find(&poolTxs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	for _, poolTx := range poolTxs {
		txs = append(txs, &Tx{BaseTx: poolTx.BaseTx, Rollback: poolTx.Rollback, L1RequestId: poolTx.L1RequestId})
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetTxsByStatusAndCreateTime(status int, fromCreatedAt time.Time, toId uint) (txs []*Tx, err error) {
	var poolTxs []*PoolTx
	dbTx := m.DB.Table(m.table).Where("tx_status = ? and created_at >= ? and id <= ?", status, fromCreatedAt, toId).Order("id asc").Find(&poolTxs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	for _, poolTx := range poolTxs {
		txs = append(txs, &Tx{BaseTx: poolTx.BaseTx, Rollback: poolTx.Rollback})
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

func (m *defaultTxPoolModel) GetTxUnscopedByTxHash(hash string) (tx *Tx, err error) {
	dbTx := m.DB.Unscoped().Table(m.table).Where("tx_hash = ?", hash).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultTxPoolModel) GetTxsByAccountIndex(accountIndex int64, limit int64, offset int64, options ...GetTxOptionFunc) (txs []*Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	dbTx := m.DB.Table(m.table).Where("account_index = ? and deleted_at is null", accountIndex)
	if len(opt.Types) > 0 {
		dbTx = dbTx.Where("tx_type IN ?", opt.Types)
	}

	dbTx = dbTx.Limit(int(limit)).Offset(int(offset)).Order("created_at desc").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetTxsCountByAccountIndex(accountIndex int64, options ...GetTxOptionFunc) (count int64, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	dbTx := m.DB.Table(m.table).Where("account_index = ? and deleted_at is null", accountIndex)
	if len(opt.Types) > 0 {
		dbTx = dbTx.Where("tx_type IN ?", opt.Types)
	}

	dbTx = dbTx.Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultTxPoolModel) CreateTxs(txs []*PoolTx) error {
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

func (m *defaultTxPoolModel) CreateTxsInTransact(tx *gorm.DB, txs []*PoolTx) error {
	dbTx := tx.Table(m.table).CreateInBatches(txs, len(txs))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreatePoolTx
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
		"tx_status":    status,
		"block_height": blockHeight,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}

	return nil
}

func (m *defaultTxPoolModel) UpdateTxsToPending(tx *gorm.DB) error {
	dbTx := tx.Model(&PoolTx{}).Select("DeletedAt", "ExpiredAt", "TxStatus").Where("tx_status = ? ", StatusExecuted).Updates(map[string]interface{}{
		"deleted_at": nil,
		"expired_at": time.Now().Unix(),
		"tx_status":  StatusPending,
	})
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

func (m *defaultTxPoolModel) DeleteTxIdsBatchInTransact(tx *gorm.DB, ids []uint) error {
	if len(ids) == 0 {
		return nil
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
func (m *defaultTxPoolModel) UpdateTxsToPendingByHeights(tx *gorm.DB, blockHeights []int64) error {
	if len(blockHeights) == 0 {
		return nil
	}
	dbTx := tx.Model(&PoolTx{}).Unscoped().Select("DeletedAt", "TxStatus", "Rollback").Where("block_height in ? and tx_status = ? ", blockHeights, StatusExecuted).Updates(map[string]interface{}{
		"deleted_at": nil,
		"tx_status":  StatusPending,
		"rollback":   true,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}

func (m *defaultTxPoolModel) UpdateTxsToPendingByMaxId(tx *gorm.DB, maxId uint) error {
	if maxId == 0 {
		return nil
	}
	dbTx := tx.Model(&PoolTx{}).Unscoped().Select("DeletedAt", "TxStatus").Where("id > ? and tx_status= ? and deleted_at is not null", maxId, StatusFailed).Updates(map[string]interface{}{
		"deleted_at": nil,
		"tx_status":  StatusPending,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}

func (m *defaultTxPoolModel) DeleteTxsBatch(poolTxIds []uint, status int, blockHeight int64) error {
	if len(poolTxIds) == 0 {
		return nil
	}
	dbTx := m.DB.Model(&PoolTx{}).Select("DeletedAt", "BlockHeight", "TxStatus").Where("id in ?", poolTxIds).Updates(map[string]interface{}{
		"deleted_at":   time.Now(),
		"block_height": blockHeight,
		"tx_status":    status,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}

func (m *defaultTxPoolModel) GetLatestTx(txTypes []int64, statuses []int) (tx *Tx, err error) {
	var poolTx *PoolTx
	dbTx := m.DB.Table(m.table).Where("tx_status IN ? AND tx_type IN ?", statuses, txTypes).Order("id DESC").Limit(1).Find(&poolTx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	tx = &Tx{BaseTx: poolTx.BaseTx, Rollback: poolTx.Rollback, L1RequestId: poolTx.L1RequestId}
	return tx, nil
}

func (m *defaultTxPoolModel) GetLatestMintNft() (tx *Tx, err error) {

	dbTx := m.DB.Table(m.table).Unscoped().Where("tx_type = ?", types.TxTypeMintNft).Order("nft_index DESC").Limit(1).Find(&tx)
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

func (m *defaultTxPoolModel) GetLatestExecutedTx() (tx *Tx, err error) {
	dbTx := m.DB.Table(m.table).Unscoped().Where("tx_status = ?", StatusExecuted).Order("id DESC").Limit(1).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultTxPoolModel) BatchUpdateNftIndexOrCollectionId(txs []*PoolTx) (err error) {
	dbTx := m.DB.Table(m.table).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"nft_index", "collection_id"}),
	}).CreateInBatches(&txs, len(txs))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if int(dbTx.RowsAffected) != len(txs) {
		logx.Errorf("BatchUpdateNftIndexOrCollectionId failed,rows affected not equal txs length,dbTx.RowsAffected:%s,len(txs):%s", int(dbTx.RowsAffected), len(txs))
		return types.DbErrFailToUpdatePoolTx
	}
	return nil
}

func (m *defaultTxPoolModel) GetLatestRollback(status int, rollback bool) (tx *PoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ? and rollback = ?", status, rollback).Order("id DESC").Limit(1).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultTxPoolModel) GetCountByGreaterHeight(blockHeight int64) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height > ? and tx_status= ?", blockHeight, StatusExecuted).Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}
