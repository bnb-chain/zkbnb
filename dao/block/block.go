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
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"sort"
	"strconv"
	"time"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/types"
)

const (
	_ = iota
	StatusProposing
	StatusPacked
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
		GetPendingBlocksBetween(start int64, end int64) (blocks []*Block, err error)
		GetPendingBlocksBetweenWithoutTx(start int64, end int64) (blocks []*Block, err error)
		GetBlockByHeight(blockHeight int64) (block *Block, err error)
		GetBlockByHeightWithoutTx(blockHeight int64) (block *Block, err error)
		GetCommittedBlocksCount() (count int64, err error)
		GetVerifiedBlocksCount() (count int64, err error)
		GetLatestVerifiedHeight() (height int64, err error)
		GetLatestCommittedHeight() (height int64, err error)
		GetBlockByCommitment(blockCommitment string) (block *Block, err error)
		GetCommittedBlocksBetween(start, end int64) (blocks []*Block, err error)
		GetBlocksTotalCount() (count int64, err error)
		CreateGenesisBlock(block *Block) error
		GetCurrentBlockHeight() (blockHeight int64, err error)
		GetCurrentBlockHeightInTransact(tx *gorm.DB) (blockHeight int64, err error)
		CreateBlockInTransact(tx *gorm.DB, oBlock *Block) error
		UpdateBlocksWithoutTxsInTransact(tx *gorm.DB, blocks []*Block) (err error)
		UpdateBlockInTransact(tx *gorm.DB, block *Block) (err error)
		DeleteBlockInTransact(tx *gorm.DB, statuses []int) error
		DeleteBlockGreaterThanHeight(blockHeight int64, statuses []int) error
		GetProposingBlockHeights() (blockHeights []int64, err error)
		PreSaveBlockDataInTransact(tx *gorm.DB, block *Block) error
		UpdateBlockToPendingInTransact(tx *gorm.DB, block *Block) error
		GetBlockByStatus(statuses []int) (blocks []*Block, err error)
		GetLatestHeight(statuses []int) (height int64, err error)
		UpdateBlockToProposingInTransact(tx *gorm.DB, blockHeights []int64) error
		UpdateGreaterOrEqualHeight(blockHeight int64, targetBlockStatus int64) error
		GetBlockByStatusAndTime(status int, time time.Time) (block *Block, err error)
	}

	defaultBlockModel struct {
		table string
		DB    *gorm.DB
	}

	Block struct {
		gorm.Model
		BlockSize uint16
		// pubdata
		BlockCommitment                 string `gorm:"index"`
		BlockHeight                     int64  `gorm:"uniqueIndex"`
		StateRoot                       string
		PriorityOperations              int64
		PendingOnChainOperationsHash    string
		PendingOnChainOperationsPubData string
		CommittedTxHash                 string
		CommittedAt                     int64
		VerifiedTxHash                  string
		VerifiedAt                      int64
		Txs                             []*tx.Tx `gorm:"-"`
		BlockStatus                     int64    `gorm:"index"`
		AccountIndexes                  string
		NftIndexes                      string
	}

	BlockStates struct {
		Block           *Block
		CompressedBlock *compressedblock.CompressedBlock

		PendingAccount        []*account.Account
		PendingAccountHistory []*account.AccountHistory
		PendingNft            []*nft.L2Nft
		PendingNftHistory     []*nft.L2NftHistory
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
		createdAt := b.CreatedAt
		blockTx.Model = gorm.Model{}
		blockTx.CreatedAt = createdAt
	}
}

func (m *defaultBlockModel) CreateBlockTable() error {
	return m.DB.AutoMigrate(Block{})
}

func (m *defaultBlockModel) DropBlockTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultBlockModel) GetBlocks(limit int64, offset int64) (blocks []*Block, err error) {

	dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("block_height desc").Find(&blocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}

	for _, block := range blocks {
		dbTx := m.DB.Table(tx.TxTableName).Where("block_height =? ", block.BlockHeight).Find(&block.Txs)
		if dbTx.Error != nil {
			return nil, types.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			continue
		}
		sort.Slice(block.Txs, func(i, j int) bool {
			return block.Txs[i].TxIndex < block.Txs[j].TxIndex
		})
	}

	return blocks, nil
}

func (m *defaultBlockModel) GetPendingBlocksBetween(start int64, end int64) (blocks []*Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height >= ? AND block_height <= ? and block_status in ?", start, end, []int{StatusPending, StatusCommitted}).
		Order("block_height").
		Find(&blocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}

	for _, block := range blocks {
		dbTx = m.DB.Table(tx.TxTableName).Where("block_height =? ", block.BlockHeight).Find(&block.Txs)
		if dbTx.Error != nil {
			return nil, types.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			continue
		}
		sort.Slice(block.Txs, func(i, j int) bool {
			return block.Txs[i].TxIndex < block.Txs[j].TxIndex
		})

		for _, txInfo := range block.Txs {
			dbTx = m.DB.Table(tx.TxDetailTableName).Where("pool_tx_id=?", txInfo.PoolTxId).Find(&txInfo.TxDetails)
			if dbTx.Error != nil {
				return nil, types.DbErrSqlOperation
			} else if dbTx.RowsAffected == 0 {
				continue
			}
			sort.Slice(txInfo.TxDetails, func(i, j int) bool {
				return txInfo.TxDetails[i].Order < txInfo.TxDetails[j].Order
			})
		}
	}
	return blocks, nil
}

func (m *defaultBlockModel) GetPendingBlocksBetweenWithoutTx(start int64, end int64) (blocks []*Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height >= ? AND block_height <= ? and block_status in ?", start, end, []int{StatusPending, StatusCommitted}).
		Order("block_height").
		Find(&blocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return blocks, nil
}

func (m *defaultBlockModel) GetBlockByCommitment(blockCommitment string) (block *Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_commitment = ?", blockCommitment).Find(&block)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	dbTx = m.DB.Table(tx.TxTableName).Where("block_height =? ", block.BlockHeight).Find(&block.Txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return block, nil
	}
	sort.Slice(block.Txs, func(i, j int) bool {
		return block.Txs[i].TxIndex < block.Txs[j].TxIndex
	})

	return block, nil
}

func (m *defaultBlockModel) GetBlockByHeight(blockHeight int64) (block *Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	dbTx = m.DB.Table(tx.TxTableName).Where("block_height =? ", block.BlockHeight).Find(&block.Txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return block, nil
	}
	sort.Slice(block.Txs, func(i, j int) bool {
		return block.Txs[i].TxIndex < block.Txs[j].TxIndex
	})
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

func (m *defaultBlockModel) GetCurrentBlockHeightInTransact(tx *gorm.DB) (blockHeight int64, err error) {
	dbTx := tx.Table(m.table).Select("block_height").Order("block_height desc").Limit(1).Find(&blockHeight)
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
		Find(&block)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return block.BlockHeight, nil
}

func (m *defaultBlockModel) GetLatestCommittedHeight() (height int64, err error) {
	block := &Block{}
	dbTx := m.DB.Table(m.table).Where("block_status = ?", StatusCommitted).
		Order("block_height DESC").
		Limit(1).
		Find(&block)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return block.BlockHeight, nil
}

func (m *defaultBlockModel) GetLatestHeight(statuses []int) (height int64, err error) {
	block := &Block{}
	dbTx := m.DB.Table(m.table).Where("block_status in ?", statuses).
		Order("block_height DESC").
		Limit(1).
		Find(&block)
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
		return types.DbErrFailToCreateBlock
	}

	return nil
}

func (m *defaultBlockModel) UpdateBlocksWithoutTxsInTransact(tx *gorm.DB, blocks []*Block) (err error) {
	for _, block := range blocks {
		dbTx := tx.Table(m.table).Where("id = ?", block.ID).
			Select("*").
			Updates(&block)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
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
		return types.DbErrFailToUpdateBlock
	}
	return nil
}

func (m *defaultBlockModel) DeleteBlockInTransact(tx *gorm.DB, statuses []int) error {
	if len(statuses) == 0 {
		return nil
	}
	dbTx := tx.Table(m.table).Unscoped().Where("block_status in ?", statuses).Delete(&Block{})
	if dbTx.Error != nil {
		return types.DbErrSqlOperation
	}
	return nil
}

func (m *defaultBlockModel) DeleteBlockGreaterThanHeight(blockHeight int64, statuses []int) error {
	dbTx := m.DB.Table(m.table).Unscoped().Where("block_status in ? and block_height > ?", statuses, blockHeight).Delete(&Block{})
	if dbTx.Error != nil {
		return types.DbErrSqlOperation
	}
	return nil
}

func (m *defaultBlockModel) GetProposingBlockHeights() (blockHeights []int64, err error) {
	dbTx := m.DB.Table(m.table).Select("block_height").Where("block_status = ?", StatusProposing).Order("block_height desc").Find(&blockHeights)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return blockHeights, nil
}

func (m *defaultBlockModel) GetBlockByStatus(statuses []int) (blocks []*Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status in ?", statuses).Order("block_height").Find(&blocks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return blocks, nil
}

func (m *defaultBlockModel) PreSaveBlockDataInTransact(tx *gorm.DB, block *Block) (err error) {
	dbTx := tx.Model(&Block{}).Select("BlockStatus", "AccountIndexes", "NftIndexes", "CreatedAt").Where("id = ? and  block_status in ?", block.ID, []int{StatusProposing, StatusPacked}).Updates(map[string]interface{}{
		"block_status":    StatusPacked,
		"account_indexes": block.AccountIndexes,
		"nft_indexes":     block.NftIndexes,
		"created_at":      block.CreatedAt,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToUpdateBlock
	}
	return nil
}

func (m *defaultBlockModel) UpdateBlockToPendingInTransact(tx *gorm.DB, block *Block) error {
	dbTx := tx.Model(&Block{}).Select("BlockStatus", "BlockSize", "BlockCommitment", "StateRoot", "PriorityOperations", "PendingOnChainOperationsHash", "PendingOnChainOperationsPubData", "CreatedAt").Where("id = ? and  block_status = ?", block.ID, StatusPacked).Updates(map[string]interface{}{
		"block_status":                         StatusPending,
		"block_size":                           block.BlockSize,
		"block_commitment":                     block.BlockCommitment,
		"state_root":                           block.StateRoot,
		"priority_operations":                  block.PriorityOperations,
		"pending_on_chain_operations_hash":     block.PendingOnChainOperationsHash,
		"pending_on_chain_operations_pub_data": block.PendingOnChainOperationsPubData,
		"created_at":                           block.CreatedAt,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != 1 {
		logx.Errorf("update block status failed,rowsAffected = %d, not equal length = 1", strconv.FormatInt(dbTx.RowsAffected, 10))
		return types.AppErrFailUpdateBlockStatus
	}
	return nil
}

func (m *defaultBlockModel) UpdateBlockToProposingInTransact(tx *gorm.DB, blockHeights []int64) error {
	dbTx := tx.Model(&Block{}).Select("BlockStatus", "BlockSize", "BlockCommitment", "StateRoot", "PriorityOperations", "PendingOnChainOperationsHash", "PendingOnChainOperationsPubData", "CommittedTxHash", "CommittedAt", "AccountIndexes", "NftIndexes").Where("block_height in ?", blockHeights).Updates(map[string]interface{}{
		"block_status":                         StatusProposing,
		"block_size":                           0,
		"block_commitment":                     "",
		"state_root":                           "",
		"priority_operations":                  0,
		"pending_on_chain_operations_hash":     "",
		"pending_on_chain_operations_pub_data": "",
		"committed_tx_hash":                    "",
		"committed_at":                         0,
		"account_indexes":                      "[]",
		"nft_indexes":                          "[]",
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}

func (m *defaultBlockModel) UpdateGreaterOrEqualHeight(blockHeight int64, targetBlockStatus int64) error {
	dbTx := m.DB.Model(&Block{}).Select("BlockStatus").Where("block_height >= ? and block_status != ?", blockHeight, StatusVerifiedAndExecuted).Updates(map[string]interface{}{
		"block_status": targetBlockStatus,
	})
	if dbTx.Error != nil {
		return dbTx.Error
	}
	return nil
}

func (m *defaultBlockModel) GetBlockByStatusAndTime(status int, time time.Time) (block *Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ? and created_at < ?", status, time).
		Order("created_at").
		Limit(1).
		Find(&block)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, types.DbErrFailToUpdateBlock
	}
	return block, nil
}
