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

package nft

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	L2NftHistoryModel interface {
		CreateL2NftHistoryTable() error
		DropL2NftHistoryTable() error
		GetLatestNftAssetsByBlockHeight(height int64) (
			rowsAffected int64, nftAssets []*L2NftHistory, err error,
		)
		GetLatestNftAsset(nftIndex int64) (
			nftAsset *L2NftHistory, err error,
		)
		GetNftAssetsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, nftAssets []*L2NftHistory, err error)
	}
	defaultL2NftHistoryModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2NftHistory struct {
		gorm.Model
		NftIndex            int64
		CreatorAccountIndex int64
		OwnerAccountIndex   int64
		NftContentHash      string
		NftL1Address        string
		NftL1TokenId        string
		CreatorTreasuryRate int64
		CollectionId        int64
		Status              int
		L2BlockHeight       int64
	}
)

func NewL2NftHistoryModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2NftHistoryModel {
	return &defaultL2NftHistoryModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      L2NftHistoryTableName,
		DB:         db,
	}
}

func (*L2NftHistory) TableName() string {
	return L2NftHistoryTableName
}

/*
	Func: CreateL2NftHistoryTable
	Params:
	Return: err error
	Description: create account l2 nft table
*/
func (m *defaultL2NftHistoryModel) CreateL2NftHistoryTable() error {
	return m.DB.AutoMigrate(L2NftHistory{})
}

/*
	Func: DropL2NftHistoryTable
	Params:
	Return: err error
	Description: drop account l2 nft history table
*/
func (m *defaultL2NftHistoryModel) DropL2NftHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultL2NftHistoryModel) GetLatestNftAssetsByBlockHeight(height int64) (
	rowsAffected int64, accountNftAssets []*L2NftHistory, err error,
) {
	// TODO sql
	dbTx := m.DB.Table(m.table).
		Raw("SELECT a.* FROM l2_nft_history a WHERE NOT EXISTS"+
			"(SELECT * FROM l2_nft_history WHERE nft_index = a.nft_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height) "+
			"AND l2_block_height <= ? ORDER BY nft_index", height, height).
		Find(&accountNftAssets)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestNftAssetsByBlockHeight] unable to get related nft assets: %s", dbTx.Error.Error())
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, accountNftAssets, nil
}

func (m *defaultL2NftHistoryModel) GetLatestNftAsset(nftIndex int64) (
	nftAsset *L2NftHistory, err error,
) {
	dbTx := m.DB.Table(m.table).Where("nft_index = ?", nftIndex).Order("l2_block_height desc").First(&nftAsset)

	if dbTx.Error != nil {
		logx.Errorf("[GetLatestNftAsset] unable to get related nft asset: %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[GetLatestNftAsset] no such info")
		return nil, ErrNotFound
	}
	return nftAsset, nil
}

func (m *defaultL2NftHistoryModel) GetNftAssetsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, nftAssets []*L2NftHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height = ?", l2BlockHeight).Find(&nftAssets)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[GetLiquidityAssetsByBlockHeight] unable to get related assets: %s", err.Error())
		logx.Error(errInfo)
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, nftAssets, nil
}
