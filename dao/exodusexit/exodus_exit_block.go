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

package exodusexit

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
	BlockTableName = `exodus_exit_block`
)

type (
	ExodusExitBlockModel interface {
		CreateExodusExitBlockTable() error
		DropExodusExitBlockTable() error
		GetBlockByHeight(blockHeight int64) (block *ExodusExitBlock, err error)
		GetBlocksByHeights(blockHeights []int64) (blocks []*ExodusExitBlock, err error)
		BatchInsertOrUpdateInTransact(tx *gorm.DB, exodusExitBlocks []*ExodusExitBlock) (err error)
		GetBlocksByStatusAndMaxHeight(status int, maxHeight int64, limit int64) (exodusExitBlocks []*ExodusExitBlock, err error)
		GetLatestExecutedBlock() (exodusExitBlock *ExodusExitBlock, err error)
		GetLatestBlock() (exodusExitBlock *ExodusExitBlock, err error)
		UpdateBlockToExecutedInTransact(tx *gorm.DB, exodusExitBlock *ExodusExitBlock) error
	}

	defaultBlockModel struct {
		table string
		DB    *gorm.DB
	}

	ExodusExitBlock struct {
		gorm.Model
		BlockSize       uint16
		BlockHeight     int64 `gorm:"uniqueIndex"`
		PubData         string
		CommittedTxHash string
		CommittedAt     int64
		VerifiedTxHash  string
		VerifiedAt      int64
		BlockStatus     int64 `gorm:"index"`
	}
)

func NewExodusExitBlockModel(db *gorm.DB) ExodusExitBlockModel {
	return &defaultBlockModel{
		table: BlockTableName,
		DB:    db,
	}
}

func (*ExodusExitBlock) TableName() string {
	return BlockTableName
}

func (m *defaultBlockModel) CreateExodusExitBlockTable() error {
	return m.DB.AutoMigrate(ExodusExitBlock{})
}

func (m *defaultBlockModel) DropExodusExitBlockTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
func (m *defaultBlockModel) GetBlockByHeight(blockHeight int64) (block *ExodusExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return block, nil
}

func (m *defaultBlockModel) GetBlocksByHeights(blockHeights []int64) (blocks []*ExodusExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height in ?", blockHeights).Find(&blocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return blocks, nil
}

func (m *defaultBlockModel) BatchInsertOrUpdateInTransact(tx *gorm.DB, exodusExitBlocks []*ExodusExitBlock) (err error) {
	dbTx := tx.Table(m.table).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_size", "block_height", "pub_data", "committed_tx_hash", "committed_at", "verified_tx_hash", "verified_at", "block_status"}),
	}).CreateInBatches(&exodusExitBlocks, len(exodusExitBlocks))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if int(dbTx.RowsAffected) != len(exodusExitBlocks) {
		logx.Errorf("BatchInsertOrUpdateInTransact failed,rows affected not equal exodusExitBlocks length,dbTx.RowsAffected:%s,len(exodusExitBlocks):%s", int(dbTx.RowsAffected), len(exodusExitBlocks))
		return types.DbErrFailToUpdateAccount
	}
	return nil
}

func (m *defaultBlockModel) GetBlocksByStatusAndMaxHeight(status int, maxHeight int64, limit int64) (exodusExitBlocks []*ExodusExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Where("block_status = ? and block_height > ?", status, maxHeight).Order("block_height asc").Find(&exodusExitBlocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return exodusExitBlocks, nil
}

func (m *defaultBlockModel) GetLatestExecutedBlock() (exodusExitBlock *ExodusExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ?", StatusExecuted).Order("block_height DESC").Limit(1).Find(&exodusExitBlock)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return exodusExitBlock, nil
}
func (m *defaultBlockModel) GetLatestBlock() (exodusExitBlock *ExodusExitBlock, err error) {
	dbTx := m.DB.Table(m.table).Order("block_height DESC").Limit(1).Find(&exodusExitBlock)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return exodusExitBlock, nil
}

func (m *defaultBlockModel) UpdateBlockToExecutedInTransact(tx *gorm.DB, exodusExitBlock *ExodusExitBlock) error {
	dbTx := tx.Model(&ExodusExitBlock{}).Select("BlockStatus").Where("id = ? and  block_status = ?", exodusExitBlock.ID, StatusVerified).Updates(map[string]interface{}{
		"block_status": StatusExecuted,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != 1 {
		return errors.New("update exodusExitBlock status failed,rowsAffected =" + strconv.FormatInt(dbTx.RowsAffected, 10) + "not equal length=1")
	}
	return nil
}
