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

package asset

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
	cacheAccountAssetIdPrefix = "cache::accountAsset:id:"
)

type (
	AccountAssetModel interface {
		CreateAccountAssetTable() error
		DropAccountAssetTable() error
		CreateAccountAsset(accountAsset *AccountAsset) error
		CreateAccountAssetsInBatches(accountAssets []*AccountAsset) error
		GetAccountAssetsByIndex(accountIndex int64) (accountAssets []*AccountAsset, err error)
		GetSingleAccountAsset(accountIndex int64, assetId int64) (accountAsset *AccountAsset, err error)
		UpdateAccountAsset(accountAsset *AccountAsset) error
		UpdateAccountAssets(accountAssets []*AccountAsset) error
		GetAllAccountAssets() (accountAsset []*AccountAsset, err error)
	}

	defaultAccountAssetModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	AccountAsset struct {
		gorm.Model
		AccountIndex int64 `gorm:"index"`
		AssetId      int64 `gorm:"index"`
		Balance      string
	}
)

func (*AccountAsset) TableName() string {
	return GeneralAssetTable
}

func NewAccountAssetModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AccountAssetModel {
	return &defaultAccountAssetModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      GeneralAssetTable,
		DB:         db,
	}
}

/*
	Func: CreateAccountAssetTable
	Params:
	Return: err error
	Description: create account asset table
*/
func (m *defaultAccountAssetModel) CreateAccountAssetTable() error {
	return m.DB.AutoMigrate(AccountAsset{})
}

/*
	Func: DropAccountAssetTable
	Params:
	Return: err error
	Description: drop account asset table
*/
func (m *defaultAccountAssetModel) DropAccountAssetTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func:  IsAccountAssetExist
	Params: accountIndex uint32, assetId uint32
	Return: bool, error
	Description: Check whether accountasset exists by account index and asset id.
				 This is used for Create func in this module
*/
func (m *defaultAccountAssetModel) isAccountAssetExist(accountIndex int64, assetId int64) (bool, error) {
	// todo cache optimization
	var count int64
	dbTx := m.DB.Table(m.table).Where("account_index = ? and asset_id = ? and deleted_at is NULL", accountIndex, assetId).Count(&count)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[asset.isAccountAssetExist] %s", dbTx.Error)
		logx.Error(err)
		return true, dbTx.Error
	}
	return count != 0, nil
}

/*
	Func:  CreateAccountAsset
	Params: accountAsset *AccountAsset
	Return: error
	Description: create account related l1asset
*/
func (m *defaultAccountAssetModel) CreateAccountAsset(accountAsset *AccountAsset) error {
	isExist, err := m.isAccountAssetExist(accountAsset.AccountIndex, accountAsset.AssetId)
	if err != nil {
		res := fmt.Sprintf("[asset.CreateAccountAsset] %s", err)
		logx.Error(res)
		return err
	}
	if isExist {
		err := fmt.Sprintf("[asset.CreateAccountAsset] %s", ErrAccountExist)
		logx.Error(err)
		return ErrAccountExist
	}
	dbTx := m.DB.Table(m.table).Create(accountAsset)
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.CreateAccountAsset] %s", ErrInvalidAccountAssetInput)
		logx.Error(err)
		return ErrInvalidAccountAssetInput
	}
	return dbTx.Error
}

/*
	Func:  CreateAccountAssets
	Params: accountAssets []*AccountAsset
	Return: error
	Description: create account related assets
*/
func (m *defaultAccountAssetModel) CreateAccountAssetsInBatches(accountAssets []*AccountAsset) error {
	for _, accountAsset := range accountAssets {
		isExist, err := m.isAccountAssetExist(accountAsset.AccountIndex, accountAsset.AssetId)
		if err != nil {
			res := fmt.Sprintf("[asset.CreateAccountAssetsInBatches] %s", err)
			logx.Error(res)
			return err
		}
		if isExist {
			return ErrAccountExist
		}
	}
	dbTx := m.DB.Table(m.table).CreateInBatches(accountAssets, len(accountAssets))
	if dbTx.RowsAffected == 0 {
		res := fmt.Sprintf("[asset.CreateAccountAssetsInBatches] %s", ErrInvalidAccountAssetInput)
		logx.Error(res)
		return ErrInvalidAccountAssetInput
	}
	return dbTx.Error
}

/*
	Func:  GetAccountAssetByIndex
	Params: accountIndex int64
	Return: accountAssets []*AccountAsset, err error
	Description: get account's asset info by accountIndex. This func is used for Account related api
*/
func (m *defaultAccountAssetModel) GetAccountAssetsByIndex(accountIndex int64) (accountAssets []*AccountAsset, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&accountAssets)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[asset.GetAccountAssetsByIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.GetAccountAssetsByIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return accountAssets, nil
}

/*
	Func: GetSingleAccountAsset
	Params: accountIndex int64, assetId int64
	Return: accountAsset *AccountAsset, err error
	Description: get single account's asset info by accountIndex and assetId. This func is used for Account related api
*/
func (m *defaultAccountAssetModel) GetSingleAccountAsset(accountIndex int64, assetId int64) (accountAsset *AccountAsset, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and asset_id = ?", accountIndex, assetId).Find(&accountAsset)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[asset.GetSingleAccountAsset] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.GetSingleAccountAsset] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return accountAsset, nil
}

/*
	Func: UpdateAccountAsset
	Params: accountAsset *AccountAsset
	Return: error
	Description: update account asset
*/
func (m *defaultAccountAssetModel) UpdateAccountAsset(accountAsset *AccountAsset) error {
	dbTx := m.DB.Table(m.table).Where("id = ?", accountAsset.ID).
		Select("*").
		Updates(accountAsset)
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.UpdateAccountAsset] %s", dbTx.Error)
		logx.Error(err)
		return ErrInvalidAccountAssetInput
	}
	return dbTx.Error
}

/*
	Func: UpdateAccountAssets
	Params: accountAssets []*AccountAsset
	Return: error
	Description: update account assets
*/
func (m *defaultAccountAssetModel) UpdateAccountAssets(accountAssets []*AccountAsset) error {
	err := m.DB.Table(m.table).Transaction(
		func(tx *gorm.DB) error { // transact
			for _, accountAsset := range accountAssets {
				dbTx := m.DB.Table(m.table).Where("id = ?", accountAsset.ID).
					Select("*").
					Updates(accountAsset)
				if dbTx.Error != nil {
					err := fmt.Sprintf("[asset.UpdateAccountAssets] %s", dbTx.Error)
					logx.Error(err)
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					accountAssetInfo, err := json.Marshal(accountAsset)
					if err != nil {
						res := fmt.Sprintf("[asset.UpdateAccountAssets] %s", err)
						logx.Error(res)
						return err
					}
					logx.Error("[asset.UpdateAccountAssets] %s" + "Invalid accountAsset:  " + string(accountAssetInfo))
					return errors.New("Invalid accountAsset:  " + string(accountAssetInfo))
				}
			}
			return nil
		})
	return err
}

/*
	Func: GetAllAccountAssets
	Params:
	Return: accountAsset *AccountAsset, err error
	Description: Used for construct MPT
*/
func (m *defaultAccountAssetModel) GetAllAccountAssets() (accountAsset []*AccountAsset, err error) {
	dbTx := m.DB.Table(m.table).Order("account_index, asset_id").Find(&accountAsset)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[asset.GetAllAccountAssets] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[asset.GetAllAccountAssets] %s", ErrNotFound)
		logx.Error(err)
		return accountAsset, ErrNotFound
	}
	return accountAsset, nil
}
