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

package compressedblock

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	CompressedBlockTableName = `compressed_block`
)

type (
	CompressedBlockModel interface {
		CreateCompressedBlockTable() error
		DropCompressedBlockTable() error
		GetCompressedBlocksBetween(start, end int64) (blocksForCommit []*CompressedBlock, err error)
		CreateCompressedBlock(block *CompressedBlock) error
		DeleteByHeightInTransact(tx *gorm.DB, heights []int) error
	}

	defaultCompressedBlockModel struct {
		table string
		DB    *gorm.DB
	}

	CompressedBlock struct {
		gorm.Model
		BlockSize         uint16
		BlockHeight       int64 `gorm:"index"`
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
	dbTx := m.DB.Table(m.table).Where("block_height >= ? AND block_height <= ?", start, end).Order("block_height").Find(&blocksForCommit)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return blocksForCommit, nil
}

func (m *defaultCompressedBlockModel) CreateCompressedBlock(block *CompressedBlock) error {
	dbTx := m.DB.Table(m.table).Create(block)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateCompressedBlock
	}
	return nil
}
func (m *defaultCompressedBlockModel) DeleteByHeightInTransact(tx *gorm.DB, heights []int) error {
	dbTx := tx.Model(&CompressedBlock{}).Unscoped().Where("block_height in ?", heights).Delete(&CompressedBlock{})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}
