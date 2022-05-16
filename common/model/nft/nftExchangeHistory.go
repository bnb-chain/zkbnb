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
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	L2NftExchangeHistoryModel interface {
		CreateL2NftExchangeHistoryTable() error
		DropL2NftExchangeHistoryTable() error
	}
	defaultL2NftExchangeHistoryModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2NftExchangeHistory struct {
		gorm.Model
		BuyerAccountIndex int64
		OwnerAccountIndex int64
		NftIndex          int64
		AssetId           int64
		AssetAmount       int64
		L2BlockHeight     int64
	}
)

func NewL2NftExchangeHistoryModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2NftExchangeHistoryModel {
	return &defaultL2NftExchangeHistoryModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      L2NftExchangeHistoryTableName,
		DB:         db,
	}
}

func (*L2NftExchangeHistory) TableName() string {
	return L2NftExchangeHistoryTableName
}

/*
	Func: CreateL2NftExchangeHistoryTable
	Params:
	Return: err error
	Description: create account l2 nft exchange history table
*/
func (m *defaultL2NftExchangeHistoryModel) CreateL2NftExchangeHistoryTable() error {
	return m.DB.AutoMigrate(L2NftExchangeHistory{})
}

/*
	Func: DropL2NftExchangeHistoryTable
	Params:
	Return: err error
	Description: drop accountnft table
*/
func (m *defaultL2NftExchangeHistoryModel) DropL2NftExchangeHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
