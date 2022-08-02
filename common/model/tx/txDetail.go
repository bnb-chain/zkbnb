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

package tx

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

var (
	cacheZkbasTxDetailIdPrefix = "cache:zkbas:txDetail:id:"
)

type (
	TxDetailModel interface {
		CreateTxDetailTable() error
		DropTxDetailTable() error
		GetTxDetailsByAccountName(name string) (txDetails []*TxDetail, err error)
		UpdateTxDetail(detail *TxDetail) error
	}

	defaultTxDetailModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	TxDetail struct {
		gorm.Model
		TxId            int64 `gorm:"index"`
		AssetId         int64
		AssetType       int64
		AccountIndex    int64 `gorm:"index"`
		AccountName     string
		Balance         string
		BalanceDelta    string
		Order           int64
		AccountOrder    int64
		Nonce           int64
		CollectionNonce int64
	}
)

func NewTxDetailModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) TxDetailModel {
	return &defaultTxDetailModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TxDetailTableName,
		DB:         db,
	}
}

func (*TxDetail) TableName() string {
	return TxDetailTableName
}

/*
	Func: CreateTxDetailTable
	Params:
	Return: err error
	Description: create txVerification detail table
*/
func (m *defaultTxDetailModel) CreateTxDetailTable() error {
	return m.DB.AutoMigrate(TxDetail{})
}

/*
	Func: DropTxDetailTable
	Params:
	Return: err error
	Description: drop txVerification detail table
*/
func (m *defaultTxDetailModel) DropTxDetailTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetTxDetailsByAccountName
	Params: name string
	Return: txDetails []*TxDetail, err error
	Description: GetTxDetailsByAccountName
*/
func (m *defaultTxDetailModel) GetTxDetailsByAccountName(name string) (txDetails []*TxDetail, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name = ?", name).Find(&txDetails)
	if dbTx.Error != nil {
		logx.Errorf("fail to get tx by account: %s, error: %s", name, dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return txDetails, nil
}

func (m *defaultTxDetailModel) UpdateTxDetail(detail *TxDetail) error {
	dbTx := m.DB.Save(&detail)
	if dbTx.Error != nil {
		if dbTx.Error == errorcode.DbErrNotFound {
			return nil
		} else {
			return dbTx.Error
		}
	} else {
		return nil
	}
}
