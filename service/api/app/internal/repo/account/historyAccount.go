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

package account

import (
	"fmt"
	"log"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

type historyAccount struct {
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	redisConn  *redis.Redis
	cache      multcache.MultCache
}

/*
	Func: GetAccountsList
	Params: limit int, offset int64
	Return: err error
	Description:  For API /api/v1/info/getAccountsList

*/
func (m *historyAccount) GetAccountsList(limit int, offset int64) (accounts []*AccountHistoryInfo, err error) {
	cacheKeyAccountsList := fmt.Sprintf("cache:AccountsHistoryList_%v_%v", limit, offset)
	result, err := m.cache.GetWithSet(cacheKeyAccountsList, accounts,
		multcache.SqlBatchQuery, m.db, m.table, limit, offset, "account_index desc")
	if err != nil {
		return nil, err
	}
	accounts, ok := result.([]*AccountHistoryInfo)
	if !ok {
		log.Fatal("Error type!")
	}
	return accounts, nil
}

/*
	Func: GetAccountsTotalCount
	Params:
	Return: count int64, err error
	Description: used for counting total accounts for explorer dashboard
*/
func (m *historyAccount) GetAccountsTotalCount() (count int64, err error) {
	cacheKeyAccountsTotalCount := "cache:AccountsTotalCount"
	result, err := m.cache.GetWithSet(cacheKeyAccountsTotalCount, count,
		multcache.SqlQueryCount, m.db, m.table,
		"deleted_at is NULL")
	if err != nil {
		return 0, err
	}
	count, ok := result.(int64)
	if !ok {
		log.Fatal("Error type!")
	}
	return count, nil
}

/*
	Func: GetAccountByAccountName
	Params: accountName string
	Return: account Account, err error
	Description: get account info by account name
*/

func (m *historyAccount) GetAccountByAccountName(accountName string) (account *AccountHistoryInfo, err error) {
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

func (m *historyAccount) GetAccountByAccountIndex(accountIndex int64) (account *AccountHistoryInfo, err error) {
	// dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&account)
	// if dbTx.Error != nil {
	// 	err := fmt.Sprintf("[accountHistory.GetAccountByAccountIndex] %s", dbTx.Error)
	// 	logx.Error(err)
	// 	return nil, dbTx.Error
	// } else if dbTx.RowsAffected == 0 {
	// 	err := fmt.Sprintf("[accountHistory.GetAccountByAccountIndex] %s", ErrNotFound)
	// 	logx.Error(err)
	// 	return nil, ErrNotFound
	// }
	return nil, nil
}
