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

package l1syncedblock

import (
	"errors"

	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/dao/asset"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/priorityrequest"
	"github.com/bnb-chain/zkbas/dao/sysconfig"
	"github.com/bnb-chain/zkbas/types"
)

const (
	TableName = "l1_synced_block"

	TypeGeneric    int = 0
	TypeGovernance int = 1
)

type (
	L1SyncedBlockModel interface {
		CreateL1SyncedBlockTable() error
		DropL1SyncedBlockTable() error
		CreateGenericBlock(
			block *L1SyncedBlock,
			priorityRequests []*priorityrequest.PriorityRequest,
			pendingUpdateBlocks []*block.Block,
			pendingUpdateMempoolTxs []*mempool.MempoolTx,
		) (err error)

		CreateGovernanceBlock(
			block *L1SyncedBlock,
			l2Assets []*asset.Asset,
			pendingUpdateL2Assets []*asset.Asset,
			pendingNewSysConfigs []*sysconfig.SysConfig,
			pendingUpdateSysConfigs []*sysconfig.SysConfig,
		) (err error)
		GetLatestL1BlockByType(blockType int) (blockInfo *L1SyncedBlock, err error)
	}

	defaultL1EventModel struct {
		table string
		DB    *gorm.DB
	}

	L1SyncedBlock struct {
		gorm.Model
		// l1 block height
		L1BlockHeight int64
		// block info, array of hashes
		BlockInfo string
		Type      int
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

func (m *defaultL1EventModel) CreateGenericBlock(
	blockInfo *L1SyncedBlock,
	priorityRequests []*priorityrequest.PriorityRequest,
	pendingUpdateBlocks []*block.Block,
	pendingUpdateMempoolTxs []*mempool.MempoolTx,
) (err error) {
	const (
		Txs = "Txs"
	)

	err = m.DB.Transaction(
		func(tx *gorm.DB) error { // transact
			dbTx := tx.Table(m.table).Create(blockInfo)
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errors.New("unable to create l1 block info")
			}

			dbTx = tx.Table(priorityrequest.TableName).CreateInBatches(priorityRequests, len(priorityRequests))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(priorityRequests)) {
				return errors.New("unable to create priority requests")
			}

			// update blocks
			for _, pendingUpdateBlock := range pendingUpdateBlocks {
				dbTx := tx.Table(block.BlockTableName).Where("id = ?", pendingUpdateBlock.ID).
					Omit(Txs).
					Select("*").
					Updates(&pendingUpdateBlock)
				if dbTx.Error != nil {
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					if err != nil {
						return err
					}
					return errors.New("invalid block")
				}
			}

			// delete mempool txs
			for _, pendingDeleteMempoolTx := range pendingUpdateMempoolTxs {
				dbTx := tx.Table(mempool.MempoolTableName).Where("id = ?", pendingDeleteMempoolTx.ID).Delete(&pendingDeleteMempoolTx)
				if dbTx.Error != nil {
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					return errors.New("delete invalid mempool tx")
				}
			}
			return nil
		},
	)
	return err
}

func (m *defaultL1EventModel) CreateGovernanceBlock(
	block *L1SyncedBlock,
	pendingNewL2Assets []*asset.Asset,
	pendingUpdateL2Assets []*asset.Asset,
	pendingNewSysConfigs []*sysconfig.SysConfig,
	pendingUpdateSysConfigs []*sysconfig.SysConfig,
) (err error) {
	err = m.DB.Transaction(
		func(tx *gorm.DB) error {
			// create data for l1 block info
			dbTx := tx.Table(m.table).Create(block)
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errors.New("unable to create l1 block info")
			}
			// create l2 asset info
			if len(pendingNewL2Assets) != 0 {
				dbTx = tx.Table(asset.AssetTableName).CreateInBatches(pendingNewL2Assets, len(pendingNewL2Assets))
				if dbTx.Error != nil {
					return dbTx.Error
				}
				if dbTx.RowsAffected != int64(len(pendingNewL2Assets)) {
					return errors.New("invalid l2 asset info")
				}
			}
			// update l2 asset info
			for _, pendingUpdateL2AssetInfo := range pendingUpdateL2Assets {
				dbTx = tx.Table(asset.AssetTableName).Where("id = ?", pendingUpdateL2AssetInfo.ID).Select("*").Updates(&pendingUpdateL2AssetInfo)
				if dbTx.Error != nil {
					return dbTx.Error
				}
			}
			// create new sys config
			if len(pendingNewSysConfigs) != 0 {
				dbTx = tx.Table(sysconfig.TableName).CreateInBatches(pendingNewSysConfigs, len(pendingNewSysConfigs))
				if dbTx.Error != nil {
					return dbTx.Error
				}
				if dbTx.RowsAffected != int64(len(pendingNewSysConfigs)) {
					return errors.New("invalid sys config info")
				}
			}
			// update sys config
			for _, pendingUpdateSysConfig := range pendingUpdateSysConfigs {
				dbTx = tx.Table(sysconfig.TableName).Where("id = ?", pendingUpdateSysConfig.ID).Select("*").Updates(&pendingUpdateSysConfig)
				if dbTx.Error != nil {
					return dbTx.Error
				}
			}
			return nil
		},
	)
	return err
}

func (m *defaultL1EventModel) GetLatestL1BlockByType(blockType int) (blockInfo *L1SyncedBlock, err error) {
	dbTx := m.DB.Table(m.table).Where("type = ?", blockType).Order("l1_block_height desc").Find(&blockInfo)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return blockInfo, nil
}
