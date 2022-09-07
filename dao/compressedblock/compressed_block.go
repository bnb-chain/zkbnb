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

package compressedblock

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/types"
)

const (
	CompressedBlockTableName = `compressed_block`
)

type (
	CompressedBlockModel interface {
		CreateCompressedBlockTable() error
		DropCompressedBlockTable() error
		GetCompressedBlocksBetween(start, end int64) (blocksForCommit []*CompressedBlock, err error)
		CreateCompressedBlockInTransact(tx *gorm.DB, block *CompressedBlock) error
	}

	defaultCompressedBlockModel struct {
		table string
		DB    *gorm.DB
	}

	CompressedBlock struct {
		gorm.Model
		BlockSize         uint16
		BlockHeight       int64
		StateRoot         string
		PublicData        string
		Timestamp         int64
		PublicDataOffsets string
	}
)

func NewCompressedBlockModel(db *gorm.DB) CompressedBlockModel {
	return &defaultCompressedBlockModel{
		table: CompressedBlockTableName,
		DB:    db,
	}
}

func (*CompressedBlock) TableName() string {
	return CompressedBlockTableName
}

func (m *defaultCompressedBlockModel) CreateCompressedBlockTable() error {
	return m.DB.AutoMigrate(CompressedBlock{})
}

func (m *defaultCompressedBlockModel) DropCompressedBlockTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultCompressedBlockModel) GetCompressedBlocksBetween(start, end int64) (blocksForCommit []*CompressedBlock, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height >= ? AND block_height <= ?", start, end).Find(&blocksForCommit)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return blocksForCommit, nil
}

func (m *defaultCompressedBlockModel) CreateCompressedBlockInTransact(tx *gorm.DB, block *CompressedBlock) error {
	dbTx := tx.Table(m.table).Create(block)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateCompressedBlock
	}
	return nil
}
