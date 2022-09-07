/*
 * Copyright © 2021 ZkBNB Protocol
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
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	AssetTableName = `asset`

	StatusActive   uint32 = 0
	StatusInactive uint32 = 1

	IsGasAsset = 1
)

type (
	AssetModel interface {
		CreateAssetTable() error
		DropAssetTable() error
		CreateAssets(assets []*Asset) (rowsAffected int64, err error)
		GetAssetsTotalCount() (count int64, err error)
		GetAssets(limit int64, offset int64) (assets []*Asset, err error)
		GetAssetById(assetId int64) (asset *Asset, err error)
		GetAssetBySymbol(symbol string) (asset *Asset, err error)
		GetAssetByAddress(address string) (asset *Asset, err error)
		GetGasAssets() (assets []*Asset, err error)
		GetMaxAssetId() (max int64, err error)
		CreateAssetsInTransact(tx *gorm.DB, assets []*Asset) error
		UpdateAssetsInTransact(tx *gorm.DB, assets []*Asset) error
	}

	defaultAssetModel struct {
		table string
		DB    *gorm.DB
	}

	Asset struct {
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

func (*Asset) TableName() string {
	return AssetTableName
}

func NewAssetModel(db *gorm.DB) AssetModel {
	return &defaultAssetModel{
		table: AssetTableName,
		DB:    db,
	}
}

func (m *defaultAssetModel) CreateAssetTable() error {
	return m.DB.AutoMigrate(Asset{})
}

func (m *defaultAssetModel) DropAssetTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultAssetModel) GetAssetsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultAssetModel) GetAssets(limit int64, offset int64) (res []*Asset, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("id asc").Find(&res)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return res, nil
}

func (m *defaultAssetModel) CreateAssets(l2Assets []*Asset) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(l2Assets, len(l2Assets))
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

func (m *defaultAssetModel) GetAssetById(assetId int64) (res *Asset, err error) {
	dbTx := m.DB.Table(m.table).Where("asset_id = ?", assetId).Find(&res)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return res, nil
}

func (m *defaultAssetModel) GetAssetBySymbol(symbol string) (res *Asset, err error) {
	dbTx := m.DB.Table(m.table).Where("asset_symbol = ?", symbol).Find(&res)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return res, nil
}

func (m *defaultAssetModel) GetAssetByAddress(address string) (asset *Asset, err error) {
	dbTx := m.DB.Table(m.table).Where("asset_address = ?", address).Find(&asset)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return asset, nil
}

func (m *defaultAssetModel) GetGasAssets() (assets []*Asset, err error) {
	dbTx := m.DB.Table(m.table).Where("is_gas_asset = ?", IsGasAsset).Find(&assets)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return assets, nil
}

func (m *defaultAssetModel) GetMaxAssetId() (max int64, err error) {
	dbTx := m.DB.Table(m.table).Select("id").Order("id desc").Limit(1).Find(&max)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return max, nil
}

func (m *defaultAssetModel) CreateAssetsInTransact(tx *gorm.DB, assets []*Asset) error {
	dbTx := tx.Table(m.table).CreateInBatches(assets, len(assets))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateAsset
	}
	return nil
}

func (m *defaultAssetModel) UpdateAssetsInTransact(tx *gorm.DB, assets []*Asset) error {
	for _, asset := range assets {
		dbTx := tx.Table(m.table).Where("id = ?", asset.ID).Delete(&asset)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToUpdateAsset
		}
	}
	return nil
}
