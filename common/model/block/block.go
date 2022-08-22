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

package block

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
)

//goland:noinspection GoNameStartsWithPackageName,GoNameStartsWithPackageName
type (
	BlockModel interface {
		CreateBlockTable() error
		DropBlockTable() error
		GetBlocksList(limit int64, offset int64) (blocks []*Block, err error)
		GetBlocksBetween(start int64, end int64) (blocks []*Block, err error)
		GetBlockByHeight(blockHeight int64) (block *Block, err error)
		GetBlockByHeightWithoutTx(blockHeight int64) (block *Block, err error)
		GetCommittedBlocksCount() (count int64, err error)
		GetVerifiedBlocksCount() (count int64, err error)
		GetLatestVerifiedHeight() (height int64, err error)
		GetBlockByCommitment(blockCommitment string) (block *Block, err error)
		GetBlocksForProverBetween(start, end int64) (blocks []*Block, err error)
		GetBlocksTotalCount() (count int64, err error)
		CreateGenesisBlock(block *Block) error
		GetCurrentHeight() (blockHeight int64, err error)
		CreateNewBlock(oBlock *Block) (err error)
		CreateBlockForCommitter(pendingMempoolTxs []*mempool.MempoolTx, blockStates *BlockStates) error
	}

	defaultBlockModel struct {
		sqlc.CachedConn
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
		Block          *Block
		BlockForCommit *blockForCommit.BlockForCommit

		PendingNewAccount            []*account.Account
		PendingUpdateAccount         []*account.Account
		PendingNewAccountHistory     []*account.AccountHistory
		PendingNewLiquidity          []*liquidity.Liquidity
		PendingUpdateLiquidity       []*liquidity.Liquidity
		PendingNewLiquidityHistory   []*liquidity.LiquidityHistory
		PendingNewNft                []*nft.L2Nft
		PendingUpdateNft             []*nft.L2Nft
		PendingNewNftHistory         []*nft.L2NftHistory
		PendingNewNftWithdrawHistory []*nft.L2NftWithdrawHistory

		PendingNewOffer         []*nft.Offer
		PendingNewL2NftExchange []*nft.L2NftExchange
	}
)

func NewBlockModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) BlockModel {
	return &defaultBlockModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      BlockTableName,
		DB:         db,
	}
}

func (*Block) TableName() string {
	return BlockTableName
}

/*
	Func: CreateBlockTable
	Params:
	Return: err error
	Description: create Block table
*/

func (m *defaultBlockModel) CreateBlockTable() error {
	return m.DB.AutoMigrate(Block{})
}

/*
	Func: DropBlockTable
	Params:
	Return: err error
	Description: drop block table
*/

func (m *defaultBlockModel) DropBlockTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultBlockModel) GetBlocksList(limit int64, offset int64) (blocks []*Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)

	dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("block_height desc").Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("get blocks error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}

	for _, block := range blocks {
		err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
		if err != nil {
			logx.Errorf("get associate txs error, err: %s", err.Error())
			return nil, errorcode.DbErrSqlOperation
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
		logx.Errorf("get bocks error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}

	for _, block := range blocks {
		err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
		if err != nil {
			logx.Errorf("get associate txs error, err: %s", err.Error())
			return nil, errorcode.DbErrSqlOperation
		}
		sort.Slice(block.Txs, func(i, j int) bool {
			return block.Txs[i].TxIndex < block.Txs[j].TxIndex
		})

		for _, txInfo := range block.Txs {
			err = m.DB.Model(&txInfo).Association(txDetailsForeignKeyColumn).Find(&txInfo.TxDetails)
			if err != nil {
				logx.Errorf("get associate tx details error, err: %s", err.Error())
				return nil, errorcode.DbErrSqlOperation
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
		logx.Errorf("get block by commitment error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	sort.Slice(block.Txs, func(i, j int) bool {
		return block.Txs[i].TxIndex < block.Txs[j].TxIndex
	})
	if err != nil {
		logx.Errorf("get associate txs error, err: %s", err.Error())
		return nil, errorcode.DbErrSqlOperation
	}
	return block, nil
}

func (m *defaultBlockModel) GetBlockByHeight(blockHeight int64) (block *Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		logx.Errorf("get block by height error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	sort.Slice(block.Txs, func(i, j int) bool {
		return block.Txs[i].TxIndex < block.Txs[j].TxIndex
	})
	if err != nil {
		logx.Errorf("get associate txs error, err: %s", err.Error())
		return nil, errorcode.DbErrSqlOperation
	}

	return block, nil
}

func (m *defaultBlockModel) GetBlockByHeightWithoutTx(blockHeight int64) (block *Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		logx.Errorf("get block by height error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return block, nil
}

func (m *defaultBlockModel) GetCommittedBlocksCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status >= ? and deleted_at is NULL", StatusCommitted).Count(&count)
	if dbTx.Error != nil {
		if dbTx.Error == errorcode.DbErrNotFound {
			return 0, nil
		}
		logx.Errorf("get committed block count error, err: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	}

	return count, nil
}

func (m *defaultBlockModel) GetVerifiedBlocksCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ? and deleted_at is NULL", StatusVerifiedAndExecuted).Count(&count)
	if dbTx.Error != nil {
		if dbTx.Error == errorcode.DbErrNotFound {
			return 0, nil
		}
		logx.Errorf("get verified block count error, err: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	}
	return count, nil
}

func (m *defaultBlockModel) CreateGenesisBlock(block *Block) error {
	dbTx := m.DB.Table(m.table).Omit("BlockDetails").Omit("Txs").Create(block)

	if dbTx.Error != nil {
		logx.Errorf("create genesis block error, err: %s", dbTx.Error.Error())
		return errorcode.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return errorcode.DbErrFailToCreateBlock
	}
	return nil
}

func (m *defaultBlockModel) GetCurrentHeight() (blockHeight int64, err error) {
	dbTx := m.DB.Table(m.table).Select("block_height").Order("block_height desc").Limit(1).Find(&blockHeight)
	if dbTx.Error != nil {
		logx.Errorf("get current block error, err: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, errorcode.DbErrNotFound
	}
	return blockHeight, nil
}

func (m *defaultBlockModel) GetBlocksTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("get total block count error, err: %s", dbTx.Error.Error())
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetBlockStatusCacheByBlockHeight
	Params: blockHeight int64
	Return: blockStatus int64, err
	Description: get blockStatus cache by blockHeight
*/

//goland:noinspection GoNameStartsWithPackageName
type BlockStatusInfo struct {
	BlockStatus int64
	CommittedAt int64
	VerifiedAt  int64
}

func (m *defaultBlockModel) CreateBlockForCommitter(pendingMempoolTxs []*mempool.MempoolTx, blockStates *BlockStates) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		// update mempool
		for _, mempoolTx := range pendingMempoolTxs {
			dbTx := tx.Table(mempool.MempoolTableName).Where("id = ?", mempoolTx.ID).
				Select("*").
				Updates(&mempoolTx)
			if dbTx.Error != nil {
				logx.Errorf("unable to update mempool tx: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errors.New("no new mempoolTx")
			}
		}
		// create block
		if blockStates.Block != nil {
			dbTx := tx.Table(m.table).Create(blockStates.Block)
			if dbTx.Error != nil {
				logx.Errorf("unable to create block: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				blockInfo, err := json.Marshal(blockStates.Block)
				if err != nil {
					logx.Errorf("unable to marshal block, err: %s", err.Error())
					return err
				}
				logx.Errorf("invalid block info: %s", string(blockInfo))
				return errors.New("invalid block info")
			}
		}
		// create block for commit
		if blockStates.BlockForCommit != nil {
			dbTx := tx.Table(blockForCommit.BlockForCommitTableName).Create(blockStates.BlockForCommit)
			if dbTx.Error != nil {
				logx.Errorf("unable to create block for commit: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				commitInfo, err := json.Marshal(blockStates.BlockForCommit)
				if err != nil {
					logx.Errorf("unable to marshal block for commit, err=%s", err.Error())
					return err
				}
				logx.Errorf("invalid block for commit info: %s", string(commitInfo))
				return errors.New("invalid block for commit info")
			}
		}
		// create new account
		if len(blockStates.PendingNewAccount) != 0 {
			dbTx := tx.Table(account.AccountTableName).CreateInBatches(blockStates.PendingNewAccount, len(blockStates.PendingNewAccount))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockStates.PendingNewAccount)) {
				logx.Errorf("unable to create new account, rowsAffected=%d, rowsCreated=%d", dbTx.RowsAffected, len(blockStates.PendingNewAccount))
				return errors.New("unable to create new account")
			}
		}
		// update account
		for _, pendingAccount := range blockStates.PendingUpdateAccount {
			dbTx := tx.Table(account.AccountTableName).Where("id = ?", pendingAccount.ID).
				Select("*").
				Updates(&pendingAccount)
			if dbTx.Error != nil {
				logx.Errorf("unable to update account: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errors.New("no new account")
			}
		}
		// create new account history
		if len(blockStates.PendingNewAccountHistory) != 0 {
			dbTx := tx.Table(account.AccountHistoryTableName).CreateInBatches(blockStates.PendingNewAccountHistory, len(blockStates.PendingNewAccountHistory))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockStates.PendingNewAccountHistory)) {
				logx.Errorf("unable to create new account history, rowsAffected=%d, rowsCreated=%d", dbTx.RowsAffected, len(blockStates.PendingNewAccountHistory))
				return errors.New("unable to create new account history")
			}
		}
		// create new liquidity
		if len(blockStates.PendingNewLiquidity) != 0 {
			dbTx := tx.Table(liquidity.LiquidityTable).CreateInBatches(blockStates.PendingNewLiquidity, len(blockStates.PendingNewLiquidity))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockStates.PendingNewLiquidity)) {
				logx.Errorf("unable to create new liquidity, rowsAffected=%d, rowsCreated=%d", dbTx.RowsAffected, len(blockStates.PendingNewLiquidity))
				return errors.New("unable to create new liquidity")
			}
		}
		// update liquidity
		for _, pendingLiquidity := range blockStates.PendingUpdateLiquidity {
			dbTx := tx.Table(liquidity.LiquidityTable).Where("id = ?", pendingLiquidity.ID).
				Select("*").
				Updates(&pendingLiquidity)
			if dbTx.Error != nil {
				logx.Errorf("unable to update liquidity: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errors.New("no new liquidity")
			}
		}
		// create new liquidity history
		if len(blockStates.PendingNewLiquidityHistory) != 0 {
			dbTx := tx.Table(liquidity.LiquidityHistoryTable).CreateInBatches(blockStates.PendingNewLiquidityHistory, len(blockStates.PendingNewLiquidityHistory))
			if dbTx.Error != nil {
				logx.Errorf("create liquidity history error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockStates.PendingNewLiquidityHistory)) {
				logx.Errorf("unable to create new liquidity history, rowsAffected=%d, rowsToCreate=%d",
					dbTx.RowsAffected, len(blockStates.PendingNewLiquidityHistory))
				return errors.New("unable to create new liquidity history")
			}
		}
		// create new nft
		if len(blockStates.PendingNewNft) != 0 {
			dbTx := tx.Table(nft.L2NftTableName).CreateInBatches(blockStates.PendingNewNft, len(blockStates.PendingNewNft))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockStates.PendingNewNft)) {
				logx.Errorf("unable to create new nft, rowsAffected=%d, rowsCreated=%d", dbTx.RowsAffected, len(blockStates.PendingNewNft))
				return errors.New("unable to create new nft")
			}
		}
		// update nft
		for _, pendingNft := range blockStates.PendingUpdateNft {
			dbTx := tx.Table(nft.L2NftTableName).Where("id = ?", pendingNft.ID).
				Select("*").
				Updates(&pendingNft)
			if dbTx.Error != nil {
				logx.Errorf("unable to update nft: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errors.New("no new nft")
			}
		}
		// new nft history
		if len(blockStates.PendingNewNftHistory) != 0 {
			dbTx := tx.Table(nft.L2NftHistoryTableName).CreateInBatches(blockStates.PendingNewNftHistory, len(blockStates.PendingNewNftHistory))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockStates.PendingNewNftHistory)) {
				logx.Errorf("unable to create new nft history, rowsAffected=%d, rowsToCreate=%d",
					dbTx.RowsAffected, len(blockStates.PendingNewNftHistory))
				return errors.New("unable to create new nft history")
			}
		}
		// new nft withdraw history
		if len(blockStates.PendingNewNftWithdrawHistory) != 0 {
			dbTx := tx.Table(nft.L2NftWithdrawHistoryTableName).CreateInBatches(blockStates.PendingNewNftWithdrawHistory, len(blockStates.PendingNewNftWithdrawHistory))
			if dbTx.Error != nil {
				logx.Errorf("create nft withdraw history error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockStates.PendingNewNftWithdrawHistory)) {
				return errors.New("unable to create new nft withdraw")
			}
		}
		// new offer
		if len(blockStates.PendingNewOffer) != 0 {
			dbTx := tx.Table(nft.OfferTableName).CreateInBatches(blockStates.PendingNewOffer, len(blockStates.PendingNewOffer))
			if dbTx.Error != nil {
				logx.Errorf("create new offer error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockStates.PendingNewOffer)) {
				return errors.New("unable to create new offer")
			}
		}
		// new nft exchange
		if len(blockStates.PendingNewL2NftExchange) != 0 {
			dbTx := tx.Table(nft.L2NftExchangeTableName).CreateInBatches(blockStates.PendingNewL2NftExchange, len(blockStates.PendingNewL2NftExchange))
			if dbTx.Error != nil {
				logx.Errorf("create new nft exchange error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockStates.PendingNewL2NftExchange)) {
				return errors.New("unable to create new nft exchange")
			}
		}
		return nil
	})
}

func (m *defaultBlockModel) CreateNewBlock(oBlock *Block) (err error) {
	if oBlock == nil {
		return errors.New("nil block")
	}
	if oBlock.BlockStatus != StatusProposing {
		return errors.New("new block status isn't proposing")
	}

	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(oBlock)
		if dbTx.Error != nil {
			logx.Errorf("unable to create block: %s", dbTx.Error.Error())
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			blockInfo, err := json.Marshal(oBlock)
			if err != nil {
				logx.Errorf("unable to marshal block, err: %s", err.Error())
				return err
			}
			logx.Errorf("invalid block info: %s", string(blockInfo))
			return errors.New("invalid block info")
		}

		return nil
	})
}

func (m *defaultBlockModel) GetBlocksForProverBetween(start, end int64) (blocks []*Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ? AND block_height >= ? AND block_height <= ?", StatusCommitted, start, end).
		Order("block_height").
		Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("unable to get blocks, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
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
		logx.Errorf("unable to get block: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, errorcode.DbErrNotFound
	}
	return block.BlockHeight, nil
}
