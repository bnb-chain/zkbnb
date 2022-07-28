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

package liquidity

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	LiquidityHistoryModel interface {
		CreateLiquidityHistoryTable() error
		DropLiquidityHistoryTable() error
		CreateLiquidityHistory(liquidity *LiquidityHistory) error
		CreateLiquidityHistoryInBatches(entities []*LiquidityHistory) error
		GetAccountLiquidityHistoryByPairIndex(pairIndex int64) (entities []*LiquidityHistory, err error)
		GetLatestLiquidityByBlockHeight(blockHeight int64) (entities []*LiquidityHistory, err error)
		GetLatestLiquidityByPairIndex(pairIndex int64) (entity *LiquidityHistory, err error)
	}

	defaultLiquidityHistoryModel struct {
		sqlc.CachedConn
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

func NewLiquidityHistoryModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) LiquidityHistoryModel {
	return &defaultLiquidityHistoryModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      LiquidityHistoryTable,
		DB:         db,
	}
}

func (*LiquidityHistory) TableName() string {
	return LiquidityHistoryTable
}

/*
	Func: CreateAccountLiquidityHistoryTable
	Params:
	Return: err error
	Description: create account liquidity table
*/
func (m *defaultLiquidityHistoryModel) CreateLiquidityHistoryTable() error {
	return m.DB.AutoMigrate(LiquidityHistory{})
}

/*
	Func: DropAccountLiquidityHistoryTable
	Params:
	Return: err error
	Description: drop account liquidity table
*/
func (m *defaultLiquidityHistoryModel) DropLiquidityHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateAccountLiquidityHistory
	Params: liquidity *LiquidityHistory
	Return: err error
	Description: create account liquidity entity
*/
func (m *defaultLiquidityHistoryModel) CreateLiquidityHistory(liquidity *LiquidityHistory) error {
	dbTx := m.DB.Table(m.table).Create(liquidity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateLiquidityHistory] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.CreateLiquidityHistory] %s", ErrInvalidAccountLiquidityInput)
		logx.Error(err)
		return ErrInvalidAccountLiquidityInput
	}
	return nil
}

/*
	Func: CreateAccountLiquidityHistoryInBatches
	Params: entities []*LiquidityHistory
	Return: err error
	Description: create account liquidity entities
*/
func (m *defaultLiquidityHistoryModel) CreateLiquidityHistoryInBatches(entities []*LiquidityHistory) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(entities, len(entities))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateLiquidityHistoryInBatches] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.CreateLiquidityHistoryInBatches] %s", ErrInvalidAccountLiquidityInput)
		logx.Error(err)
		return ErrInvalidAccountLiquidityInput
	}
	return nil
}

/*
	Func: GetAccountLiquidityHistoryByPairIndex
	Params: pairIndex int64
	Return: entities []*LiquidityHistory, err error
	Description: get account liquidity entities by account index
*/
func (m *defaultLiquidityHistoryModel) GetAccountLiquidityHistoryByPairIndex(pairIndex int64) (entities []*LiquidityHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("pair_index = ?", pairIndex).Find(&entities)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetAccountLiquidityHistoryByPairIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetAccountLiquidityHistoryByPairIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return entities, nil
}

func (m *defaultLiquidityHistoryModel) GetLatestLiquidityByBlockHeight(blockHeight int64) (entities []*LiquidityHistory, err error) {
	dbTx := m.DB.Table(m.table).
		Raw("SELECT a.* FROM liquidity_history a WHERE NOT EXISTS"+
			"(SELECT * FROM liquidity_history WHERE pair_index = a.pair_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height) "+
			"AND l2_block_height <= ? ORDER BY pair_index", blockHeight, blockHeight).
		Find(&entities)
	if dbTx.Error != nil {
		logx.Errorf("[GetValidAccounts] unable to get related accounts: %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return entities, nil
}

func (m *defaultLiquidityHistoryModel) GetLatestLiquidityByPairIndex(pairIndex int64) (entity *LiquidityHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("pair_index = ?", pairIndex).Order("l2_block_height desc").Find(&entity)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestLiquidityByPairIndex] unable to get related liquidity: %s", dbTx.Error.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return entity, nil
}
