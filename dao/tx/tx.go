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

type (
	TxModel interface {
		CreateTxTable() error
		DropTxTable() error
		GetTxsTotalCount() (count int64, err error)
		GetTxs(limit int64, offset int64) (txList []*Tx, err error)
		GetTxsByAccountIndex(accountIndex int64, limit int64, offset int64) (txList []*Tx, err error)
		GetTxsCountByAccountIndex(accountIndex int64) (count int64, err error)
		GetTxByHash(txHash string) (tx *Tx, err error)
		GetTxsTotalCountBetween(from, to time.Time) (count int64, err error)
		GetDistinctAccountsCountBetween(from, to time.Time) (count int64, err error)
		UpdateTxsStatusInTransact(tx *gorm.DB, blockTxStatus map[int64]int) error
	}

	defaultTxModel struct {
		table string
		DB    *gorm.DB
	}

	Tx struct {
		gorm.Model

		// Assigned when created in the tx pool.
		TxHash       string `gorm:"uniqueIndex"`
		TxType       int64
		TxInfo       string
		AccountIndex int64
		Nonce        int64
		ExpiredAt    int64

		// Assigned after executed.
		GasFee        string
		GasFeeAssetId int64
		PairIndex     int64
		NftIndex      int64
		CollectionId  int64
		AssetId       int64
		TxAmount      string
		Memo          string
		ExtraInfo     string
		NativeAddress string      // a. Priority tx, assigned when created b. Other tx, assigned after executed.
		TxDetails     []*TxDetail `gorm:"foreignKey:TxId"`

		TxIndex     int64
		BlockHeight int64 `gorm:"index"`
		BlockId     int64 `gorm:"index"`
		TxStatus    int   `gorm:"index"`
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

func (m *defaultTxModel) GetTxsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		if dbTx.Error == types.DbErrNotFound {
			return 0, nil
		}
		return 0, types.DbErrSqlOperation
	}
	return count, nil
}

func (m *defaultTxModel) GetTxs(limit int64, offset int64) (txList []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("created_at desc").Find(&txList)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txList, nil
}

func (m *defaultTxModel) GetTxsByAccountIndex(accountIndex int64, limit int64, offset int64) (txList []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Limit(int(limit)).Offset(int(offset)).Order("created_at desc").Find(&txList)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txList, nil
}

func (m *defaultTxModel) GetTxsCountByAccountIndex(accountIndex int64) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Count(&count)
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
	for height, status := range blockTxStatus {
		dbTx := tx.Table(m.table).Where("block_height = ?", height).Update("tx_status", status)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToUpdateTx
		}
	}
	return nil
}
