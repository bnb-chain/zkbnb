/*
 * Copyright © 2021 ZkBNB Protocol
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
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const TxDetailTableName = `tx_detail`

type (
	TxDetailModel interface {
		CreateTxDetailTable() error
		DropTxDetailTable() error
		CreateTxDetails(txDetails []*TxDetail) error
		DeleteByHeightsInTransact(tx *gorm.DB, heights []int64) error
	}

	defaultTxDetailModel struct {
		table string
		DB    *gorm.DB
	}

	TxDetail struct {
		gorm.Model
		PoolTxId        uint `gorm:"index"`
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
		IsGas           bool  `gorm:"default:false"`
		BlockHeight     int64 `gorm:"index"`
	}
)

func NewTxDetailModel(db *gorm.DB) TxDetailModel {
	return &defaultTxDetailModel{
		table: TxDetailTableName,
		DB:    db,
	}
}

func (*TxDetail) TableName() string {
	return TxDetailTableName
}

func (m *defaultTxDetailModel) CreateTxDetailTable() error {
	return m.DB.AutoMigrate(TxDetail{})
}

func (m *defaultTxDetailModel) DropTxDetailTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultTxDetailModel) CreateTxDetails(txDetails []*TxDetail) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(txDetails, len(txDetails))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(txDetails)) {
		logx.Errorf("CreateTxDetails failed,rows affected not equal txDetails length,dbTx.RowsAffected:%s,len(txDetails):%s", int(dbTx.RowsAffected), len(txDetails))
		return types.DbErrFailToCreateTxDetail
	}
	return nil
}

func (m *defaultTxDetailModel) DeleteByHeightsInTransact(tx *gorm.DB, heights []int64) error {
	if len(heights) == 0 {
		return nil
	}
	dbTx := tx.Model(&TxDetail{}).Unscoped().Where("block_height in ?", heights).Delete(&TxDetail{})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}
