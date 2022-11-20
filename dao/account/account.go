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
	"errors"
	"gorm.io/gorm"
	"strconv"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	AccountTableName = `account`
)

const (
	AccountStatusPending = iota
	AccountStatusConfirmed
)

const (
	AccountTransactionStatusProcessing = iota
	AccountTransactionStatusCommitted
)

type (
	AccountModel interface {
		CreateAccountTable() error
		DropAccountTable() error
		GetAccountByIndex(accountIndex int64) (account *Account, err error)
		GetConfirmedAccountByIndex(accountIndex int64) (account *Account, err error)
		GetAccountByPk(pk string) (account *Account, err error)
		GetAccountByName(name string) (account *Account, err error)
		GetAccountByNameHash(nameHash string) (account *Account, err error)
		GetAccounts(limit int, offset int64) (accounts []*Account, err error)
		GetAccountsTotalCount() (count int64, err error)
		UpdateAccountsInTransact(tx *gorm.DB, accounts []*Account) error
		UpdateAccountInTransact(account *Account) error
		UpdateAccountTransactionToCommitted(tx *gorm.DB, accounts []*Account) error
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
		AssetInfo string
		AssetRoot string
		// 0 - registered, not committer 1 - committer
		Status            int
		TransactionStatus int
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

func (m *defaultAccountModel) UpdateAccountInTransact(account *Account) error {
	const CreatedAt = "CreatedAt"
	dbTx := m.DB.Table(m.table).Where("account_index = ?", account.AccountIndex).
		Omit(CreatedAt).
		Select("*").
		Updates(&account)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		// this account is new, we need create first
		dbTx = m.DB.Table(m.table).Create(&account)
		if dbTx.Error != nil {
			return dbTx.Error
		}
	}
	return nil
}

func (m *defaultAccountModel) UpdateAccountTransactionToCommitted(tx *gorm.DB, accounts []*Account) error {
	length := len(accounts)
	if length == 0 {
		return nil
	}
	accountIndexes := make([]int64, 0, length)
	for _, account := range accounts {
		accountIndexes = append(accountIndexes, account.AccountIndex)
	}
	dbTx := tx.Model(&Account{}).Where("account_index in ? ", accountIndexes).Update("transaction_status", AccountTransactionStatusCommitted)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(length) {
		return errors.New("update accounts transaction status failed,rowsAffected =" + strconv.FormatInt(dbTx.RowsAffected, 10) + "not equal accounts length=" + strconv.Itoa(length))
	}
	return nil
}
