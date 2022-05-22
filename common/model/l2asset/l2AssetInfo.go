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

package l2asset

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheZecreyL2AssetInfoIdPrefix          = "cache:zecrey:l2AssetInfo:id:"
	cacheZecreyL2AssetInfoL2AssetIdPrefix   = "cache:zecrey:l2AssetInfo:l2AssetId:"
	cacheZecreyL2AssetInfoL2AssetNamePrefix = "cache:zecrey:l2AssetInfo:l2AssetName:"
)

type (
	L2AssetInfoModel interface {
		CreateL2AssetInfoTable() error
		DropL2AssetInfoTable() error
		CreateL2AssetInfo(l2AssetInfo *L2AssetInfo) (bool, error)
		CreateL2AssetsInfoInBatches(l2AssetsInfo []*L2AssetInfo) (rowsAffected int64, err error)
		GetL2AssetsCount() (latestHeight int64, err error)
		GetL2AssetsList() (res []*L2AssetInfo, err error)
		GetL2AssetsListWithoutL1AssetsInfo() (res []*L2AssetInfo, err error)
		GetSimpleL2AssetInfoByAssetId(assetId int64) (res *L2AssetInfo, err error)
		GetAssetIdCount() (res int64, err error)
		GetL2AssetInfoBySymbol(symbol string) (res *L2AssetInfo, err error)
	}

	defaultL2AssetInfoModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2AssetInfo struct {
		gorm.Model
		AssetId     int64 `gorm:"uniqueIndex"`
		AssetName   string
		AssetSymbol string
		Decimals    int64
		Status      int
	}
)

func (*L2AssetInfo) TableName() string {
	return L2AssetInfoTableName
}

func NewL2AssetInfoModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2AssetInfoModel {
	return &defaultL2AssetInfoModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      L2AssetInfoTableName,
		DB:         db,
	}
}

/*
	Func: CreateL2AssetInfoTable
	Params:
	Return: err error
	Description: create l2 asset info table
*/
func (m *defaultL2AssetInfoModel) CreateL2AssetInfoTable() error {
	return m.DB.AutoMigrate(L2AssetInfo{})
}

/*
	Func: DropL2AssetInfoTable
	Params:
	Return: err error
	Description: drop l2 asset info table
*/
func (m *defaultL2AssetInfoModel) DropL2AssetInfoTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetL2AssetsList
	Params:
	Return: err error
	Description: create account table
*/
func (m *defaultL2AssetInfoModel) GetL2AssetsList() (res []*L2AssetInfo, err error) {
	dbTx := m.DB.Table(m.table).Find(&res)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2asset.GetL2AssetsList] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2asset.GetL2AssetsList] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return res, nil
}

/*
	Func: GetL2AssetsListWithoutL1AssetsInfo
	Params:
	Return: err error
	Description: GetL2AssetsListWithoutL1AssetsInfo
*/
func (m *defaultL2AssetInfoModel) GetL2AssetsListWithoutL1AssetsInfo() (res []*L2AssetInfo, err error) {
	dbTx := m.DB.Table(m.table).Find(&res)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2asset.GetL2AssetsList] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2asset.GetL2AssetsList] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return res, nil
}

/*
	Func: CreateL2AssetInfo
	Params: l2AssetInfo *L2AssetInfo
	Return: bool, error
	Description: create L2AssetsInfo batches
*/
func (m *defaultL2AssetInfoModel) CreateL2AssetInfo(l2AssetInfo *L2AssetInfo) (bool, error) {
	dbTx := m.DB.Table(m.table).Create(l2AssetInfo)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2asset.CreateL2AssetInfo] %s", dbTx.Error)
		logx.Error(err)
		return false, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2asset.CreateL2AssetInfo] %s", ErrInvalidL2AssetInput)
		logx.Error(err)
		return false, ErrInvalidL2AssetInput
	}
	return true, nil
}

/*
	Func: CreateL2AssetsInfoInBatches
	Params: []*L2AssetInfo
	Return: rowsAffected int64, err error
	Description: create L2AssetsInfo batches
*/
func (m *defaultL2AssetInfoModel) CreateL2AssetsInfoInBatches(l2AssetsInfo []*L2AssetInfo) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(l2AssetsInfo, len(l2AssetsInfo))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2asset.CreateL2AssetsInfoInBatches] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

/*
	Func: GetL2AssetsCount
	Params:
	Return: latestHeight int64, err error
	Description: get latest l1asset id to active accounts
*/
func (m *defaultL2AssetInfoModel) GetL2AssetsCount() (latestHeight int64, err error) {
	var asset *L2AssetInfo
	dbTx := m.DB.Table(m.table).Order("l2_asset_id desc").First(&asset)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2asset.GetL2AssetsCount] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2asset.GetL2AssetsCount] %s", ErrNotFound)
		logx.Error(err)
		return 0, ErrNotFound
	}
	return asset.AssetId + 1, nil
}

/*
	Func: GetSimpleL2AssetInfoByAssetId
	Params: assetId int64
	Return: L2AssetInfo, error
	Description: get layer-2 asset info by assetId
*/
func (m *defaultL2AssetInfoModel) GetSimpleL2AssetInfoByAssetId(assetId int64) (res *L2AssetInfo, err error) {
	dbTx := m.DB.Table(m.table).Where("asset_id = ?", assetId).Find(&res)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[l2asset.GetL2AssetInfoByAssetId] %s", dbTx.Error)
		logx.Error(errInfo)
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		errInfo := fmt.Sprintf("[l2asset.GetL2AssetInfoByAssetId] %s", ErrNotFound)
		logx.Error(errInfo)
		return nil, ErrNotFound
	}
	return res, nil
}

/*
	Func: GetAssetIdCount
	Params:
	Return: res int64, err error
	Description: get l2 asset id count
*/
func (m *defaultL2AssetInfoModel) GetAssetIdCount() (res int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&res)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[l2asset.GetAssetIdCount] %s", dbTx.Error)
		logx.Error(errInfo)
		// TODO : to be modified
		return 0, dbTx.Error
	} else {
		return res, nil
	}
}

/*
	Func: GetL2AssetInfoBySymbol
	Params: symbol string
	Return: res *L2AssetInfo, err error
	Description: get l2 asset info by l2 symbol
*/
func (m *defaultL2AssetInfoModel) GetL2AssetInfoBySymbol(symbol string) (res *L2AssetInfo, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_symbol = ?", symbol).Find(&res)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[l2asset.GetL2AssetInfoBySymbol] %s", dbTx.Error)
		logx.Error(errInfo)
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		errInfo := fmt.Sprintf("[l2asset.GetL2AssetInfoBySymbol] %s", ErrNotFound)
		logx.Error(errInfo)
		return nil, ErrNotFound
	}
	return res, nil
}
