/*
 * Copyright © 2021 ZkBAS Protocol
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

	"github.com/bnb-chain/zkbas/types"
)

type (
	AccountHistoryModel interface {
		CreateAccountHistoryTable() error
		DropAccountHistoryTable() error
		GetValidAccounts(height int64, limit int, offset int) (rowsAffected int64, accounts []*AccountHistory, err error)
		GetValidAccountCount(height int64) (accounts int64, err error)
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

func (m *defaultAccountHistoryModel) CreateAccountHistoryTable() error {
	return m.DB.AutoMigrate(AccountHistory{})
}

func (m *defaultAccountHistoryModel) DropAccountHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultAccountHistoryModel) CreateNewAccount(nAccount *AccountHistory) (err error) {
	dbTx := m.DB.Table(m.table).Create(&nAccount)
	if dbTx.Error != nil {
		logx.Errorf("create new account error, err: %s", dbTx.Error.Error())
		return types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return errors.New("create new account no rows affected")
	}

	return nil
}

func (m *defaultAccountHistoryModel) GetValidAccounts(height int64, limit int, offset int) (rowsAffected int64, accounts []*AccountHistory, err error) {
	subQuery := m.DB.Table(m.table).Select("*").
		Where("account_index = a.account_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height AND l2_block_height != -1", height)

	dbTx := m.DB.Table(m.table+" as a").Select("*").
		Where("NOT EXISTS (?) AND l2_block_height <= ? AND l2_block_height != -1", subQuery, height).
		Limit(limit).Offset(offset).
		Order("account_index")

	if dbTx.Find(&accounts).Error != nil {
		logx.Errorf("[GetValidAccounts] unable to get related accounts: %s", dbTx.Error.Error())
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
		logx.Errorf("[GetValidAccountCount] unable to get related accounts: %s", dbTx.Error.Error())
		return 0, types.DbErrSqlOperation
	}
	return count, nil
}
