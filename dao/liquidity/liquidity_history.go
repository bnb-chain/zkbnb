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

package liquidity

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	LiquidityHistoryTable = `liquidity_history`
)

type (
	LiquidityHistoryModel interface {
		CreateLiquidityHistoryTable() error
		DropLiquidityHistoryTable() error
		GetLatestLiquidityByBlockHeight(blockHeight int64, limit int, offset int) (entities []*LiquidityHistory, err error)
		GetLatestLiquidityCountByBlockHeight(blockHeight int64) (count int64, err error)
		CreateLiquidityHistoriesInTransact(tx *gorm.DB, histories []*LiquidityHistory) error
		GetLiquidityByIndex(pairIndex, blockHeight int64) (*LiquidityHistory, error)
	}

	defaultLiquidityHistoryModel struct {
		table string
		DB    *gorm.DB
	}

	LiquidityHistory struct {
		gorm.Model
		PairIndex            int64
		AssetAId             int64
		AssetA               string
		AssetBId             int64
		AssetB               string
		LpAmount             string
		KLast                string
		FeeRate              int64
		TreasuryAccountIndex int64
		TreasuryRate         int64
		L2BlockHeight        int64
	}
)

func NewLiquidityHistoryModel(db *gorm.DB) LiquidityHistoryModel {
	return &defaultLiquidityHistoryModel{
		table: LiquidityHistoryTable,
		DB:    db,
	}
}

func (*LiquidityHistory) TableName() string {
	return LiquidityHistoryTable
}

func (m *defaultLiquidityHistoryModel) CreateLiquidityHistoryTable() error {
	return m.DB.AutoMigrate(LiquidityHistory{})
}

func (m *defaultLiquidityHistoryModel) DropLiquidityHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultLiquidityHistoryModel) GetLatestLiquidityByBlockHeight(blockHeight int64, limit int, offset int) (entities []*LiquidityHistory, err error) {
	subQuery := m.DB.Table(m.table).Select("*").
		Where("pair_index = a.pair_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height", blockHeight)

	dbTx := m.DB.Table(m.table+" as a").Select("*").
		Where("NOT EXISTS (?) AND l2_block_height <= ?", subQuery, blockHeight).
		Limit(limit).Offset(offset).
		Order("pair_index")

	if dbTx.Find(&entities).Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return entities, nil
}

func (m *defaultLiquidityHistoryModel) GetLatestLiquidityCountByBlockHeight(blockHeight int64) (count int64, err error) {
	subQuery := m.DB.Table(m.table).Select("*").
		Where("pair_index = a.pair_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height", blockHeight)

	dbTx := m.DB.Table(m.table+" as a").
		Where("NOT EXISTS (?) AND l2_block_height <= ?", subQuery, blockHeight)

	if dbTx.Count(&count).Error != nil {
		return 0, dbTx.Error
	}
	return count, nil
}

func (m *defaultLiquidityHistoryModel) CreateLiquidityHistoriesInTransact(tx *gorm.DB, histories []*LiquidityHistory) error {
	dbTx := tx.Table(m.table).CreateInBatches(histories, len(histories))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(histories)) {
		return types.DbErrFailToCreateLiquidityHistory
	}
	return nil
}

func (m *defaultLiquidityHistoryModel) GetLiquidityByIndex(pairIndex, blockHeight int64) (*LiquidityHistory, error) {
	var history LiquidityHistory
	dbTx := m.DB.Table(m.table).Where("pair_index = ? AND l2_block_height <= ?", pairIndex, blockHeight).
		Order("l2_block_height desc").Limit(1).Find(&history)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return &history, nil
}
