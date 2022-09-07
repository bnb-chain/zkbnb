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

package nft

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	L2NftHistoryTableName = `l2_nft_history`
)

type (
	L2NftHistoryModel interface {
		CreateL2NftHistoryTable() error
		DropL2NftHistoryTable() error
		GetLatestNftsCountByBlockHeight(height int64) (count int64, err error)
		GetLatestNftsByBlockHeight(height int64, limit int, offset int) (
			rowsAffected int64, nftAssets []*L2NftHistory, err error,
		)
		CreateNftHistoriesInTransact(tx *gorm.DB, histories []*L2NftHistory) error
	}
	defaultL2NftHistoryModel struct {
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

func NewL2NftHistoryModel(db *gorm.DB) L2NftHistoryModel {
	return &defaultL2NftHistoryModel{
		table: L2NftHistoryTableName,
		DB:    db,
	}
}

func (*L2NftHistory) TableName() string {
	return L2NftHistoryTableName
}

func (m *defaultL2NftHistoryModel) CreateL2NftHistoryTable() error {
	return m.DB.AutoMigrate(L2NftHistory{})
}

func (m *defaultL2NftHistoryModel) DropL2NftHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultL2NftHistoryModel) GetLatestNftsCountByBlockHeight(height int64) (
	count int64, err error,
) {
	subQuery := m.DB.Table(m.table).Select("*").
		Where("nft_index = a.nft_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height", height)

	dbTx := m.DB.Table(m.table+" as a").
		Where("NOT EXISTS (?) AND l2_block_height <= ?", subQuery, height)

	if dbTx.Count(&count).Error != nil {
		return 0, types.DbErrSqlOperation
	}

	return count, nil
}

func (m *defaultL2NftHistoryModel) GetLatestNftsByBlockHeight(height int64, limit int, offset int) (
	rowsAffected int64, accountNftAssets []*L2NftHistory, err error,
) {
	subQuery := m.DB.Table(m.table).Select("*").
		Where("nft_index = a.nft_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height", height)

	dbTx := m.DB.Table(m.table+" as a").Select("*").
		Where("NOT EXISTS (?) AND l2_block_height <= ?", subQuery, height).
		Limit(limit).Offset(offset).
		Order("nft_index")

	if dbTx.Find(&accountNftAssets).Error != nil {
		return 0, nil, types.DbErrSqlOperation
	}
	return dbTx.RowsAffected, accountNftAssets, nil
}

func (m *defaultL2NftHistoryModel) CreateNftHistoriesInTransact(tx *gorm.DB, histories []*L2NftHistory) error {
	dbTx := tx.Table(m.table).CreateInBatches(histories, len(histories))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(histories)) {
		return types.DbErrFailToCreateNftHistory
	}
	return nil
}
