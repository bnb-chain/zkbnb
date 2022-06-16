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
 */

package accounthistory

import (
	"errors"
	"fmt"

	table "github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type accountHistory struct {
	table     string
	db        *gorm.DB
	redisConn *redis.Redis
	cache     multcache.MultCache
}

/*
	Func: GetAccountByAccountName
	Params: accountName string
	Return: account Account, err error
	Description: get account info by account name
*/

func (m *accountHistory) GetAccountByAccountName(accountName string) (account *table.AccountHistory, err error) {
	dbTx := m.db.Table(m.table).Where("account_name = ?", accountName).Find(&account)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, nil
	}
	return account, nil
}

/*
	Func: GetAccountByAccountIndex
	Params: accountIndex int64
	Return: account Account, err error
	Description: get account info by index
*/

func (m *accountHistory) GetAccountByAccountIndex(accountIndex int64) (account *table.AccountHistory, err error) {
	// dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&account)
	// if dbTx.Error != nil {
	// 	err := fmt.Sprintf("[accountHistory.GetAccountByAccountIndex] %s", dbTx.Error)
	// 	logx.Error(err)
	// 	return nil, dbTx.Error
	// } else if dbTx.RowsAffected == 0 {
	// 	err := fmt.Sprintf("[accountHistory.GetAccountByAccountIndex] %s", ErrNotExistInSql)
	// 	logx.Error(err)
	// 	return nil, ErrNotExistInSql
	// }
	return nil, nil
}

/*
	Func: GetAccountByPk
	Params: pk string
	Return: account Account, err error
	Description: get account info by public key
*/

func (m *accountHistory) GetAccountByPk(pk string) (account *table.AccountHistory, err error) {
	dbTx := m.db.Table(m.table).Where("public_key = ?", pk).Find(&account)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[accountHistory.GetAccountByPk] %s", dbTx.Error)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[accountHistory.GetAccountByPk] %s", ErrNotExistInSql)
		logx.Error(err)
		return nil, ErrNotExistInSql
	}
	return account, nil
}

/*
	Func:  GetAccountAssetByIndex
	Params: accountIndex int64
	Return: accountAssets []*AccountAsset, err error
	Description: get account's asset info by accountIndex. This func is used for Account related api
*/
func (m *accountHistory) GetAccountAssetsByIndex(accountIndex int64) (accountAssets []*table.AccountHistory, err error) {
	dbTx := m.db.Table(m.table).Where("account_index = ?", accountIndex).Find(&accountAssets)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[asset.GetAccountAssetsByIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.GetAccountAssetsByIndex] %s", ErrNotExistInSql)
		logx.Error(err)
		return nil, ErrNotExistInSql
	}
	return accountAssets, nil
}
