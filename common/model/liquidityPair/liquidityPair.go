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

package liquidityPair

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheTradingPairIdPrefix        = "cache::tradingPair:id:"
	cacheTradingPairPairIndexPrefix = "cache::tradingPair:pairIndex:"
)

type (
	LiquidityPairModel interface {
		CreateLiquidityPairTable() error
		DropLiquidityPairTable() error
		CreateLiquidityPair(liquidityPair *LiquidityPair) error
		CreateLiquidityPairsInBatches(liquidityPairs []*LiquidityPair) (rowsAffected int64, err error)
		GetLiquidityPairByIndex(pairIndex int64) (tradingPair *LiquidityPair, err error)
		GetAllLiquidityPairs() (tradingPairs []*LiquidityPair, err error)
		GetPairIndexCount() (res int64, err error)
	}

	defaultLiquidityPairModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	LiquidityPair struct {
		gorm.Model
		PairIndex    int64 `gorm:"uniqueIndex"`
		AssetAId     int64
		AssetAName   string
		AssetBId     int64
		AssetBName   string
		FeeRate      int64
		TreasuryRate int64
		Status       int
	}
)

func NewLiquidityPairModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) LiquidityPairModel {
	return &defaultLiquidityPairModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      `liquidity_pair`,
		DB:         db,
	}
}

func (*LiquidityPair) TableName() string {
	return `liquidity_pair`
}

/*
	Func: CreateLiquidityPairTable
	Params:
	Return: err error
	Description: create liquidity pair table
*/
func (m *defaultLiquidityPairModel) CreateLiquidityPairTable() error {
	return m.DB.AutoMigrate(LiquidityPair{})
}

/*
	Func: DropLiquidityPairTable
	Params:
	Return: err error
	Description: drop liquidity pair table
*/
func (m *defaultLiquidityPairModel) DropLiquidityPairTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
 */
func (m *defaultLiquidityPairModel) CreateLiquidityPair(liquidityPair *LiquidityPair) error {
	dbTx := m.DB.Table(m.table).Create(liquidityPair)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateLiquidityPair] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.CreateLiquidityPair] %s", ErrInvalidLiquidityPairInput)
		logx.Error(err)
		return ErrInvalidLiquidityPairInput
	}
	return nil
}

/*
	Func: CreateLiquidityPairsInBatches
	Params: []*LiquidityPair
	Return: rowsAffected int64, err error
	Description: create LiquidityPair batches
*/

func (m *defaultLiquidityPairModel) CreateLiquidityPairsInBatches(liquidityPairs []*LiquidityPair) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(liquidityPairs, len(liquidityPairs))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateLiquidityPairsInBatches] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(liquidityPairs)) {
		err := fmt.Sprintf("[liquidity.CreateLiquidityPairsInBatches] %s", ErrInvalidLiquidityPairInput)
		logx.Error(err)
		return 0, ErrInvalidLiquidityPairInput
	}
	return dbTx.RowsAffected, nil
}

/*
	Func: GetLiquidityPairByIndex
	Params: pairIndex uint32
	Return: tradingPair *LiquidityPair, err error
	Description:  get liquidity pair by pair index
*/
func (m *defaultLiquidityPairModel) GetLiquidityPairByIndex(pairIndex int64) (tradingPair *LiquidityPair, err error) {
	dbTx := m.DB.Table(m.table).Where("pair_index = ?", pairIndex).Find(&tradingPair)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetLiquidityPairByIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetLiquidityPairByIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return tradingPair, nil
}

/*
	Func: GetAllLiquidityPairs
	Params:
	Return: tradingPair []*LiquidityPair, err error
	Description:  get all liquidity pairs
*/
func (m *defaultLiquidityPairModel) GetAllLiquidityPairs() (tradingPairs []*LiquidityPair, err error) {
	dbTx := m.DB.Table(m.table).Find(&tradingPairs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetAllLiquidityPairs] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetAllLiquidityPairs] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return tradingPairs, nil
}

/*
	Func: GetAvailableLiquidityPairsByAssetId
	Params: assetId uint3
	Return: tradingPairs []*LiquidityPair, err error
	Description:  get available liquidity pair by asset id
*/
func (m *defaultLiquidityPairModel) GetAvailableLiquidityPairsByAssetId(assetId uint32) (tradingPairs []*LiquidityPair, err error) {
	dbTx := m.DB.Table(m.table).Where("left_asset_id = ? OR right_asset_id = ?", assetId, assetId).Find(&tradingPairs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetAvailableLiquidityPairsByAssetId] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetAvailableLiquidityPairsByAssetId] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return tradingPairs, nil
}

/*
	Func: GetPairIndexCount
	Params:
	Return: res int64, err error
	Description: get l2 asset id count
*/
func (m *defaultLiquidityPairModel) GetPairIndexCount() (res int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&res)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[liquidity.GetPairIndexCount] %s", dbTx.Error)
		logx.Error(errInfo)
		// TODO : to be modified
		return 0, dbTx.Error
	} else {
		return res, nil
	}
}
