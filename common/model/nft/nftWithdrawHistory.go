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

package nft

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	L2NftWithdrawHistoryModel interface {
		CreateL2NftWithdrawHistoryTable() error
		DropL2NftWithdrawHistoryTable() error
		GetNftAsset(nftIndex int64) (nftAsset *L2NftWithdrawHistory, err error)
	}
	defaultL2NftWithdrawHistoryModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2NftWithdrawHistory struct {
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

func NewL2NftWithdrawHistoryModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2NftWithdrawHistoryModel {
	return &defaultL2NftWithdrawHistoryModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      L2NftWithdrawHistoryTableName,
		DB:         db,
	}
}

func (*L2NftWithdrawHistory) TableName() string {
	return L2NftWithdrawHistoryTableName
}

/*
	Func: CreateL2NftWithdrawHistoryTable
	Params:
	Return: err error
	Description: create account l2 nft table
*/
func (m *defaultL2NftWithdrawHistoryModel) CreateL2NftWithdrawHistoryTable() error {
	return m.DB.AutoMigrate(L2NftWithdrawHistory{})
}

/*
	Func: DropL2NftWithdrawHistoryTable
	Params:
	Return: err error
	Description: drop account l2 nft table
*/
func (m *defaultL2NftWithdrawHistoryModel) DropL2NftWithdrawHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultL2NftWithdrawHistoryModel) GetNftAsset(nftIndex int64) (nftAsset *L2NftWithdrawHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("nft_index = ?", nftIndex).Find(&nftAsset)
	if dbTx.Error != nil {
		logx.Errorf("[GetNftAsset] unable to get nft asset: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[GetNftAsset] no such info")
		return nil, errorcode.DbErrNotFound
	}
	return nftAsset, nil
}
