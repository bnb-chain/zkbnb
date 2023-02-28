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
	"time"

	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	TxTableName = `tx`
)

const (
	StatusFailed = iota
	StatusPending
	StatusExecuted
	StatusPacked
	StatusCommitted
	StatusVerified
)

type getTxOption struct {
	Types       []int64
	Statuses    []int64
	FromHash    string
	WithDeleted bool
}

type GetTxOptionFunc func(*getTxOption)

func GetTxWithTypes(txTypes []int64) GetTxOptionFunc {
	return func(o *getTxOption) {
		o.Types = txTypes
	}
}

func GetTxWithStatuses(statuses []int64) GetTxOptionFunc {
	return func(o *getTxOption) {
		o.Statuses = statuses
	}
}

func GetTxWithFromHash(hash string) GetTxOptionFunc {
	return func(o *getTxOption) {
		o.FromHash = hash
	}
}

func GetTxWithDeleted() GetTxOptionFunc {
	return func(o *getTxOption) {
		o.WithDeleted = true
	}
}

type (
	TxModel interface {
		CreateTxTable() error
		DropTxTable() error
		GetTxsTotalCount(options ...GetTxOptionFunc) (count int64, err error)
		GetTxs(limit int64, offset int64, options ...GetTxOptionFunc) (txList []*Tx, err error)
		GetTxsByAccountIndex(accountIndex int64, limit int64, offset int64, options ...GetTxOptionFunc) (txList []*Tx, err error)
		GetTxsCountByAccountIndex(accountIndex int64, options ...GetTxOptionFunc) (count int64, err error)
		GetReplicateTxsCountByAccountIndex(accountIndex int64, options ...GetTxOptionFunc) (count int64, err error)
		GetTxByHash(txHash string) (tx *Tx, err error)
		GetTxsTotalCountBetween(from, to time.Time) (count int64, err error)
		GetDistinctAccountsCountBetween(from, to time.Time) (count int64, err error)
		UpdateTxsStatusInTransact(tx *gorm.DB, blockTxStatus map[int64]int) error
		CreateTxs(txs []*Tx) error
		DeleteByHeightsInTransact(tx *gorm.DB, heights []int64) error
		GetMaxPoolTxIdByHeightInTransact(tx *gorm.DB, height int64) (poolTxId uint, err error)
		GetCountByGreaterHeight(blockHeight int64) (count int64, err error)
	}

	defaultTxModel struct {
		table string
		DB    *gorm.DB
	}

	Tx struct {
		BaseTx
		PoolTxId uint      `gorm:"uniqueIndex"`
		VerifyAt time.Time // verify time when the transaction status changes to be StatusVerified

		Rollback bool `gorm:"-"`
		// l1 request id
		L1RequestId int64 `gorm:"-"`
	}
)

func NewTxModel(db *gorm.DB) TxModel {
	return &defaultTxModel{
		table: TxTableName,
		DB:    db,
	}
}

func (*Tx) TableName() string {
	return TxTableName
}

func (m *defaultTxModel) CreateTxTable() error {
	return m.DB.AutoMigrate(Tx{})
}

func (m *defaultTxModel) DropTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultTxModel) GetTxsTotalCount(options ...GetTxOptionFunc) (count int64, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	dbTx := m.DB.Table(m.table)
	if len(opt.Statuses) > 0 {
		dbTx = dbTx.Where("tx_status IN ?", opt.Statuses)
	}

	dbTx = dbTx.Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		if dbTx.Error == types.DbErrNotFound {
			return 0, nil
		}
		return 0, types.DbErrSqlOperation
	}
	return count, nil
}

func (m *defaultTxModel) GetTxs(limit int64, offset int64, options ...GetTxOptionFunc) (txList []*Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	dbTx := m.DB.Table(m.table)
	if len(opt.Statuses) > 0 {
		dbTx = dbTx.Where("tx_status IN ?", opt.Statuses)
	}

	dbTx = dbTx.Limit(int(limit)).Offset(int(offset)).Order("created_at desc").Find(&txList)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txList, nil
}

func (m *defaultTxModel) GetTxsByAccountIndex(accountIndex int64, limit int64, offset int64, options ...GetTxOptionFunc) (txList []*Tx, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	dbTx := m.DB.Table(m.table).Where("account_index = ? and deleted_at is null", accountIndex)
	if len(opt.Types) > 0 {
		dbTx = dbTx.Where("tx_type IN ?", opt.Types)
	}

	dbTx = dbTx.Limit(int(limit)).Offset(int(offset)).Order("id desc").Find(&txList)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txList, nil
}

func (m *defaultTxModel) GetTxsCountByAccountIndex(accountIndex int64, options ...GetTxOptionFunc) (count int64, err error) {
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

func (m *defaultTxModel) GetReplicateTxsCountByAccountIndex(accountIndex int64, options ...GetTxOptionFunc) (count int64, err error) {
	opt := &getTxOption{}
	for _, f := range options {
		f(opt)
	}

	var dbTx *gorm.DB
	if len(opt.Types) > 0 {
		dbTx = m.DB.Raw("select count(distinct tx.tx_hash) from pool_tx, tx where pool_tx.account_index = ? "+
			"and pool_tx.tx_type in (?) and pool_tx.deleted_at is null and tx.account_index = ? "+
			"and tx.tx_type in (?) and tx.deleted_at is null and pool_tx.tx_hash = tx.tx_hash", accountIndex, opt.Types, accountIndex, opt.Types).Count(&count)
	} else {
		dbTx = m.DB.Raw("select count(distinct tx.tx_hash) from pool_tx, tx where pool_tx.account_index = ? "+
			"and pool_tx.deleted_at is null and tx.account_index = ? "+
			"and tx.deleted_at is null and pool_tx.tx_hash = tx.tx_hash", accountIndex, accountIndex).Count(&count)
	}
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}

	return count, nil
}

func (m *defaultTxModel) GetTxByHash(txHash string) (tx *Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_hash = ?", txHash).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}

	return tx, nil
}

func (m *defaultTxModel) GetMaxPoolTxIdByHeightInTransact(tx *gorm.DB, blockHeight int64) (poolTxId uint, err error) {
	dbTx := tx.Table(m.table).Select("pool_tx_id").Where("block_height = ?", blockHeight).Order("pool_tx_id desc").Limit(1).Find(&poolTxId)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return poolTxId, nil
}

func (m *defaultTxModel) GetTxsTotalCountBetween(from, to time.Time) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("created_at BETWEEN ? AND ?", from, to).Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultTxModel) GetDistinctAccountsCountBetween(from, to time.Time) (count int64, err error) {
	dbTx := m.DB.Raw("SELECT count (distinct account_index) FROM tx WHERE created_at BETWEEN ? AND ? AND account_index != -1", from, to).Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultTxModel) UpdateTxsStatusInTransact(tx *gorm.DB, blockTxStatus map[int64]int) error {

	//parameters to update the transaction status
	for height, status := range blockTxStatus {

		data := &Tx{}
		data.BlockHeight = height
		data.TxStatus = status

		// For status = StatusVerified case, the verifyAt field need to be updated meanwhile
		if status == StatusVerified {
			data.VerifyAt = time.Now()
		}

		dbTx := tx.Model(&Tx{}).Where("block_height = ?", height).Updates(data)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToUpdateTx
		}
	}
	return nil
}

func (m *defaultTxModel) CreateTxs(txs []*Tx) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(txs, len(txs))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(txs)) {
		logx.Errorf("CreateTxs failed,rows affected not equal txs length,dbTx.RowsAffected:%s,len(txs):%s", int(dbTx.RowsAffected), len(txs))
		return types.DbErrFailToCreateTx
	}
	return nil
}

func (m *defaultTxModel) DeleteByHeightsInTransact(tx *gorm.DB, heights []int64) error {
	if len(heights) == 0 {
		return nil
	}
	dbTx := tx.Model(&Tx{}).Unscoped().Where("block_height in ?", heights).Delete(&Tx{})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}

func (m *defaultTxModel) GetCountByGreaterHeight(blockHeight int64) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height > ?", blockHeight).Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (ai *Tx) DeepCopy() *Tx {
	tx := &Tx{}
	tx.TxHash = ai.TxHash
	tx.TxType = ai.TxType
	tx.TxInfo = ai.TxInfo
	tx.AccountIndex = ai.AccountIndex
	tx.Nonce = ai.Nonce
	tx.ExpiredAt = ai.ExpiredAt
	tx.GasFee = ai.GasFee
	tx.GasFeeAssetId = ai.GasFeeAssetId
	tx.NftIndex = ai.NftIndex
	tx.CollectionId = ai.CollectionId
	tx.AssetId = ai.AssetId
	tx.TxAmount = ai.TxAmount
	tx.Memo = ai.Memo
	tx.ExtraInfo = ai.ExtraInfo
	tx.NativeAddress = ai.NativeAddress // a. Priority tx, assigned when created b. Other tx, assigned after executed.
	//TxDetails:     []*TxDetail `gorm:"foreignKey:TxId"`
	tx.TxIndex = ai.TxIndex
	tx.BlockHeight = ai.BlockHeight
	tx.BlockId = ai.BlockId
	tx.TxStatus = ai.TxStatus
	tx.ID = ai.ID
	tx.PoolTxId = ai.PoolTxId
	tx.CreatedAt = ai.CreatedAt
	tx.UpdatedAt = ai.UpdatedAt
	tx.DeletedAt = ai.DeletedAt
	return tx
}
