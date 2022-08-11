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

package blockForProof

import (
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/util"
)

type (
	BlockForProofModel interface {
		CreateBlockForProofTable() error
		DropBlockForProofTable() error
		GetLatestUnprovedBlockHeight() (blockNumber int64, err error)
		GetUnprovedCryptoBlockByBlockNumber(height int64) (block *BlockForProof, err error)
		UpdateUnprovedCryptoBlockStatus(block *BlockForProof, status int64) error
		GetUnprovedCryptoBlockByMode(mode int64) (block *BlockForProof, err error)
		CreateConsecutiveUnprovedCryptoBlock(block *BlockForProof) error
	}

	defaultBlockForProofModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	BlockForProof struct {
		gorm.Model
		BlockHeight int64 `gorm:"index:idx_height,unique"`
		BlockData   string
		Status      int64
	}
)

func NewBlockForProofModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) BlockForProofModel {
	return &defaultBlockForProofModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      BlockForProofTableName,
		DB:         db,
	}
}

func (*BlockForProof) TableName() string {
	return BlockForProofTableName
}

/*
	Func: CreateBlockForProofTable
	Params:
	Return: err error
	Description: create Block table
*/

func (m *defaultBlockForProofModel) CreateBlockForProofTable() error {
	return m.DB.AutoMigrate(BlockForProof{})
}

/*
	Func: DropBlockForProofTable
	Params:
	Return: err error
	Description: drop block table
*/

func (m *defaultBlockForProofModel) DropBlockForProofTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultBlockForProofModel) GetLatestUnprovedBlockHeight() (blockNumber int64, err error) {
	var row *BlockForProof
	dbTx := m.DB.Table(m.table).Order("block_height desc").Limit(1).Find(&row)
	if dbTx.Error != nil {
		logx.Errorf("unable to get latest unproved block: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, errorcode.DbErrNotFound
	}
	return row.BlockHeight, nil
}

func (m *defaultBlockForProofModel) GetUnprovedCryptoBlockByMode(mode int64) (block *BlockForProof, err error) {
	switch mode {
	case util.COO_MODE:
		dbTx := m.DB.Table(m.table).Where("status = ?", StatusPublished).Order("block_height asc").Limit(1).Find(&block)
		if dbTx.Error != nil {
			logx.Errorf("unable to get unproved block: %s", dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		return block, nil
	case util.COM_MODE:
		dbTx := m.DB.Table(m.table).Where("status <= ?", StatusReceived).Order("block_height asc").Limit(1).Find(&block)
		if dbTx.Error != nil {
			logx.Errorf("unable to get unproved block: %s", dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		return block, nil
	default:
		return nil, nil
	}
}

func (m *defaultBlockForProofModel) GetUnprovedCryptoBlockByBlockNumber(height int64) (block *BlockForProof, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height = ?", height).Limit(1).Find(&block)
	if dbTx.Error != nil {
		logx.Errorf("unable to get unproved block: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return block, nil
}

func (m *defaultBlockForProofModel) CreateConsecutiveUnprovedCryptoBlock(block *BlockForProof) error {
	if block.BlockHeight > 1 {
		_, err := m.GetUnprovedCryptoBlockByBlockNumber(block.BlockHeight - 1)
		if err != nil {
			logx.Infof("get unproved block error, err: %s", err.Error())
			return fmt.Errorf("previous block does not exist")
		}
	}

	dbTx := m.DB.Table(m.table).Create(block)
	if dbTx.Error != nil {
		logx.Errorf("create block error, err: %s", dbTx.Error.Error())
		return errorcode.DbErrSqlOperation
	}
	return nil
}

func (m *defaultBlockForProofModel) UpdateUnprovedCryptoBlockStatus(block *BlockForProof, status int64) error {
	block.Status = status
	block.UpdatedAt = time.Now()
	dbTx := m.DB.Table(m.table).Save(block)
	if dbTx.Error != nil {
		logx.Errorf("update block status error: %s", dbTx.Error.Error())
		return errorcode.DbErrSqlOperation
	}
	return nil
}
