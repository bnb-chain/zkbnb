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

package mempool

import (
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheZecreyMempoolDetailIdPrefix = "cache:zecrey:mempoolDetail:id:"
)

type (
	MempoolTxDetailModel interface {
		CreateMempoolDetailTable() error
		DropMempoolDetailTable() error
		GetLatestMempoolDetail(accountIndex int64, assetId int64, assetType int64) (mempoolTxDetail *MempoolTxDetail, err error)
		GetAccountAssetsMempoolDetails(accountIndex int64, assetType int64) (mempoolTxDetails []*MempoolTxDetail, err error)
		GetAccountMempoolDetails(accountIndex int64) (mempoolTxDetails []*MempoolTxDetail, err error)
		GetMempoolTxDetailsByAccountIndex(accountIndex int64) (mempoolTxDetails []*MempoolTxDetail, err error)
		GetAccountAssetMempoolDetails(accountIndex int64, assetId int64, assetType int64) (mempoolTxDetails []*MempoolTxDetail, err error)
		GetLatestAccountAssetMempoolDetail(accountIndex int64, assetId int64, assetType int64) (mempoolTxDetail *MempoolTxDetail, err error)
		GetMempoolTxDetailsByAssetType(assetType int) (mempoolTxDetails []*MempoolTxDetail, err error)
		GetMempoolTxDetailsByAssetIdAndAssetType(assetId int64, assetType int) (mempoolTxDetails []*MempoolTxDetail, err error)
	}

	defaultMempoolDetailModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	MempoolTxDetail struct {
		gorm.Model
		TxId         int64 `json:"tx_id" gorm:"index;not null"`
		AssetId      int64
		AssetType    int64
		AccountIndex int64 `gorm:"index"`
		AccountName  string
		BalanceDelta string
		Order        int64
		AccountOrder int64
	}

	LatestTimeMempoolDetails struct {
		Max     time.Time
		AssetId int64
	}
)

func NewMempoolDetailModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) MempoolTxDetailModel {
	return &defaultMempoolDetailModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      DetailTableName,
		DB:         db,
	}
}

func (*MempoolTxDetail) TableName() string {
	return DetailTableName
}

/*
	Func: CreateMempoolDetailTable
	Params:
	Return: err error
	Description: create mempool detail table
*/

func (m *defaultMempoolDetailModel) CreateMempoolDetailTable() error {
	return m.DB.AutoMigrate(MempoolTxDetail{})
}

/*
	Func: DropMempoolDetailTable
	Params:
	Return: err error
	Description: drop MempoolDetail table
*/

func (m *defaultMempoolDetailModel) DropMempoolDetailTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetLatestMempoolDetail
	Params:	AccountIndex int64, AssetId int64, AssetType int64
	Return: err error
	Description: get latest(create_at desc) mempool detail info from mempool_detail table by accountIndex, assetId and assetType.
				It will be used to check if the value in Balance global map is valid.
*/
func (m *defaultMempoolDetailModel) GetLatestMempoolDetail(accountIndex int64, assetId int64, assetType int64) (mempoolTxDetail *MempoolTxDetail, err error) {
	dbTx := m.DB.Table(m.table).Where(
		"account_index = ? and asset_id = ? and asset_type = ?", accountIndex, assetId, assetType).
		Order("created_at desc, id desc").Limit(1).Find(&mempoolTxDetail)
	if dbTx.Error != nil {
		logx.Errorf("[mempoolDetail.GetLatestMempoolDetail] %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempoolDetail.GetLatestMempoolDetail] Get MempoolTxDetail Error")
		return nil, ErrNotFound
	}
	return mempoolTxDetail, nil
}

/*
	Func: GetAccountAssetsMempoolDetails
	Params:	accountIndex int64, assetType int64
	Return: mempoolTxDetails []*MempoolTxDetail, err error
	Description: used for get globalmap data source
*/
func (m *defaultMempoolDetailModel) GetAccountAssetsMempoolDetails(accountIndex int64, assetType int64) (mempoolTxDetails []*MempoolTxDetail, err error) {
	var dbTx *gorm.DB
	dbTx = m.DB.Table(m.table).Where("account_index = ? and asset_type = ? and chain_id != -1", accountIndex, assetType).
		Order("created_at, id").Find(&mempoolTxDetails)
	if dbTx.Error != nil {
		logx.Error("[mempoolDetail.GetAccountAssetsMempoolDetails] %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[mempoolDetail.GetAccountAssetsMempoolDetails] Get MempoolTxDetails Error")
		return nil, ErrNotFound
	}
	return mempoolTxDetails, nil
}

/*
	Func: GetAccountAssetMempoolDetails
	Params:	AccountIndex int64, AssetId int64, AssetType int64
	Return: err error
	Description: used for get globalmap data source
*/
func (m *defaultMempoolDetailModel) GetAccountAssetMempoolDetails(accountIndex int64, assetId int64, assetType int64) (mempoolTxDetails []*MempoolTxDetail, err error) {
	var dbTx *gorm.DB
	dbTx = m.DB.Table(m.table).Where("account_index = ? and asset_id = ? and asset_type = ? ", accountIndex, assetId, assetType).
		Order("created_at, id").Find(&mempoolTxDetails)
	if dbTx.Error != nil {
		logx.Errorf("[mempoolDetail.GetAccountAssetMempoolDetails] %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempoolDetail.GetAccountAssetMempoolDetails] Get MempoolTxDetails Error")
		return nil, ErrNotFound
	}
	return mempoolTxDetails, nil
}

func (m *defaultMempoolDetailModel) GetLatestAccountAssetMempoolDetail(
	accountIndex int64, assetId int64, assetType int64,
) (mempoolTxDetail *MempoolTxDetail, err error) {
	var dbTx *gorm.DB
	dbTx = m.DB.Table(m.table).Where("account_index = ? and asset_id = ? and asset_type = ? ",
		accountIndex, assetId, assetType).
		Order("created_at desc, id desc").Find(&mempoolTxDetail)
	if dbTx.Error != nil {
		logx.Errorf("[mempoolDetail.GetAccountAssetMempoolDetails] %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempoolDetail.GetAccountAssetMempoolDetails] no related mempool tx detail")
		return nil, ErrNotFound
	}
	return mempoolTxDetail, nil
}

func (m *defaultMempoolDetailModel) GetAccountMempoolDetails(accountIndex int64) (mempoolTxDetails []*MempoolTxDetail, err error) {
	var dbTx *gorm.DB
	dbTx = m.DB.Table(m.table).Where("account_index = ?", accountIndex).
		Order("created_at").Find(&mempoolTxDetails)
	if dbTx.Error != nil {
		logx.Errorf("[mempoolDetail.GetAccountMempoolDetails] %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempoolDetail.GetAccountMempoolDetails] no related mempool tx details")
		return nil, ErrNotFound
	}
	return mempoolTxDetails, nil
}

func (m *defaultMempoolDetailModel) GetMempoolTxDetailsByAssetType(assetType int) (mempoolTxDetails []*MempoolTxDetail, err error) {
	var dbTx *gorm.DB
	dbTx = m.DB.Table(m.table).Where("asset_type = ?", assetType).
		Order("created_at").Find(&mempoolTxDetails)
	if dbTx.Error != nil {
		logx.Errorf("[mempoolDetail.GetAccountMempoolDetails] %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempoolDetail.GetAccountMempoolDetails] no related mempool tx details")
		return nil, ErrNotFound
	}
	return mempoolTxDetails, nil
}

func (m *defaultMempoolDetailModel) GetMempoolTxDetailsByAssetIdAndAssetType(
	assetId int64,
	assetType int,
) (
	mempoolTxDetails []*MempoolTxDetail, err error) {
	var dbTx *gorm.DB
	dbTx = m.DB.Table(m.table).Where("asset_id = ? AND asset_type = ?", assetId, assetType).
		Order("created_at").Find(&mempoolTxDetails)
	if dbTx.Error != nil {
		logx.Errorf("[mempoolDetail.GetAccountMempoolDetails] %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempoolDetail.GetAccountMempoolDetails] no related mempool tx details")
		return nil, ErrNotFound
	}
	return mempoolTxDetails, nil
}

func (m *defaultMempoolDetailModel) GetMempoolTxDetailsByAccountIndex(accountIndex int64) (mempoolTxDetails []*MempoolTxDetail, err error) {
	var dbTx *gorm.DB
	dbTx = m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&mempoolTxDetails)
	if dbTx.Error != nil {
		logx.Errorf("[GetMempoolTxDetailsByAccountIndex] unable to get by account index: %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return mempoolTxDetails, nil
}
