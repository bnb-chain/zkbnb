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
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	L2NftTableName = `l2_nft`
)

type (
	L2NftModel interface {
		CreateL2NftTable() error
		DropL2NftTable() error
		GetNft(nftIndex int64) (nftAsset *L2Nft, err error)
		GetLatestNftIndex() (nftIndex int64, err error)
		GetNftsByAccountIndex(accountIndex, limit, offset int64) (nfts []*L2Nft, err error)
		GetNftsCountByAccountIndex(accountIndex int64) (int64, error)
		UpdateNftsInTransact(tx *gorm.DB, nfts []*L2Nft) error
		BatchInsertOrUpdateInTransact(tx *gorm.DB, nfts []*L2Nft) (err error)
		DeleteByIndexesInTransact(tx *gorm.DB, nftIndexes []int64) error
		UpdateByIndexInTransact(tx *gorm.DB, l2nft *L2Nft) error
		GetNfts(limit int64, offset int64) (nfts []*L2Nft, err error)
		GetCountByGreaterHeight(blockHeight int64) (count int64, err error)
		UpdateIpfsStatusByNftIndexInTransact(tx *gorm.DB, nftIndex int64) error
		GetMaxNftIndex() (nftIndex int64, err error)
	}
	defaultL2NftModel struct {
		table string
		DB    *gorm.DB
	}

	L2Nft struct {
		gorm.Model
		NftIndex            int64 `gorm:"uniqueIndex"`
		CreatorAccountIndex int64
		OwnerAccountIndex   int64  `gorm:"index:idx_owner_account_index"`
		NftContentHash      string `gorm:"index:idx_owner_account_index"`
		CreatorTreasuryRate int64
		CollectionId        int64
		L2BlockHeight       int64 `gorm:"index:idx_nft_index"`
	}
)

func NewL2NftModel(db *gorm.DB) L2NftModel {
	return &defaultL2NftModel{
		table: L2NftTableName,
		DB:    db,
	}
}

func (*L2Nft) TableName() string {
	return L2NftTableName
}

func (m *defaultL2NftModel) CreateL2NftTable() error {
	return m.DB.AutoMigrate(L2Nft{})
}

func (m *defaultL2NftModel) DropL2NftTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultL2NftModel) GetNft(nftIndex int64) (nftAsset *L2Nft, err error) {
	dbTx := m.DB.Table(m.table).Where("nft_index = ?", nftIndex).Find(&nftAsset)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return nftAsset, nil
}

func (m *defaultL2NftModel) GetLatestNftIndex() (nftIndex int64, err error) {
	var nftInfo *L2Nft
	dbTx := m.DB.Table(m.table).Order("nft_index desc").Find(&nftInfo)
	if dbTx.Error != nil {
		return -1, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return -1, nil
	}
	return nftInfo.NftIndex, nil
}

func (m *defaultL2NftModel) GetNftsByAccountIndex(accountIndex, limit, offset int64) (nftList []*L2Nft, err error) {
	dbTx := m.DB.Table(m.table).Where("owner_account_index = ? and nft_content_hash != ?", accountIndex, "0").
		Limit(int(limit)).Offset(int(offset)).Order("nft_index desc").Find(&nftList)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return nftList, nil
}

func (m *defaultL2NftModel) GetNftsCountByAccountIndex(accountIndex int64) (int64, error) {
	var count int64
	dbTx := m.DB.Table(m.table).Where("owner_account_index = ? and nft_content_hash != ?", accountIndex, "0").Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return count, nil
}

func (m *defaultL2NftModel) UpdateNftsInTransact(tx *gorm.DB, nfts []*L2Nft) error {
	for _, pendingNft := range nfts {
		dbTx := tx.Table(m.table).Where("nft_index = ?", pendingNft.NftIndex).
			Select("*").
			Updates(&pendingNft)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			dbTx = tx.Table(m.table).Create(&pendingNft)
			if dbTx.Error != nil {
				return dbTx.Error
			}
		}
	}
	return nil
}

func (m *defaultL2NftModel) BatchInsertOrUpdateInTransact(tx *gorm.DB, nfts []*L2Nft) (err error) {
	dbTx := tx.Table(m.table).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"creator_account_index", "owner_account_index", "nft_content_hash", "creator_treasury_rate", "collection_id", "l2_block_height"}),
	}).CreateInBatches(&nfts, len(nfts))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if int(dbTx.RowsAffected) != len(nfts) {
		logx.Errorf("BatchInsertOrUpdateInTransact failed,rows affected not equal nfts length,dbTx.RowsAffected:%s,len(nfts):%s", int(dbTx.RowsAffected), len(nfts))
		return types.DbErrFailToUpdateAccount
	}
	return nil
}

func (m *defaultL2NftModel) DeleteByIndexesInTransact(tx *gorm.DB, nftIndexes []int64) error {
	if len(nftIndexes) == 0 {
		return nil
	}
	dbTx := tx.Model(&L2Nft{}).Unscoped().Where("nft_index in ?", nftIndexes).Delete(&L2Nft{})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToUpdateNft
	}
	return nil
}

func (m *defaultL2NftModel) UpdateByIndexInTransact(tx *gorm.DB, l2nft *L2Nft) error {
	dbTx := tx.Model(&L2Nft{}).Select("creator_account_index", "owner_account_index", "nft_content_hash", "creator_treasury_rate", "collection_id", "l2_block_height").Where("nft_index = ?", l2nft.NftIndex).Updates(map[string]interface{}{
		"creator_account_index": l2nft.CreatorAccountIndex,
		"owner_account_index":   l2nft.OwnerAccountIndex,
		"nft_content_hash":      l2nft.NftContentHash,
		"creator_treasury_rate": l2nft.CreatorTreasuryRate,
		"collection_id":         l2nft.CollectionId,
		"l2_block_height":       l2nft.L2BlockHeight,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToUpdateNft
	}
	return nil
}

func (m *defaultL2NftModel) GetNfts(limit int64, offset int64) (nfts []*L2Nft, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("id asc").Find(&nfts)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, nil
	}
	return nfts, nil
}

func (m *defaultL2NftModel) GetCountByGreaterHeight(blockHeight int64) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height > ?", blockHeight).Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultL2NftModel) UpdateIpfsStatusByNftIndexInTransact(tx *gorm.DB, nftIndex int64) error {
	dbTx := tx.Model(&L2Nft{}).Unscoped().Where("nft_index = ?", nftIndex).Update("ipfs_status", Confirmed)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}
func (m *defaultL2NftModel) GetMaxNftIndex() (nftIndex int64, err error) {
	dbTx := m.DB.Table(m.table).Select("max(nft_index)").Find(&nftIndex)
	if dbTx.Error != nil {
		return -1, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return -1, types.DbErrNotFound
	}
	return nftIndex, nil
}

func (ai *L2Nft) DeepCopy() *L2Nft {
	l2Nft := &L2Nft{
		Model:               gorm.Model{ID: ai.ID},
		NftIndex:            ai.NftIndex,
		CreatorAccountIndex: ai.CreatorAccountIndex,
		OwnerAccountIndex:   ai.OwnerAccountIndex,
		NftContentHash:      ai.NftContentHash,
		CreatorTreasuryRate: ai.CreatorTreasuryRate,
		CollectionId:        ai.CollectionId,
		L2BlockHeight:       ai.L2BlockHeight,
	}
	return l2Nft
}
