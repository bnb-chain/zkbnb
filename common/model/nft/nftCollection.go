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

	"github.com/bnb-chain/zkbas/common/errorcode"
)

type (
	L2NftCollectionModel interface {
		CreateL2NftCollectionTable() error
		DropL2NftCollectionTable() error
		IfCollectionExistsByCollectionId(accountIndex, collectionId int64) (bool, error)
	}
	defaultL2NftCollectionModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2NftCollection struct {
		gorm.Model
		AccountIndex int64
		CollectionId int64
		Name         string
		Introduction string
		Status       int //Collection status indicates whether it is certified by L2. 0 means no, 1 means yes，Change this state with a transaction in the future
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
	Func: DropL2NftCollectionTable
	Params:
	Return: err error
	Description: drop account nft collection table
*/
func (m *defaultL2NftCollectionModel) DropL2NftCollectionTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
func (m *defaultL2NftCollectionModel) IfCollectionExistsByCollectionId(accountIndex, collectionId int64) (bool, error) {
	var res int64
	dbTx := m.DB.Table(m.table).Where("account_index = ? and collection_id = ? and deleted_at is NULL", accountIndex, collectionId).Count(&res)

	if dbTx.Error != nil {
		logx.Errorf("get collection count error, err: %s", dbTx.Error.Error())
		return true, errorcode.DbErrSqlOperation
	} else if res == 0 {
		return false, nil
	} else if res != 1 {
		return true, errorcode.DbErrDuplicatedCollectionIndex
	} else {
		return true, nil
	}
}
