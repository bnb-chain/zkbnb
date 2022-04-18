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
	L2NftCollectionModel interface {
		CreateL2NftCollectionTable() error
		DropL2NftCollectionTable() error
	}
	defaultL2NftCollectionModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2NftCollection struct {
		gorm.Model
		AccountIndex int64
		Name         string
		Introduction string
	}
)

func NewL2NftCollectionModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2NftCollectionModel {
	return &defaultL2NftCollectionModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      L2NftCollectionTableName,
		DB:         db,
	}
}

func (*L2NftCollection) TableName() string {
	return L2NftCollectionTableName
}

/*
	Func: CreateL2NftCollectionTable
	Params:
	Return: err error
	Description: create account l2 nft table
*/
func (m *defaultL2NftCollectionModel) CreateL2NftCollectionTable() error {
	return m.DB.AutoMigrate(L2NftCollection{})
}

/*
	Func: DropAccountNFTTable
	Params:
	Return: err error
	Description: drop accountnft table
*/
func (m *defaultL2NftCollectionModel) DropL2NftCollectionTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
