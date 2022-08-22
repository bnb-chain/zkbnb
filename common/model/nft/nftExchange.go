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
)

type (
	L2NftExchangeModel interface {
		CreateL2NftExchangeTable() error
		DropL2NftExchangeTable() error
	}
	defaultL2NftExchangeModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2NftExchange struct {
		gorm.Model
		BuyerAccountIndex int64
		OwnerAccountIndex int64
		NftIndex          int64
		AssetId           int64
		AssetAmount       string
	}
)

func NewL2NftExchangeModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2NftExchangeModel {
	return &defaultL2NftExchangeModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      L2NftExchangeTableName,
		DB:         db,
	}
}

func (*L2NftExchange) TableName() string {
	return L2NftExchangeTableName
}

/*
Func: CreateL2NftExchangeTable
Params:
Return: err error
Description: create account l2 nft table
*/
func (m *defaultL2NftExchangeModel) CreateL2NftExchangeTable() error {
	return m.DB.AutoMigrate(L2NftExchange{})
}

/*
Func: DropL2NftExchangeTable
Params:
Return: err error
Description: drop account nft exchange table
*/
func (m *defaultL2NftExchangeModel) DropL2NftExchangeTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
