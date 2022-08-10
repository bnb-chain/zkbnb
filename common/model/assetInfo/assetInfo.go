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

package assetInfo

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	AssetInfoModel interface {
		CreateAssetInfoTable() error
		DropAssetInfoTable() error
		CreateAssetsInfoInBatches(l2AssetsInfo []*AssetInfo) (rowsAffected int64, err error)
		GetAssetsList() (res []*AssetInfo, err error)
		GetSimpleAssetInfoByAssetId(assetId int64) (res *AssetInfo, err error)
		GetAssetInfoBySymbol(symbol string) (res *AssetInfo, err error)
		GetAssetByAddress(address string) (info *AssetInfo, err error)
	}

	defaultAssetInfoModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	AssetInfo struct {
		gorm.Model
		AssetId     uint32 `gorm:"uniqueIndex"`
		AssetName   string
		AssetSymbol string
		L1Address   string
		Decimals    uint32
		Status      uint32
		IsGasAsset  uint32
	}
)

func (*AssetInfo) TableName() string {
	return AssetInfoTableName
}

func NewAssetInfoModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AssetInfoModel {
	return &defaultAssetInfoModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      AssetInfoTableName,
		DB:         db,
	}
}

/*
	Func: CreateL2AssetInfoTable
	Params:
	Return: err error
	Description: create l2 asset info table
*/
func (m *defaultAssetInfoModel) CreateAssetInfoTable() error {
	return m.DB.AutoMigrate(AssetInfo{})
}

/*
	Func: DropL2AssetInfoTable
	Params:
	Return: err error
	Description: drop l2 asset info table
*/
func (m *defaultAssetInfoModel) DropAssetInfoTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetL2AssetsList
	Params:
	Return: err error
	Description: create account table
*/
func (m *defaultAssetInfoModel) GetAssetsList() (res []*AssetInfo, err error) {
	dbTx := m.DB.Table(m.table).Find(&res)
	if dbTx.Error != nil {
		logx.Errorf("get assets error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return res, nil
}

/*
	Func: CreateL2AssetsInfoInBatches
	Params: []*L2AssetInfo
	Return: rowsAffected int64, err error
	Description: create L2AssetsInfo batches
*/
func (m *defaultAssetInfoModel) CreateAssetsInfoInBatches(l2AssetsInfo []*AssetInfo) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(l2AssetsInfo, len(l2AssetsInfo))
	if dbTx.Error != nil {
		logx.Errorf("create assets error, err: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

/*
	Func: GetSimpleL2AssetInfoByAssetId
	Params: assetId int64
	Return: L2AssetInfo, error
	Description: get layer-2 asset info by assetId
*/
func (m *defaultAssetInfoModel) GetSimpleAssetInfoByAssetId(assetId int64) (res *AssetInfo, err error) {
	dbTx := m.DB.Table(m.table).Where("asset_id = ?", assetId).Find(&res)
	if dbTx.Error != nil {
		logx.Errorf("get asset error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return res, nil
}

/*
	Func: GetL2AssetInfoBySymbol
	Params: symbol string
	Return: res *L2AssetInfo, err error
	Description: get l2 asset info by l2 symbol
*/
func (m *defaultAssetInfoModel) GetAssetInfoBySymbol(symbol string) (res *AssetInfo, err error) {
	dbTx := m.DB.Table(m.table).Where("asset_symbol = ?", symbol).Find(&res)
	if dbTx.Error != nil {
		logx.Errorf("get asset error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return res, nil
}

func (m *defaultAssetInfoModel) GetAssetByAddress(address string) (info *AssetInfo, err error) {
	dbTx := m.DB.Table(m.table).Where("asset_address = ?", address).Find(&info)
	if dbTx.Error != nil {
		logx.Errorf("fail to get asset by address: %s, error: %s", address, dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return info, nil
}
