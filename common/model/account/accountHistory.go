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
		GetValidAccounts(height int64) (rowsAffected int64, accounts []*AccountHistory, err error)
		GetValidAccountNums(height int64) (accounts int64, err error)
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
	Func: CreateNewAccount
	Params: nAccount *AccountHistory
	Return: err error
	Description:
*/
func (m *defaultAccountHistoryModel) CreateNewAccount(nAccount *AccountHistory) (err error) {
	dbTx := m.DB.Table(m.table).Create(&nAccount)
	if dbTx.Error != nil {
		logx.Error("create new account error, err: %s", dbTx.Error.Error())
		return errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return errors.New("create new account no rows affected")
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
		logx.Errorf("unable to get related accounts: %s", dbTx.Error.Error())
		return 0, nil, errorcode.DbErrSqlOperation
	}
	return dbTx.RowsAffected, accounts, nil

}

type countResult struct {
	Count int `json:"count"`
}

func (m *defaultAccountHistoryModel) GetValidAccountNums(height int64) (accounts int64, err error) {
	var countResult countResult
	dbTx := m.DB.Table(m.table).
		Raw("SELECT count(a.*) FROM account_history a WHERE NOT EXISTS"+
			"(SELECT * FROM account_history WHERE account_index = a.account_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height AND l2_block_height != -1) "+
			"AND l2_block_height <= ? AND l2_block_height != -1", height, height).
		Scan(&countResult)
	if dbTx.Error != nil {
		logx.Errorf("unable to get related accounts: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	}
	return int64(countResult.Count), nil
}
