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

package asset

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	AssetModel interface {
		CreateAssetTable() error
		DropAssetTable() error
		CreateAssetsInBatch(l2Assets []*Asset) (rowsAffected int64, err error)
		GetAssetsList() (res []*Asset, err error)
		GetAssetByAssetId(assetId int64) (res *Asset, err error)
		GetAssetBySymbol(symbol string) (res *Asset, err error)
		GetAssetByAddress(address string) (asset *Asset, err error)
	}

	defaultAssetModel struct {
		sqlc.CachedConn
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

func NewAssetModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AssetModel {
	return &defaultAssetModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      AssetTableName,
		DB:         db,
	}
}

func (m *defaultAssetModel) CreateAssetTable() error {
	return m.DB.AutoMigrate(Asset{})
}

func (m *defaultAssetModel) DropAssetTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultAssetModel) GetAssetsList() (res []*Asset, err error) {
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

func (m *defaultAssetModel) CreateAssetsInBatch(l2Assets []*Asset) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(l2Assets, len(l2Assets))
	if dbTx.Error != nil {
		logx.Errorf("create assets error, err: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

func (m *defaultAssetModel) GetAssetByAssetId(assetId int64) (res *Asset, err error) {
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

func (m *defaultAssetModel) GetAssetBySymbol(symbol string) (res *Asset, err error) {
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

func (m *defaultAssetModel) GetAssetByAddress(address string) (asset *Asset, err error) {
	dbTx := m.DB.Table(m.table).Where("asset_address = ?", address).Find(&asset)
	if dbTx.Error != nil {
		logx.Errorf("fail to get asset by address: %s, error: %s", address, dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return asset, nil
}
