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

package account

import (
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	AccountTableName = `account`
)

const (
	AccountStatusPending = iota
	AccountStatusConfirmed
)

type (
	AccountModel interface {
		CreateAccountTable() error
		DropAccountTable() error
		GetAccountByIndex(accountIndex int64) (account *Account, err error)
		GetAccountByIndexes(accountIndexes []int64) (accounts []*Account, err error)
		GetConfirmedAccountByIndex(accountIndex int64) (account *Account, err error)
		GetAccountByPk(pk string) (account *Account, err error)
		GetAccountByName(name string) (account *Account, err error)
		GetAccountByNameHash(nameHash string) (account *Account, err error)
		GetAccounts(limit int, offset int64) (accounts []*Account, err error)
		GetAccountsTotalCount() (count int64, err error)
		UpdateAccountsInTransact(tx *gorm.DB, accounts []*Account) error
		GetUsers(limit int64, offset int64) (accounts []*Account, err error)
		BatchInsertOrUpdateInTransact(tx *gorm.DB, accounts []*Account) (err error)
		UpdateByIndexInTransact(tx *gorm.DB, account *Account) error
		DeleteByIndexesInTransact(tx *gorm.DB, accountIndexes []int64) error
		GetCountByGreaterHeight(blockHeight int64) (count int64, err error)
		GetMaxAccountIndex() (accountIndex int64, err error)
		GetByAccountIndexRange(fromAccountIndex int64, toAccountIndex int64) (accounts []*Account, err error)
	}

	defaultAccountModel struct {
		table string
		DB    *gorm.DB
	}

	/*
		always keep the latest data of committer
	*/
	Account struct {
		gorm.Model
		AccountIndex    int64  `gorm:"uniqueIndex"`
		AccountName     string `gorm:"uniqueIndex"`
		PublicKey       string `gorm:"uniqueIndex"`
		AccountNameHash string `gorm:"uniqueIndex"`
		L1Address       string
		Nonce           int64
		CollectionNonce int64
		// map[int64]*AccountAsset
		AssetInfo     string
		AssetRoot     string
		L2BlockHeight int64 `gorm:"index"`
		// 0 - registered, not committer 1 - committer
		Status int
	}
)

func NewAccountModel(db *gorm.DB) AccountModel {
	return &defaultAccountModel{
		table: AccountTableName,
		DB:    db,
	}
}

func (*Account) TableName() string {
	return AccountTableName
}

func (m *defaultAccountModel) CreateAccountTable() error {
	return m.DB.AutoMigrate(Account{})
}

func (m *defaultAccountModel) DropAccountTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultAccountModel) GetAccountByIndex(accountIndex int64) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&account)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountModel) GetAccountByIndexes(accountIndexes []int64) (accounts []*Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index in ?", accountIndexes).Find(&accounts)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return accounts, nil
}

func (m *defaultAccountModel) GetAccountByPk(pk string) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("public_key = ?", pk).Find(&account)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountModel) GetAccountByName(accountName string) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name = ?", accountName).Find(&account)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountModel) GetAccountByNameHash(accountNameHash string) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name_hash = ?", accountNameHash).Find(&account)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountModel) GetAccounts(limit int, offset int64) (accounts []*Account, err error) {
	dbTx := m.DB.Table(m.table).Limit(limit).Offset(int(offset)).Order("account_index desc").Find(&accounts)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return accounts, nil
}

func (m *defaultAccountModel) GetAccountsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultAccountModel) GetConfirmedAccountByIndex(accountIndex int64) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and status = ?", accountIndex, AccountStatusConfirmed).Find(&account)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountModel) UpdateAccountsInTransact(tx *gorm.DB, accounts []*Account) error {
	const CreatedAt = "CreatedAt"
	for _, account := range accounts {
		dbTx := tx.Table(m.table).Where("account_index = ?", account.AccountIndex).
			Omit(CreatedAt).
			Select("*").
			Updates(&account)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			// this account is new, we need create first
			dbTx = tx.Table(m.table).Create(&account)
			if dbTx.Error != nil {
				return dbTx.Error
			}
		}
	}
	return nil
}

func (m *defaultAccountModel) UpdateByIndexInTransact(tx *gorm.DB, account *Account) error {
	dbTx := tx.Model(&Account{}).Select("Nonce", "CollectionNonce", "AssetInfo", "AssetRoot", "L2BlockHeight").Where("account_index = ?", account.AccountIndex).Updates(map[string]interface{}{
		"nonce":            account.Nonce,
		"collection_nonce": account.CollectionNonce,
		"asset_info":       account.AssetInfo,
		"asset_root":       account.AssetRoot,
		"l2_block_height":  account.L2BlockHeight,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToUpdateAccount
	}
	return nil
}

func (m *defaultAccountModel) DeleteByIndexesInTransact(tx *gorm.DB, accountIndexes []int64) error {
	if len(accountIndexes) == 0 {
		return nil
	}
	dbTx := tx.Model(&Account{}).Unscoped().Where("account_index in ?", accountIndexes).Delete(&Account{})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToUpdateAccount
	}
	return nil
}

func (m *defaultAccountModel) GetUsers(limit int64, offset int64) (accounts []*Account, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("id asc").Find(&accounts)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, nil
	}
	return accounts, nil
}

func (m *defaultAccountModel) BatchInsertOrUpdateInTransact(tx *gorm.DB, accounts []*Account) (err error) {
	dbTx := tx.Table(m.table).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"nonce", "collection_nonce", "asset_info", "asset_root", "l2_block_height"}),
	}).CreateInBatches(&accounts, len(accounts))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if int(dbTx.RowsAffected) != len(accounts) {
		logx.Errorf("BatchInsertOrUpdateInTransact failed,rows affected not equal accounts length,dbTx.RowsAffected:%s,len(accounts):%s", int(dbTx.RowsAffected), len(accounts))
		return types.DbErrFailToUpdateAccount
	}
	return nil
}

func (m *defaultAccountModel) GetCountByGreaterHeight(blockHeight int64) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height > ?", blockHeight).Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultAccountModel) GetMaxAccountIndex() (accountIndex int64, err error) {
	var result Account
	dbTx := m.DB.Table(m.table).Select("account_index").Order("account_index desc").Limit(1).Find(&result)
	if dbTx.Error != nil {
		return -1, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return -1, types.DbErrNotFound
	}
	return result.AccountIndex, nil
}

func (m *defaultAccountModel) GetByAccountIndexRange(fromAccountIndex int64, toAccountIndex int64) (accounts []*Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index >= ? and account_index <= ?", fromAccountIndex, toAccountIndex).Find(&accounts)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return accounts, nil
}
