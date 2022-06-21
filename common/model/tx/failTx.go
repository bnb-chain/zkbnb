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

package tx

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
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
		TxStatus      int64
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

/*
	Func: CreateFailTxTable
	Params:
	Return: err error
	Description: create txVerification fail table
*/
func (m *defaultFailTxModel) CreateFailTxTable() error {
	return m.DB.AutoMigrate(FailTx{})
}

/*
	Func: DropFailTxTable
	Params:
	Return: err error
	Description: drop txVerification fail table
*/
func (m *defaultFailTxModel) DropFailTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateFailTx
	Params: failTx *FailTx
	Return: err error
	Description: create fail txVerification
*/
func (m *defaultFailTxModel) CreateFailTx(failTx *FailTx) error {
	dbTx := m.DB.Table(m.table).Create(failTx)
	if dbTx.Error != nil {
		logx.Error("[txVerification.CreateFailTx] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Error("[txVerification.CreateFailTx] Create Invalid Fail Tx")
		return ErrInvalidFailTx
	}
	return nil
}
