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
 *
 */

package account

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/errorcode"
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
		GetAccountsList(limit int, offset int64) (accounts []*Account, err error)
		GetAccountsTotalCount() (count int64, err error)
	}

	defaultAccountModel struct {
		sqlc.CachedConn
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
		Status int
	}
)

func NewAccountModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AccountModel {
	return &defaultAccountModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      AccountTableName,
		DB:         db,
	}
}

func (*Account) TableName() string {
	return AccountTableName
}

/*
	Func: CreateAccountTable
	Params:
	Return: err error
	Description: create account table
*/
func (m *defaultAccountModel) CreateAccountTable() error {
	return m.DB.AutoMigrate(Account{})
}

/*
	Func: DropAccountTable
	Params:
	Return: err error
	Description: drop account table
*/
func (m *defaultAccountModel) DropAccountTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetAccountByIndex
	Params: accountIndex int64
	Return: account Account, err error
	Description: get account info by index
*/

func (m *defaultAccountModel) GetAccountByIndex(accountIndex int64) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&account)
	if dbTx.Error != nil {
		logx.Errorf("get account by index error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

/*
	Func: GetAccountByPk
	Params: pk string
	Return: account Account, err error
	Description: get account info by public key
*/

func (m *defaultAccountModel) GetAccountByPk(pk string) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("public_key = ?", pk).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByPk] %s", dbTx.Error.Error())
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByPk] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

/*
	Func: GetAccountByName
	Params: accountName string
	Return: account Account, err error
	Description: get account info by account name
*/

func (m *defaultAccountModel) GetAccountByName(accountName string) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name = ?", accountName).Find(&account)
	if dbTx.Error != nil {
		logx.Errorf("get account by account name error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountModel) GetAccountByNameHash(accountNameHash string) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name_hash = ?", accountNameHash).Find(&account)
	if dbTx.Error != nil {
		logx.Errorf("get account by account name hash error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
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
func (m *defaultAccountModel) GetAccountsList(limit int, offset int64) (accounts []*Account, err error) {
	dbTx := m.DB.Table(m.table).Limit(limit).Offset(int(offset)).Order("account_index desc").Find(&accounts)
	if dbTx.Error != nil {
		logx.Errorf("get account list error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
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
func (m *defaultAccountModel) GetAccountsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("get account count error, error: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultAccountModel) GetConfirmedAccountByIndex(accountIndex int64) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and status = ?", accountIndex, AccountStatusConfirmed).Find(&account)
	if dbTx.Error != nil {
		logx.Errorf("get confirmed account by account index error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}
