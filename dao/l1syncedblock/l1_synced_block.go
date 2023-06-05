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

package l1syncedblock

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	TableName = "l1_synced_block"

	TypeGeneric    int = 0
	TypeGovernance int = 1
	TypeDesert     int = 2
)

type (
	L1SyncedBlockModel interface {
		CreateL1SyncedBlockTable() error
		DropL1SyncedBlockTable() error
		GetLatestL1SyncedBlockByType(blockType int) (blockInfo *L1SyncedBlock, err error)
		DeleteL1SyncedBlocksForHeightLessThan(height int64) (err error)
		CreateL1SyncedBlockInTransact(tx *gorm.DB, block *L1SyncedBlock) error
	}

	defaultL1EventModel struct {
		table string
		DB    *gorm.DB
	}

	L1SyncedBlock struct {
		gorm.Model
		// l1 block height
		L1BlockHeight int64 `gorm:"index"`
		// block info, array of hashes
		BlockInfo string
		Type      int `gorm:"index"`
	}
)

func (*L1SyncedBlock) TableName() string {
	return TableName
}

func NewL1SyncedBlockModel(db *gorm.DB) L1SyncedBlockModel {
	return &defaultL1EventModel{
		table: TableName,
		DB:    db,
	}
}

func (m *defaultL1EventModel) CreateL1SyncedBlockTable() error {
	return m.DB.AutoMigrate(L1SyncedBlock{})
}

func (m *defaultL1EventModel) DropL1SyncedBlockTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultL1EventModel) GetLatestL1SyncedBlockByType(blockType int) (blockInfo *L1SyncedBlock, err error) {
	dbTx := m.DB.Table(m.table).Where("type = ?", blockType).Order("l1_block_height desc").Limit(1).Find(&blockInfo)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return blockInfo, nil
}

func (m *defaultL1EventModel) DeleteL1SyncedBlocksForHeightLessThan(height int64) (err error) {
	dbTx := m.DB.Table(m.table).Unscoped().Where("l1_block_height < ?", height).Delete(&L1SyncedBlock{})
	if dbTx.Error != nil {
		return types.DbErrSqlOperation
	}
	return nil
}

func (m *defaultL1EventModel) CreateL1SyncedBlockInTransact(tx *gorm.DB, block *L1SyncedBlock) error {
	dbTx := tx.Table(m.table).Create(block)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToL1SyncedBlock
	}
	return nil
}
