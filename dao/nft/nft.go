/*
 * Copyright © 2021 ZkBAS Protocol
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
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/types"
)

const (
	L2NftTableName = `l2_nft`
)

type (
	L2NftModel interface {
		CreateL2NftTable() error
		DropL2NftTable() error
		GetNftAsset(nftIndex int64) (nftAsset *L2Nft, err error)
		GetLatestNftIndex() (nftIndex int64, err error)
		GetNftListByAccountIndex(accountIndex, limit, offset int64) (nfts []*L2Nft, err error)
		GetAccountNftTotalCount(accountIndex int64) (int64, error)
	}
	defaultL2NftModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2Nft struct {
		gorm.Model
		NftIndex            int64 `gorm:"uniqueIndex"`
		CreatorAccountIndex int64
		OwnerAccountIndex   int64
		NftContentHash      string
		NftL1Address        string
		NftL1TokenId        string
		CreatorTreasuryRate int64
		CollectionId        int64
	}
)

func NewL2NftModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2NftModel {
	return &defaultL2NftModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      L2NftTableName,
		DB:         db,
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

func (m *defaultL2NftModel) GetNftAsset(nftIndex int64) (nftAsset *L2Nft, err error) {
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

func (m *defaultL2NftModel) GetNftListByAccountIndex(accountIndex, limit, offset int64) (nftList []*L2Nft, err error) {
	dbTx := m.DB.Table(m.table).Where("owner_account_index = ? and deleted_at is NULL", accountIndex).
		Limit(int(limit)).Offset(int(offset)).Order("nft_index desc").Find(&nftList)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return nftList, nil
}

func (m *defaultL2NftModel) GetAccountNftTotalCount(accountIndex int64) (int64, error) {
	var count int64
	dbTx := m.DB.Table(m.table).Where("owner_account_index = ? and deleted_at is NULL", accountIndex).Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return count, nil
}
