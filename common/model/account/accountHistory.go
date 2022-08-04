/*
 * Copyright © 2021 Zkbas Protocol
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
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	AccountHistoryModel interface {
		CreateAccountHistoryTable() error
		DropAccountHistoryTable() error
		IfAccountNameExist(name string) (bool, error)
		IfAccountExistsByAccountIndex(accountIndex int64) (bool, error)
		GetAccountsByBlockHeight(blockHeight int64) (accounts []*AccountHistory, err error)
		GetAccountByAccountIndex(accountIndex int64) (account *AccountHistory, err error)
		GetLatestAccountNonceByAccountIndex(accountIndex int64) (nonce int64, err error)
		GetAccountByPk(pk string) (account *AccountHistory, err error)
		GetAccountByAccountName(accountName string) (account *AccountHistory, err error)
		GetAccountByAccountNameHash(accountNameHash string) (account *AccountHistory, err error)
		GetAccountsList(limit int, offset int64) (accounts []*AccountHistory, err error)
		GetAccountsTotalCount() (count int64, err error)
		GetLatestAccountIndex() (accountIndex int64, err error)
		GetValidAccounts(height int64) (rowsAffected int64, accounts []*AccountHistory, err error)
		GetLatestAccountInfoByAccountIndex(accountIndex int64) (account *AccountHistory, err error)
	}

	defaultAccountHistoryModel struct {
		sqlc.CachedConn
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

func NewAccountHistoryModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AccountHistoryModel {
	return &defaultAccountHistoryModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      AccountHistoryTableName,
		DB:         db,
	}
}

func (*AccountHistory) TableName() string {
	return AccountHistoryTableName
}

/*
	Func: CreateAccountHistoryTable
	Params:
	Return: err error
	Description: create account history table
*/
func (m *defaultAccountHistoryModel) CreateAccountHistoryTable() error {
	return m.DB.AutoMigrate(AccountHistory{})
}

/*
	Func: DropAccountHistoryTable
	Params:
	Return: err error
	Description: drop account history table
*/
func (m *defaultAccountHistoryModel) DropAccountHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: IfAccountNameExist
	Params: name string
	Return: bool, error
	Description: check account name existence
*/
func (m *defaultAccountHistoryModel) IfAccountNameExist(name string) (bool, error) {
	var res int64
	dbTx := m.DB.Table(m.table).Where("account_name = ? and deleted_at is NULL", strings.ToLower(name)).Count(&res)

	if dbTx.Error != nil {
		err := fmt.Sprintf("[accountHistory.IfAccountNameExist] %s", dbTx.Error)
		logx.Error(err)
		return true, errors.New(err)
	} else if res == 0 {
		return false, nil
	} else if res != 1 {
		logx.Errorf("[accountHistory.IfAccountNameExist] %s", errorcode.DbErrDuplicatedAccountName)
		return true, errorcode.DbErrDuplicatedAccountName
	} else {
		return true, nil
	}
}

/*
	Func: IfAccountExistsByAccountIndex
	Params: accountIndex int64
	Return: bool, error
	Description: check account index existence
*/
func (m *defaultAccountHistoryModel) IfAccountExistsByAccountIndex(accountIndex int64) (bool, error) {
	var res int64
	dbTx := m.DB.Table(m.table).Where("account_index = ? and deleted_at is NULL", accountIndex).Count(&res)

	if dbTx.Error != nil {
		err := fmt.Sprintf("[accountHistory.IfAccountExistsByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return true, errors.New(err)
	} else if res == 0 {
		return false, nil
	} else if res != 1 {
		logx.Errorf("[accountHistory.IfAccountExistsByAccountIndex] %s", errorcode.DbErrDuplicatedAccountIndex)
		return true, errorcode.DbErrDuplicatedAccountIndex
	} else {
		return true, nil
	}
}

/*
	Func: GetAccountByAccountIndex
	Params: accountIndex int64
	Return: account Account, err error
	Description: get account info by index
*/

func (m *defaultAccountHistoryModel) GetAccountByAccountIndex(accountIndex int64) (account *AccountHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Order("l2_block_height desc").Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[accountHistory.GetAccountByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[accountHistory.GetAccountByAccountIndex] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountHistoryModel) GetLatestAccountNonceByAccountIndex(accountIndex int64) (nonce int64, err error) {
	var account *AccountHistory
	dbTx := m.DB.Table(m.table).Where("account_index = ? and nonce != -1", accountIndex).Order("l2_block_height desc").Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[accountHistory.GetAccountByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[accountHistory.GetAccountByAccountIndex] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return 0, errorcode.DbErrNotFound
	}
	return account.Nonce, nil
}

/*
	Func: GetAccountByPk
	Params: pk string
	Return: account Account, err error
	Description: get account info by public key
*/

func (m *defaultAccountHistoryModel) GetAccountByPk(pk string) (account *AccountHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("public_key = ?", pk).Find(&account)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountByPk] %s", dbTx.Error)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[accountHistory.GetAccountByPk] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

/*
	Func: GetAccountByAccountName
	Params: accountName string
	Return: account Account, err error
	Description: get account info by account name
*/

func (m *defaultAccountHistoryModel) GetAccountByAccountName(accountName string) (account *AccountHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name = ?", accountName).Find(&account)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountByAccountName] %s", dbTx.Error)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	} else if dbTx.RowsAffected == 0 {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountByAccountName] %s", errorcode.DbErrNotFound)
		logx.Info(errInfo)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

/*
	Func: GetAccountsList
	Params: limit int, offset int64
	Return: err error
	Description:  For API /api/v1/info/getAccountsList

*/
func (m *defaultAccountHistoryModel) GetAccountsList(limit int, offset int64) (accounts []*AccountHistory, err error) {
	dbTx := m.DB.Table(m.table).Limit(limit).Offset(int(offset)).Order("account_index desc").Find(&accounts)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountsList] %s", dbTx.Error)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[accountHistory.GetAccountsList] Get Accounts Error")
		return nil, errorcode.DbErrNotFound
	}
	return accounts, nil
}

/*
	Func: GetAccountsTotalCount
	Params:
	Return: count int64, err error
	Description: used for counting total accounts for explorer dashboard
*/
func (m *defaultAccountHistoryModel) GetAccountsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountsTotalCount] %s", dbTx.Error)
		logx.Error(errInfo)
		return 0, errors.New(errInfo)
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[accountHistory.GetAccountsTotalCount] No Accounts in Account Table")
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetLatestAccountIndex
	Params:
	Return: accountIndex int64, err error
	Description: get max accountIndex
*/
func (m *defaultAccountHistoryModel) GetLatestAccountIndex() (accountIndex int64, err error) {
	dbTx := m.DB.Table(m.table).Select("account_index").Order("account_index desc").Limit(1).Find(&accountIndex)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[accountHistory.GetLatestAccountIndex] %s", dbTx.Error)
		logx.Error(errInfo)
		return 0, errors.New(errInfo)
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[accountHistory.GetLatestAccountIndex] No Account in Account Table")
		return 0, errorcode.DbErrNotFound
	}
	logx.Info(accountIndex)
	return accountIndex, nil
}

/*
	Func: CreateNewAccount
	Params: nAccount *AccountHistory
	Return: err error
	Description:
*/
func (m *defaultAccountHistoryModel) CreateNewAccount(nAccount *AccountHistory) (err error) {
	dbTx := m.DB.Table(m.table).Create(&nAccount)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[accountHistory.CreateNewAccount] %s", dbTx.Error)
		logx.Error(errInfo)
		return errors.New(errInfo)
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[accountHistory.CreateNewAccount] Create nAccount no rows affected")
		return errors.New("[accountHistory.CreateNewAccount] Create nAccount no rows affected")
	}

	return nil
}

func (m *defaultAccountHistoryModel) GetValidAccounts(height int64) (rowsAffected int64, accounts []*AccountHistory, err error) {

	dbTx := m.DB.Table(m.table).
		Raw("SELECT a.* FROM account_history a WHERE NOT EXISTS"+
			"(SELECT * FROM account_history WHERE account_index = a.account_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height AND l2_block_height != -1) "+
			"AND l2_block_height <= ? AND l2_block_height != -1 ORDER BY account_index", height, height).
		Find(&accounts)
	if dbTx.Error != nil {
		logx.Errorf("[GetValidAccounts] unable to get related accounts: %s", dbTx.Error.Error())
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, accounts, nil

}

func (m *defaultAccountHistoryModel) GetLatestAccountInfoByAccountIndex(accountIndex int64) (account *AccountHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Order("l2_block_height desc").Find(&account)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] unable to get related account: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountHistoryModel) GetAccountByAccountNameHash(accountNameHash string) (account *AccountHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name_hash = ?", accountNameHash).Find(&account)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountByAccountName] %s", dbTx.Error)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	} else if dbTx.RowsAffected == 0 {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountByAccountName] %s", errorcode.DbErrNotFound)
		logx.Info(errInfo)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountHistoryModel) GetAccountsByBlockHeight(blockHeight int64) (accounts []*AccountHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height = ?", blockHeight).Find(&accounts)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountByAccountName] %s", dbTx.Error)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	} else if dbTx.RowsAffected == 0 {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountByAccountName] %s", errorcode.DbErrNotFound)
		logx.Info(errInfo)
		return nil, errorcode.DbErrNotFound
	}
	return accounts, nil
}
