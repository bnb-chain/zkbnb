/*
 * Copyright © 2021 Zecrey Protocol
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
	"context"
	"fmt"
	"strings"

	table "github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type accountModel struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

/*
	Func: IfAccountNameExist
	Params: name string
	Return: bool, error
	Description: check account name existence
*/
func (m *accountModel) IfAccountNameExist(name string) (bool, error) {
	var res int64
	dbTx := m.db.Table(m.table).Where("account_name = ? and deleted_at is NULL", strings.ToLower(name)).Count(&res)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.IfAccountNameExist] %s", dbTx.Error)
		logx.Error(err)
		return true, dbTx.Error
	} else if res == 0 {
		return false, nil
	} else if res != 1 {
		logx.Errorf("[account.IfAccountNameExist] %s", ErrDuplicatedAccountName)
		return true, ErrDuplicatedAccountName
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
func (m *accountModel) IfAccountExistsByAccountIndex(accountIndex int64) (bool, error) {
	var res int64
	dbTx := m.db.Table(m.table).Where("account_index = ? and deleted_at is NULL", accountIndex).Count(&res)

	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.IfAccountExistsByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		// TODO : to be modified
		return true, dbTx.Error
	} else if res == 0 {
		return false, nil
	} else if res != 1 {
		logx.Errorf("[account.IfAccountExistsByAccountIndex] %s", ErrDuplicatedAccountIndex)
		return true, ErrDuplicatedAccountIndex
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

func (m *accountModel) GetAccountByAccountIndex(accountIndex int64) (account *table.Account, err error) {
	dbTx := m.db.Table(m.table).Where("account_index = ?", accountIndex).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return account, nil
}

func (m *accountModel) GetVerifiedAccountByAccountIndex(accountIndex int64) (account *table.Account, err error) {
	dbTx := m.db.Table(m.table).Where("account_index = ? and status = ?", accountIndex, AccountStatusVerified).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return account, nil
}

/*
	Func: GetAccountByPk
	Params: pk string
	Return: account Account, err error
	Description: get account info by public key
*/

func (m *accountModel) GetAccountByPk(pk string) (account *table.Account, err error) {
	dbTx := m.db.Table(m.table).Where("public_key = ?", pk).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByPk] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByPk] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return account, nil
}

/*
	Func: GetAccountByAccountName
	Params: accountName string
	Return: account Account, err error
	Description: get account info by account name
*/
func (m *accountModel) GetAccountByAccountName(ctx context.Context, accountName string) (*table.Account, error) {
	f := func() (interface{}, error) {
		account := &table.Account{}
		dbTx := m.db.Table(m.table).Where("account_name = ?", accountName).Find(&account)
		if dbTx.Error != nil {
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			return nil, ErrNotFound
		}
		return account, nil
	}
	account := &table.Account{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyAccountAccountName, account, 10, f)
	if err != nil {
		return nil, err
	}
	account, _ = value.(*table.Account)
	return account, nil
}

/*
	Func: GetAccountsList
	Params: limit int, offset int64
	Return: err error
	Description:  For API /api/v1/info/getAccountsList

*/
func (m *accountModel) GetAccountsList(limit int, offset int64) (accounts []*table.Account, err error) {
	dbTx := m.db.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("account_index desc").Find(&accounts)
	if dbTx.Error != nil {
		logx.Error("[account.GetAccountsList] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[account.GetAccountsList] Get Accounts Error")
		return nil, ErrNotFound
	}
	return accounts, nil
}

/*
	Func: GetAccountsTotalCount
	Params:
	Return: count int64, err error
	Description: used for counting total accounts for explorer dashboard
*/
func (m *accountModel) GetAccountsTotalCount() (count int64, err error) {
	dbTx := m.db.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		logx.Error("[account.GetAccountsTotalCount] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[account.GetAccountsTotalCount] No Accounts in Account Table")
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetAllAccounts
	Params:
	Return: count int64, err error
	Description: used for construct MPT
*/
func (m *accountModel) GetAllAccounts() (accounts []*table.Account, err error) {
	dbTx := m.db.Table(m.table).Order("account_index").Find(&accounts)
	if dbTx.Error != nil {
		logx.Error("[account.GetAllAccounts] %s", dbTx.Error)
		return accounts, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[account.GetAllAccounts] No Account in Account Table")
		return accounts, nil
	}
	return accounts, nil
}

/*
	Func: GetLatestAccountIndex
	Params:
	Return: accountIndex int64, err error
	Description: get max accountIndex
*/
func (m *accountModel) GetLatestAccountIndex() (accountIndex int64, err error) {
	dbTx := m.db.Table(m.table).Select("account_index").Order("account_index desc").Limit(1).Find(&accountIndex)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetLatestAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[account.GetLatestAccountIndex] No Account in Account Table")
		return 0, ErrNotFound
	}
	logx.Info(accountIndex)
	return accountIndex, nil
}

func (m *accountModel) GetAccountByAccountNameHash(accountNameHash string) (account *table.Account, err error) {
	dbTx := m.db.Table(m.table).Where("account_name_hash = ?", accountNameHash).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByAccountNameHash] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByAccountNameHash] %s", ErrNotFound)
		logx.Info(err)
		return nil, ErrNotFound
	}
	return account, nil
}

func (m *accountModel) GetConfirmedAccounts() (accounts []*table.Account, err error) {
	dbTx := m.db.Table(m.table).Where("status = ?", AccountStatusConfirmed).Order("account_index").Find(&accounts)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetConfirmedAccounts] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetConfirmedAccounts] %s", ErrNotFound)
		logx.Info(err)
		return nil, ErrNotFound
	}
	return accounts, nil
}

func (m *accountModel) GetConfirmedAccountByAccountIndex(accountIndex int64) (account *table.Account, err error) {
	dbTx := m.db.Table(m.table).Where("account_index = ? and status = ?", accountIndex, AccountStatusConfirmed).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return account, nil
}
