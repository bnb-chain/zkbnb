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
	L2NftWithdrawHistoryModel interface {
		CreateL2NftWithdrawHistoryTable() error
		DropL2NftWithdrawHistoryTable() error
	}
	defaultL2NftWithdrawHistoryModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2NftWithdrawHistory struct {
		gorm.Model
		NftIndex            int64
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

func (m *defaultL2NftWithdrawHistoryModel) CreateL2NftWithdrawHistoryTable() error {
	return m.DB.AutoMigrate(L2NftWithdrawHistory{})
}

func (m *defaultL2NftWithdrawHistoryModel) DropL2NftWithdrawHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
