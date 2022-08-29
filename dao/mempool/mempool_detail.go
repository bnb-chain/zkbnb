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

package mempool

import (
	"time"

	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/types"
)

const (
	DetailTableName = `mempool_tx_detail`
)

type (
	MempoolTxDetailModel interface {
		CreateMempoolDetailTable() error
		DropMempoolDetailTable() error
		GetMempoolTxDetailsByAccountIndex(accountIndex int64) (mempoolTxDetails []*MempoolTxDetail, err error)
	}

	defaultMempoolDetailModel struct {
		table string
		DB    *gorm.DB
	}

	MempoolTxDetail struct {
		gorm.Model
		TxId         int64 `json:"tx_id" gorm:"index;not null"`
		AssetId      int64
		AssetType    int64
		AccountIndex int64 `gorm:"index"`
		AccountName  string
		BalanceDelta string
		Order        int64
		AccountOrder int64
	}

	LatestTimeMempoolDetails struct {
		Max     time.Time
		AssetId int64
	}
)

func NewMempoolDetailModel(db *gorm.DB) MempoolTxDetailModel {
	return &defaultMempoolDetailModel{
		table: DetailTableName,
		DB:    db,
	}
}

func (*MempoolTxDetail) TableName() string {
	return DetailTableName
}

func (m *defaultMempoolDetailModel) CreateMempoolDetailTable() error {
	return m.DB.AutoMigrate(MempoolTxDetail{})
}

func (m *defaultMempoolDetailModel) DropMempoolDetailTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultMempoolDetailModel) GetMempoolTxDetailsByAccountIndex(accountIndex int64) (mempoolTxDetails []*MempoolTxDetail, err error) {
	var dbTx *gorm.DB
	dbTx = m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&mempoolTxDetails)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return mempoolTxDetails, nil
}
