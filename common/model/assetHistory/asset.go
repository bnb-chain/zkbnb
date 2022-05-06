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

package assetHistory

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheAccountAssetHistoryIdPrefix = "cache::AccountAssetHistory:id:"
)

type (
	AccountAssetHistoryModel interface {
		CreateAccountAssetHistoryTable() error
		DropAccountAssetHistoryTable() error
		CreateAccountAssetHistory(AccountAssetHistory *AccountAssetHistory) error
		CreateAccountAssetHistorysInBatches(AccountAssetHistorys []*AccountAssetHistory) error
		GetAccountAssetHistorysByIndex(accountIndex int64) (AccountAssetHistorys []*AccountAssetHistory, err error)
		GetSingleAccountAssetHistory(accountIndex int64, assetId int64) (AccountAssetHistory *AccountAssetHistory, err error)
		UpdateAccountAssetHistory(AccountAssetHistory *AccountAssetHistory) error
		UpdateAccountAssetHistorys(AccountAssetHistorys []*AccountAssetHistory) error
		GetAllAccountAssetHistorys() (AccountAssetHistory []*AccountAssetHistory, err error)
		GetLatestAccountAssetsByBlockHeight(height int64) (
			rowsAffected int64, accountAssets []*AccountAssetHistory, err error,
		)
		GetLatestAccountAssetByIndexAndAssetId(
			accountIndex int64, assetId int64,
		) (
			rowsAffected int64, accountAsset *AccountAssetHistory, err error,
		)
		GetAccountAssetsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, accountAssets []AccountAssetHistory, err error)
	}

	defaultAccountAssetHistoryModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	AccountAssetHistory struct {
		gorm.Model
		AccountIndex  int64 `gorm:"index"`
		AssetId       int64 `gorm:"index"`
		Balance       string
		L2BlockHeight int64
	}
)

func (*AccountAssetHistory) TableName() string {
	return GeneralAssetHistoryTable
}

func NewAccountAssetHistoryModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AccountAssetHistoryModel {
	return &defaultAccountAssetHistoryModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      GeneralAssetHistoryTable,
		DB:         db,
	}
}

/*
	Func: CreateAccountAssetHistoryTable
	Params:
	Return: err error
	Description: create account asset table
*/
func (m *defaultAccountAssetHistoryModel) CreateAccountAssetHistoryTable() error {
	return m.DB.AutoMigrate(AccountAssetHistory{})
}

/*
	Func: DropAccountAssetHistoryTable
	Params:
	Return: err error
	Description: drop account asset table
*/
func (m *defaultAccountAssetHistoryModel) DropAccountAssetHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func:  IsAccountAssetHistoryExist
	Params: accountIndex uint32, assetId uint32
	Return: bool, error
	Description: Check whether AccountAssetHistory exists by account index and asset id.
				 This is used for Create func in this module
*/
func (m *defaultAccountAssetHistoryModel) isAccountAssetHistoryExist(accountIndex int64, assetId int64) (bool, error) {
	// todo cache optimization
	var count int64
	dbTx := m.DB.Table(m.table).Where("account_index = ? and asset_id = ? and deleted_at is NULL", accountIndex, assetId).Count(&count)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[asset.isAccountAssetHistoryExist] %s", dbTx.Error)
		logx.Error(err)
		return true, dbTx.Error
	}
	return count != 0, nil
}

/*
	Func:  CreateAccountAssetHistory
	Params: AccountAssetHistory *AccountAssetHistory
	Return: error
	Description: create account related l1asset
*/
func (m *defaultAccountAssetHistoryModel) CreateAccountAssetHistory(AccountAssetHistory *AccountAssetHistory) error {
	isExist, err := m.isAccountAssetHistoryExist(AccountAssetHistory.AccountIndex, AccountAssetHistory.AssetId)
	if err != nil {
		res := fmt.Sprintf("[asset.CreateAccountAssetHistory] %s", err)
		logx.Error(res)
		return err
	}
	if isExist {
		err := fmt.Sprintf("[asset.CreateAccountAssetHistory] %s", ErrAccountExist)
		logx.Error(err)
		return ErrAccountExist
	}
	dbTx := m.DB.Table(m.table).Create(AccountAssetHistory)
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.CreateAccountAssetHistory] %s", ErrInvalidAccountAssetHistoryInput)
		logx.Error(err)
		return ErrInvalidAccountAssetHistoryInput
	}
	return dbTx.Error
}

/*
	Func:  CreateAccountAssetHistorys
	Params: AccountAssetHistorys []*AccountAssetHistory
	Return: error
	Description: create account related assets
*/
func (m *defaultAccountAssetHistoryModel) CreateAccountAssetHistorysInBatches(AccountAssetHistorys []*AccountAssetHistory) error {
	for _, AccountAssetHistory := range AccountAssetHistorys {
		isExist, err := m.isAccountAssetHistoryExist(AccountAssetHistory.AccountIndex, AccountAssetHistory.AssetId)
		if err != nil {
			res := fmt.Sprintf("[asset.CreateAccountAssetHistorysInBatches] %s", err)
			logx.Error(res)
			return err
		}
		if isExist {
			return ErrAccountExist
		}
	}
	dbTx := m.DB.Table(m.table).CreateInBatches(AccountAssetHistorys, len(AccountAssetHistorys))
	if dbTx.RowsAffected == 0 {
		res := fmt.Sprintf("[asset.CreateAccountAssetHistorysInBatches] %s", ErrInvalidAccountAssetHistoryInput)
		logx.Error(res)
		return ErrInvalidAccountAssetHistoryInput
	}
	return dbTx.Error
}

/*
	Func:  GetAccountAssetHistoryByIndex
	Params: accountIndex int64
	Return: AccountAssetHistorys []*AccountAssetHistory, err error
	Description: get account's asset info by accountIndex. This func is used for Account related api
*/
func (m *defaultAccountAssetHistoryModel) GetAccountAssetHistorysByIndex(accountIndex int64) (AccountAssetHistorys []*AccountAssetHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&AccountAssetHistorys)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[asset.GetAccountAssetHistorysByIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.GetAccountAssetHistorysByIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return AccountAssetHistorys, nil
}

/*
	Func: GetSingleAccountAssetHistory
	Params: accountIndex int64, assetId int64
	Return: AccountAssetHistory *AccountAssetHistory, err error
	Description: get single account's asset info by accountIndex and assetId. This func is used for Account related api
*/
func (m *defaultAccountAssetHistoryModel) GetSingleAccountAssetHistory(accountIndex int64, assetId int64) (AccountAssetHistory *AccountAssetHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and asset_id = ?", accountIndex, assetId).Find(&AccountAssetHistory)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[asset.GetSingleAccountAssetHistory] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.GetSingleAccountAssetHistory] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return AccountAssetHistory, nil
}

/*
	Func: UpdateAccountAssetHistory
	Params: AccountAssetHistory *AccountAssetHistory
	Return: error
	Description: update account asset
*/
func (m *defaultAccountAssetHistoryModel) UpdateAccountAssetHistory(AccountAssetHistory *AccountAssetHistory) error {
	dbTx := m.DB.Table(m.table).Where("id = ?", AccountAssetHistory.ID).
		Select("*").
		Updates(AccountAssetHistory)
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.UpdateAccountAssetHistory] %s", dbTx.Error)
		logx.Error(err)
		return ErrInvalidAccountAssetHistoryInput
	}
	return dbTx.Error
}

/*
	Func: UpdateAccountAssetHistorys
	Params: AccountAssetHistorys []*AccountAssetHistory
	Return: error
	Description: update account assets
*/
func (m *defaultAccountAssetHistoryModel) UpdateAccountAssetHistorys(AccountAssetHistorys []*AccountAssetHistory) error {
	err := m.DB.Table(m.table).Transaction(
		func(tx *gorm.DB) error { // transact
			for _, AccountAssetHistory := range AccountAssetHistorys {
				dbTx := m.DB.Table(m.table).Where("id = ?", AccountAssetHistory.ID).
					Select("*").
					Updates(AccountAssetHistory)
				if dbTx.Error != nil {
					err := fmt.Sprintf("[asset.UpdateAccountAssetHistorys] %s", dbTx.Error)
					logx.Error(err)
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					AccountAssetHistoryInfo, err := json.Marshal(AccountAssetHistory)
					if err != nil {
						res := fmt.Sprintf("[asset.UpdateAccountAssetHistorys] %s", err)
						logx.Error(res)
						return err
					}
					logx.Error("[asset.UpdateAccountAssetHistorys] %s" + "Invalid AccountAssetHistory:  " + string(AccountAssetHistoryInfo))
					return errors.New("Invalid AccountAssetHistory:  " + string(AccountAssetHistoryInfo))
				}
			}
			return nil
		})
	return err
}

/*
	Func: GetAllAccountAssetHistorys
	Params:
	Return: AccountAssetHistory *AccountAssetHistory, err error
	Description: Used for construct MPT
*/
func (m *defaultAccountAssetHistoryModel) GetAllAccountAssetHistorys() (AccountAssetHistory []*AccountAssetHistory, err error) {
	dbTx := m.DB.Table(m.table).Order("account_index, asset_id").Find(&AccountAssetHistory)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[asset.GetAllAccountAssetHistorys] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.GetAllAccountAssetHistorys] %s", ErrNotFound)
		logx.Error(err)
		return AccountAssetHistory, ErrNotFound
	}
	return AccountAssetHistory, nil
}

func (m *defaultAccountAssetHistoryModel) GetLatestAccountAssetsByBlockHeight(height int64) (
	rowsAffected int64, accountAssets []*AccountAssetHistory, err error,
) {
	dbTx := m.DB.Table(m.table).
		Raw("SELECT a.* FROM account_asset_history a WHERE NOT EXISTS"+
			"(SELECT * FROM account_asset_history WHERE account_index = a.account_index AND asset_id = a.asset_id AND l2_block_height <= ? AND l2_block_height > a.l2_block_height) "+
			"AND l2_block_height <= ? ORDER BY account_index, asset_id", height, height).
		Find(&accountAssets)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestAccountAssetsByBlockHeight] unable to get related assets: %s", dbTx.Error.Error())
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, accountAssets, nil
}

func (m *defaultAccountAssetHistoryModel) GetLatestAccountAssetByIndexAndAssetId(
	accountIndex int64, assetId int64,
) (
	rowsAffected int64, accountAsset *AccountAssetHistory, err error,
) {
	// todo debug
	dbTx := m.DB.Table(m.table).Where("account_index = ? AND asset_id = ? AND l2_block_height >= 0", accountIndex, assetId).Order("l2_block_height desc").Find(&accountAsset)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestAccountAssetByIndexAndAssetId] unable to get related assets: %s", dbTx.Error.Error())
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, accountAsset, nil
}

func (m *defaultAccountAssetHistoryModel) GetAccountAssetsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, accountAssets []AccountAssetHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height = ?", l2BlockHeight).Order("account_index, asset_id").Find(&accountAssets)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[GetAccountAssetsByBlockHeight] unable to get related txs: %s", err.Error())
		logx.Error(errInfo)
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, accountAssets, nil
}
