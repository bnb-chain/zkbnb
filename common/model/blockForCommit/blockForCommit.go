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

package blockForCommit

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	BlockForCommitModel interface {
		CreateBlockForCommitTable() error
		DropBlockForCommitTable() error
		GetBlockForCommitByHeight(height int64) (blockForCommit *BlockForCommit, err error)
		GetBlockForCommitBetween(start, end int64) (blocksForCommit []*BlockForCommit, err error)
	}

	defaultBlockForCommitModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	BlockForCommit struct {
		gorm.Model
		BlockHeight       int64
		StateRoot         string
		PublicData        string
		Timestamp         int64
		PublicDataOffsets string
	}
)

func NewBlockForCommitModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) BlockForCommitModel {
	return &defaultBlockForCommitModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TableName,
		DB:         db,
	}
}

func (*BlockForCommit) TableName() string {
	return TableName
}

/*
	Func: CreateBlockForCommitTable
	Params:
	Return: err error
	Description: create Block table
*/

func (m *defaultBlockForCommitModel) CreateBlockForCommitTable() error {
	return m.DB.AutoMigrate(BlockForCommit{})
}

/*
	Func: DropBlockForCommitTable
	Params:
	Return: err error
	Description: drop block table
*/

func (m *defaultBlockForCommitModel) DropBlockForCommitTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultBlockForCommitModel) GetBlockForCommitByHeight(height int64) (blockForCommit *BlockForCommit, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height = ?", height).Find(&blockForCommit)
	if dbTx.Error != nil {
		logx.Errorf("[GetBlockForCommitBetween] unable to get block for commit by height: %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return blockForCommit, nil
}

func (m *defaultBlockForCommitModel) GetBlockForCommitBetween(start, end int64) (blocksForCommit []*BlockForCommit, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height >= ? AND block_height <= ?", start, end).Find(&blocksForCommit)
	if dbTx.Error != nil {
		logx.Errorf("[GetBlockForCommitBetween] unable to get block for commit between: %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return blocksForCommit, nil
}
