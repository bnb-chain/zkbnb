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

package desertexit

import (
	"errors"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
)

const (
	_ = iota
	StatusCommitted
	StatusVerified
	StatusExecuted
)

const (
	BlockTableName = `desert_exit_block`
)

type (
	DesertExitBlockModel interface {
		CreateDesertExitBlockTable() error
		DropDesertExitBlockTable() error
		GetBlockByHeight(blockHeight int64) (block *DesertExitBlock, err error)
		GetBlocksByHeights(blockHeights []int64) (blocks []*DesertExitBlock, err error)
		BatchInsertOrUpdateInTransact(tx *gorm.DB, desertExitBlocks []*DesertExitBlock) (err error)
		GetBlocksByStatusAndMaxHeight(status int, maxHeight int64, limit int64) (desertExitBlocks []*DesertExitBlock, err error)
		GetLatestExecutedBlock() (desertExitBlock *DesertExitBlock, err error)
		GetLatestBlock() (desertExitBlock *DesertExitBlock, err error)
		UpdateBlockToExecutedInTransact(tx *gorm.DB, desertExitBlock *DesertExitBlock) error
	}

	defaultBlockModel struct {
		table string
		DB    *gorm.DB
	}

	DesertExitBlock struct {
		gorm.Model
		BlockSize         uint16
		BlockHeight       int64 `gorm:"uniqueIndex"`
		PubData           string
		CommittedTxHash   string
		L1CommittedHeight uint64 `gorm:"index"`
		VerifiedTxHash    string
		L1VerifiedHeight  uint64 `gorm:"index"`
		BlockStatus       int64  `gorm:"index"`
	}
)

func NewDesertExitBlockModel(db *gorm.DB) DesertExitBlockModel {
	return &defaultBlockModel{
		table: BlockTableName,
		DB:    db,
	}
}

func (*DesertExitBlock) TableName() string {
	return BlockTableName
}

func (m *defaultBlockModel) CreateDesertExitBlockTable() error {
	return m.DB.AutoMigrate(DesertExitBlock{})
}

func (m *defaultBlockModel) DropDesertExitBlockTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
func (m *defaultBlockModel) GetBlockByHeight(blockHeight int64) (block *DesertExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return block, nil
}

func (m *defaultBlockModel) GetBlocksByHeights(blockHeights []int64) (blocks []*DesertExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height in ?", blockHeights).Find(&blocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return blocks, nil
}

func (m *defaultBlockModel) BatchInsertOrUpdateInTransact(tx *gorm.DB, desertExitBlocks []*DesertExitBlock) (err error) {
	dbTx := tx.Table(m.table).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_size", "block_height", "pub_data", "committed_tx_hash", "verified_tx_hash", "block_status", "l1_committed_height", "l1_verified_height"}),
	}).CreateInBatches(&desertExitBlocks, len(desertExitBlocks))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if int(dbTx.RowsAffected) != len(desertExitBlocks) {
		logx.Errorf("BatchInsertOrUpdateInTransact failed,rows affected not equal desertExitBlocks length,dbTx.RowsAffected:%d,len(desertExitBlocks):%d", int(dbTx.RowsAffected), len(desertExitBlocks))
		return types.DbErrFailToUpdateAccount
	}
	return nil
}

func (m *defaultBlockModel) GetBlocksByStatusAndMaxHeight(status int, maxHeight int64, limit int64) (desertExitBlocks []*DesertExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Where("block_status = ? and block_height > ?", status, maxHeight).Order("block_height asc").Find(&desertExitBlocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return desertExitBlocks, nil
}

func (m *defaultBlockModel) GetLatestExecutedBlock() (desertExitBlock *DesertExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ?", StatusExecuted).Order("block_height DESC").Limit(1).Find(&desertExitBlock)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return desertExitBlock, nil
}
func (m *defaultBlockModel) GetLatestBlock() (desertExitBlock *DesertExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Order("block_height DESC").Limit(1).Find(&desertExitBlock)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return desertExitBlock, nil
}

func (m *defaultBlockModel) UpdateBlockToExecutedInTransact(tx *gorm.DB, desertExitBlock *DesertExitBlock) error {
	dbTx := tx.Model(&DesertExitBlock{}).Select("BlockStatus").Where("id = ? and  block_status = ?", desertExitBlock.ID, StatusVerified).Updates(map[string]interface{}{
		"block_status": StatusExecuted,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != 1 {
		return errors.New("update desertExitBlock status failed,rowsAffected =" + strconv.FormatInt(dbTx.RowsAffected, 10) + "not equal length=1")
	}
	return nil
}
