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

package nft

import (
	"github.com/bnb-chain/zkbnb/types"
	"gorm.io/gorm"
	"time"
)

const (
	L2NftMetadataHistoryTableName = `l2_nft_metadata_history`
)

const (
	StatusFailed = iota
	StatusPending
	NotConfirmed
	Confirmed
)

type (
	L2NftMetadataHistoryModel interface {
		CreateL2NftMetadataHistoryTable() error
		DropL2NftMetadataHistoryTable() error
		CreateL2NftMetadataHistoryInTransact(tx *gorm.DB, metadata *L2NftMetadataHistory) error
		DeleteInTransact(id uint) error
		GetL2NftMetadataHistoryList(status int) (history []*L2NftMetadataHistory, err error)
		GetL2NftMetadataHistoryPage(status, limit, offset int) (history []*L2NftMetadataHistory, err error)
		GetL2NftMetadataHistory(nftIndex int64) (history *L2NftMetadataHistory, err error)
		GetL2NftMetadataHistoryByHash(txHash string) (history *L2NftMetadataHistory, err error)
		UpdateL2NftMetadataHistoryInTransact(history *L2NftMetadataHistory) error
		UpdateL2NftMetadataHistoryNoNftIndex(history *L2NftMetadataHistory) error
	}

	defaultL2NftMetadataHistoryModel struct {
		table string
		DB    *gorm.DB
	}

	L2NftMetadataHistory struct {
		gorm.Model
		Nonce    int64
		NftIndex int64  `gorm:"index"`
		TxHash   string `gorm:"index"`
		IpfsCid  string
		IpnsCid  string
		IpnsName string
		IpnsId   string
		Metadata string
		Mutable  string
		Status   int64 `gorm:"index"`
	}
)

func NewL2NftMetadataHistoryModel(db *gorm.DB) L2NftMetadataHistoryModel {
	return &defaultL2NftMetadataHistoryModel{
		table: L2NftMetadataHistoryTableName,
		DB:    db,
	}
}

func (*L2NftMetadataHistory) TableName() string {
	return L2NftMetadataHistoryTableName
}

func (m *defaultL2NftMetadataHistoryModel) CreateL2NftMetadataHistoryTable() error {
	return m.DB.AutoMigrate(L2NftMetadataHistory{})
}

func (m *defaultL2NftMetadataHistoryModel) DropL2NftMetadataHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
func (m *defaultL2NftMetadataHistoryModel) CreateL2NftMetadataHistoryInTransact(tx *gorm.DB, metadata *L2NftMetadataHistory) (err error) {
	dbTx := tx.Table(m.table).Create(metadata)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateProof
	}
	return nil
}

func (m *defaultL2NftMetadataHistoryModel) GetL2NftMetadataHistoryList(status int) (history []*L2NftMetadataHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", status).
		Limit(500).Order("id asc").Find(&history)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return history, nil
}

func (m *defaultL2NftMetadataHistoryModel) GetL2NftMetadataHistoryPage(status, limit, offset int) (history []*L2NftMetadataHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", status).
		Limit(limit).Offset(offset).Order("id asc").Find(&history)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return history, nil
}

func (m *defaultL2NftMetadataHistoryModel) GetL2NftMetadataHistory(nftIndex int64) (history *L2NftMetadataHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("nft_index = ?", nftIndex).Limit(1).Find(&history)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return history, nil
}

func (m *defaultL2NftMetadataHistoryModel) GetL2NftMetadataHistoryByHash(txHash string) (history *L2NftMetadataHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_hash = ?", txHash).Find(&history)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return history, nil
}

func (m *defaultL2NftMetadataHistoryModel) UpdateL2NftMetadataHistoryInTransact(history *L2NftMetadataHistory) error {
	dbTx := m.DB.Table(m.table).
		Select("*").Updates(history)
	if dbTx.Error != nil {
		return types.DbErrSqlOperation
	}
	return nil
}

func (m *defaultL2NftMetadataHistoryModel) UpdateL2NftMetadataHistoryNoNftIndex(history *L2NftMetadataHistory) error {
	dbTx := m.DB.Table(m.table).
		Select("*").Where("nft_index = ?", -1).Updates(history)
	if dbTx.Error != nil {
		return types.DbErrSqlOperation
	}
	return nil
}

func (m *defaultL2NftMetadataHistoryModel) DeleteInTransact(id uint) error {
	dbTx := m.DB.Model(&L2NftMetadataHistory{}).Select("DeletedAt", "status").
		Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": time.Now(),
		"status":     StatusFailed,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}
