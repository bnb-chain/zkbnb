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
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	L2NftModel interface {
		CreateL2NftTable() error
		DropL2NftTable() error
		GetNftAsset(nftIndex int64) (nftAsset *L2Nft, err error)
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
		AssetId             int64
		AssetAmount         string
		CreatorTreasuryRate int64
		CollectionId        int64
		Status              int
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

/*
	Func: CreateL2NftTable
	Params:
	Return: err error
	Description: create account l2 nft table
*/
func (m *defaultL2NftModel) CreateL2NftTable() error {
	return m.DB.AutoMigrate(L2Nft{})
}

/*
	Func: DropL2NftTable
	Params:
	Return: err error
	Description: drop account l2 nft table
*/
func (m *defaultL2NftModel) DropL2NftTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultL2NftModel) GetNftAsset(nftIndex int64) (nftAsset *L2Nft, err error) {
	dbTx := m.DB.Table(m.table).Where("nft_index = ?", nftIndex).Find(&nftAsset)
	if dbTx.Error != nil {
		logx.Errorf("[GetNftAsset] unable to get nft asset: %s", err.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[GetNftAsset] no such info")
		return nil, ErrNotFound
	}
	return nftAsset, nil
}
