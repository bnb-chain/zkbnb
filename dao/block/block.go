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

package block

import (
	"sort"

	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

const (
	_ = iota
	StatusProposing
	StatusPending
	StatusCommitted
	StatusVerifiedAndExecuted
)

const (
	BlockTableName = `block`
)

type (
	BlockModel interface {
		CreateBlockTable() error
		DropBlockTable() error
		GetBlocks(limit int64, offset int64) (blocks []*Block, err error)
		GetBlocksBetween(start int64, end int64) (blocks []*Block, err error)
		GetBlockByHeight(blockHeight int64) (block *Block, err error)
		GetBlockByHeightWithoutTx(blockHeight int64) (block *Block, err error)
		GetCommittedBlocksCount() (count int64, err error)
		GetVerifiedBlocksCount() (count int64, err error)
		GetLatestVerifiedHeight() (height int64, err error)
		GetBlockByCommitment(blockCommitment string) (block *Block, err error)
		GetCommittedBlocksBetween(start, end int64) (blocks []*Block, err error)
		GetBlocksTotalCount() (count int64, err error)
		CreateGenesisBlock(block *Block) error
		GetCurrentBlockHeight() (blockHeight int64, err error)
		CreateBlockInTransact(tx *gorm.DB, oBlock *Block) error
		UpdateBlocksWithoutTxsInTransact(tx *gorm.DB, blocks []*Block) (err error)
		UpdateBlockInTransact(tx *gorm.DB, block *Block) (err error)
	}

	defaultBlockModel struct {
		table string
		DB    *gorm.DB
	}

	Block struct {
		gorm.Model
		BlockSize uint16
		// pubdata
		BlockCommitment                 string
		BlockHeight                     int64 `gorm:"uniqueIndex"`
		StateRoot                       string
		PriorityOperations              int64
		PendingOnChainOperationsHash    string
		PendingOnChainOperationsPubData string
		CommittedTxHash                 string
		CommittedAt                     int64
		VerifiedTxHash                  string
		VerifiedAt                      int64
		Txs                             []*tx.Tx `gorm:"foreignKey:BlockId"`
		BlockStatus                     int64
	}

	BlockStates struct {
		Block           *Block
		CompressedBlock *compressedblock.CompressedBlock

		PendingNewAccount        []*account.Account
		PendingUpdateAccount     []*account.Account
		PendingNewAccountHistory []*account.AccountHistory
		PendingNewNft            []*nft.L2Nft
		PendingUpdateNft         []*nft.L2Nft
		PendingNewNftHistory     []*nft.L2NftHistory
	}
)

func NewBlockModel(db *gorm.DB) BlockModel {
	return &defaultBlockModel{
		table: BlockTableName,
		DB:    db,
	}
}

func (*Block) TableName() string {
	return BlockTableName
}

func (b *Block) ClearTxsModel() {
	for _, blockTx := range b.Txs {
		blockTx.Model = gorm.Model{}
	}
}

func (m *defaultBlockModel) CreateBlockTable() error {
	return m.DB.AutoMigrate(Block{})
}

func (m *defaultBlockModel) DropBlockTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultBlockModel) GetBlocks(limit int64, offset int64) (blocks []*Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)

	dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("block_height desc").Find(&blocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}

	for _, block := range blocks {
		err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
		if err != nil {
			return nil, types.DbErrSqlOperation
		}
		sort.Slice(block.Txs, func(i, j int) bool {
			return block.Txs[i].TxIndex < block.Txs[j].TxIndex
		})
	}

	return blocks, nil
}

func (m *defaultBlockModel) GetBlocksBetween(start int64, end int64) (blocks []*Block, err error) {
	var (
		txForeignKeyColumn        = `Txs`
		txDetailsForeignKeyColumn = `TxDetails`
	)
	dbTx := m.DB.Table(m.table).Where("block_height >= ? AND block_height <= ?", start, end).
		Order("block_height").
		Find(&blocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}

	for index, block := range blocks {
		// If the last block is proposing, skip it.
		if index == len(blocks)-1 && block.BlockStatus <= StatusProposing {
			blocks = blocks[:len(blocks)-1]
			break
		}

		err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
		if err != nil {
			return nil, types.DbErrSqlOperation
		}
		sort.Slice(block.Txs, func(i, j int) bool {
			return block.Txs[i].TxIndex < block.Txs[j].TxIndex
		})

		for _, txInfo := range block.Txs {
			err = m.DB.Model(&txInfo).Association(txDetailsForeignKeyColumn).Find(&txInfo.TxDetails)
			if err != nil {
				return nil, types.DbErrSqlOperation
			}
			sort.Slice(txInfo.TxDetails, func(i, j int) bool {
				return txInfo.TxDetails[i].Order < txInfo.TxDetails[j].Order
			})
		}
	}
	return blocks, nil
}

func (m *defaultBlockModel) GetBlockByCommitment(blockCommitment string) (block *Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)
	dbTx := m.DB.Table(m.table).Where("block_commitment = ?", blockCommitment).Find(&block)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	sort.Slice(block.Txs, func(i, j int) bool {
		return block.Txs[i].TxIndex < block.Txs[j].TxIndex
	})
	if err != nil {
		return nil, types.DbErrSqlOperation
	}
	return block, nil
}

func (m *defaultBlockModel) GetBlockByHeight(blockHeight int64) (block *Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	sort.Slice(block.Txs, func(i, j int) bool {
		return block.Txs[i].TxIndex < block.Txs[j].TxIndex
	})
	if err != nil {
		return nil, types.DbErrSqlOperation
	}

	return block, nil
}

func (m *defaultBlockModel) GetBlockByHeightWithoutTx(blockHeight int64) (block *Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return block, nil
}

func (m *defaultBlockModel) GetCommittedBlocksCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status >= ? and deleted_at is NULL", StatusCommitted).Count(&count)
	if dbTx.Error != nil {
		if dbTx.Error == types.DbErrNotFound {
			return 0, nil
		}
		return 0, types.DbErrSqlOperation
	}

	return count, nil
}

func (m *defaultBlockModel) GetVerifiedBlocksCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ? and deleted_at is NULL", StatusVerifiedAndExecuted).Count(&count)
	if dbTx.Error != nil {
		if dbTx.Error == types.DbErrNotFound {
			return 0, nil
		}
		return 0, types.DbErrSqlOperation
	}
	return count, nil
}

func (m *defaultBlockModel) CreateGenesisBlock(block *Block) error {
	dbTx := m.DB.Table(m.table).Omit("BlockDetails").Omit("Txs").Create(block)

	if dbTx.Error != nil {
		return types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateBlock
	}
	return nil
}

func (m *defaultBlockModel) GetCurrentBlockHeight() (blockHeight int64, err error) {
	dbTx := m.DB.Table(m.table).Select("block_height").Order("block_height desc").Limit(1).Find(&blockHeight)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return blockHeight, nil
}

func (m *defaultBlockModel) GetBlocksTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultBlockModel) GetCommittedBlocksBetween(start, end int64) (blocks []*Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ? AND block_height >= ? AND block_height <= ?", StatusCommitted, start, end).
		Order("block_height").
		Find(&blocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return blocks, nil
}

func (m *defaultBlockModel) GetLatestVerifiedHeight() (height int64, err error) {
	block := &Block{}
	dbTx := m.DB.Table(m.table).Where("block_status = ?", StatusVerifiedAndExecuted).
		Order("block_height DESC").
		Limit(1).
		First(&block)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return block.BlockHeight, nil
}

func (m *defaultBlockModel) CreateBlockInTransact(tx *gorm.DB, oBlock *Block) (err error) {
	dbTx := tx.Table(m.table).Create(oBlock)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		if err != nil {
			return err
		}
		return types.DbErrFailToCreateBlock
	}

	return nil
}

func (m *defaultBlockModel) UpdateBlocksWithoutTxsInTransact(tx *gorm.DB, blocks []*Block) (err error) {
	const Txs = "Txs"

	for _, block := range blocks {
		dbTx := tx.Table(m.table).Where("id = ?", block.ID).
			Omit(Txs).
			Select("*").
			Updates(&block)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			if err != nil {
				return err
			}
			return types.DbErrFailToUpdateBlock
		}
	}
	return nil
}

func (m *defaultBlockModel) UpdateBlockInTransact(tx *gorm.DB, block *Block) (err error) {
	dbTx := tx.Table(m.table).Where("id = ?", block.ID).
		Select("*").
		Updates(&block)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		if err != nil {
			return err
		}
		return types.DbErrFailToUpdateBlock
	}
	return nil
}
