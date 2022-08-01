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
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"gorm.io/gorm"

	table "github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) GetBasicAccountByAccountName(ctx context.Context, accountName string) (*table.Account, error) {
	f := func() (interface{}, error) {
		account := &table.Account{}
		dbTx := m.db.Table(m.table).Where("account_name = ?", accountName).Find(&account)
		if dbTx.Error != nil {
			logx.Errorf("fail to get account by name: %s, error: %s", accountName, dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		return account, nil
	}
	account := &table.Account{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountByAccountName(accountName), account, 10, f)
	if err != nil {
		return nil, err
	}
	account, _ = value.(*table.Account)
	return account, nil
}

func (m *model) GetBasicAccountByAccountPk(ctx context.Context, accountPk string) (*table.Account, error) {
	f := func() (interface{}, error) {
		account := &table.Account{}
		dbTx := m.db.Table(m.table).Where("public_key = ?", accountPk).Find(&account)
		if dbTx.Error != nil {
			logx.Errorf("fail to get account by pk: %s, error: %s", accountPk, dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		return account, nil
	}
	account := &table.Account{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountByAccountPk(accountPk), account, 10, f)
	if err != nil {
		return nil, err
	}
	account, _ = value.(*table.Account)
	return account, nil
}

/*
	Func: GetAccountByAccountIndex
	Params: accountIndex int64
	Return: account Account, err error
	Description: get account info by index
*/

func (m *model) GetAccountByAccountIndex(accountIndex int64) (account *table.Account, err error) {
	dbTx := m.db.Table(m.table).Where("account_index = ?", accountIndex).Find(&account)
	if dbTx.Error != nil {
		logx.Errorf("fail to get tx by account: %d, error: %s", accountIndex, dbTx.Error.Error())
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

func (m *model) GetAccountByPk(pk string) (account *table.Account, err error) {
	dbTx := m.db.Table(m.table).Where("public_key = ?", pk).Find(&account)
	if dbTx.Error != nil {
		logx.Errorf("fail to get tx by pk: %s, error: %s", pk, dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
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
func (m *model) GetAccountByAccountName(ctx context.Context, accountName string) (*table.Account, error) {
	account := &table.Account{}
	dbTx := m.db.Table(m.table).Where("account_name = ?", accountName).Find(&account)
	if dbTx.Error != nil {
		logx.Errorf("fail to get tx by account: %s, error: %s", accountName, dbTx.Error.Error())
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
func (m *model) GetAccountsList(limit int, offset int64) (accounts []*table.Account, err error) {
	dbTx := m.db.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("account_index desc").Find(&accounts)
	if dbTx.Error != nil {
		logx.Errorf("fail to get accounts, offset: %d, limit: %d, error: %s", offset, limit, dbTx.Error.Error())
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
func (m *model) GetAccountsTotalCount() (count int64, err error) {
	dbTx := m.db.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}
