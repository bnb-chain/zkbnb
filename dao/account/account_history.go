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
 */

package account

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	AccountHistoryTableName = `account_history`
)

type (
	AccountHistoryModel interface {
		CreateAccountHistoryTable() error
		DropAccountHistoryTable() error
		GetValidAccounts(height int64, limit int, offset int) (rowsAffected int64, accounts []*AccountHistory, err error)
		GetValidAccountCount(height int64) (accounts int64, err error)
		GetAccountByIndex(accountIndex int64, height int64) (*AccountHistory, error)
		CreateAccountHistoriesInTransact(tx *gorm.DB, histories []*AccountHistory) error
	}

	defaultAccountHistoryModel struct {
		table string
		DB    *gorm.DB
	}

	AccountHistory struct {
		gorm.Model
		AccountIndex    int64 `gorm:"index"`
		Nonce           int64
		CollectionNonce int64
		AssetInfo       string
		AssetRoot       string
		L2BlockHeight   int64
	}
)

func NewAccountHistoryModel(db *gorm.DB) AccountHistoryModel {
	return &defaultAccountHistoryModel{
		table: AccountHistoryTableName,
		DB:    db,
	}
}

func (*AccountHistory) TableName() string {
	return AccountHistoryTableName
}

func (m *defaultAccountHistoryModel) CreateAccountHistoryTable() error {
	return m.DB.AutoMigrate(AccountHistory{})
}

func (m *defaultAccountHistoryModel) DropAccountHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultAccountHistoryModel) GetValidAccounts(height int64, limit int, offset int) (rowsAffected int64, accounts []*AccountHistory, err error) {
	subQuery := m.DB.Table(m.table).Select("*").
		Where("account_index = a.account_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height AND l2_block_height != -1", height)

	dbTx := m.DB.Table(m.table+" as a").Select("*").
		Where("NOT EXISTS (?) AND l2_block_height <= ? AND l2_block_height != -1", subQuery, height).
		Limit(limit).Offset(offset).
		Order("account_index")

	if dbTx.Find(&accounts).Error != nil {
		return 0, nil, types.DbErrSqlOperation
	}
	return dbTx.RowsAffected, accounts, nil

}

func (m *defaultAccountHistoryModel) GetValidAccountCount(height int64) (count int64, err error) {
	subQuery := m.DB.Table(m.table).Select("*").
		Where("account_index = a.account_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height AND l2_block_height != -1", height)

	dbTx := m.DB.Table(m.table+" as a").
		Where("NOT EXISTS (?) AND l2_block_height <= ? AND l2_block_height != -1", subQuery, height)

	if dbTx.Count(&count).Error != nil {
		return 0, types.DbErrSqlOperation
	}
	return count, nil
}

func (m *defaultAccountHistoryModel) CreateAccountHistoriesInTransact(tx *gorm.DB, histories []*AccountHistory) error {
	dbTx := tx.Table(m.table).CreateInBatches(histories, len(histories))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(histories)) {
		return types.DbErrFailToCreateAccountHistory
	}
	return nil
}

func (m *defaultAccountHistoryModel) GetAccountByIndex(accountIndex int64, height int64) (*AccountHistory, error) {
	var accHistory AccountHistory
	dbTx := m.DB.Table(m.table).Where("account_index = ? AND l2_block_height <= ?", accountIndex, height).Order("l2_block_height desc").Limit(1).Find(&accHistory)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return &accHistory, nil
}
