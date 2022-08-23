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

package tx

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/types"
)

type (
	FailTxModel interface {
		CreateFailTxTable() error
		DropFailTxTable() error
		CreateFailTx(failTx *FailTx) error
	}

	defaultFailTxModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	FailTx struct {
		gorm.Model
		TxHash        string `gorm:"uniqueIndex"`
		TxType        int64
		GasFee        string
		GasFeeAssetId int64
		TxStatus      int64 // tx status, 1 - success(default), 2 - failure
		AssetAId      int64
		AssetBId      int64
		TxAmount      string
		NativeAddress string
		TxInfo        string
		ExtraInfo     string
		Memo          string
	}
)

func NewFailTxModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) FailTxModel {
	return &defaultFailTxModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      `fail_tx`,
		DB:         db,
	}
}

func (*FailTx) TableName() string {
	return `fail_tx`
}

func (m *defaultFailTxModel) CreateFailTxTable() error {
	return m.DB.AutoMigrate(FailTx{})
}

func (m *defaultFailTxModel) DropFailTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultFailTxModel) CreateFailTx(failTx *FailTx) error {
	dbTx := m.DB.Table(m.table).Create(failTx)
	if dbTx.Error != nil {
		logx.Errorf("create fail tx error, err: %s", dbTx.Error.Error())
		return types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateFailTx
	}
	return nil
}
