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
	"github.com/bnb-chain/zkbnb-crypto/legend/circuit/bn254/std"
	"gorm.io/gorm"
	"math/big"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	LiquidityTable = `liquidity`
)

type (
	LiquidityModel interface {
		CreateLiquidityTable() error
		DropLiquidityTable() error
		GetLiquidityByIndex(index int64) (entity *Liquidity, err error)
		GetAllLiquidity() (liquidityList []*Liquidity, err error)
		CreateLiquidityInTransact(tx *gorm.DB, liquidity []*Liquidity) error
		UpdateLiquidityInTransact(tx *gorm.DB, liquidity []*Liquidity) error
	}

	defaultLiquidityModel struct {
		table string
		DB    *gorm.DB
	}

	Liquidity struct {
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
	}
)

func NewLiquidityModel(db *gorm.DB) LiquidityModel {
	return &defaultLiquidityModel{
		table: LiquidityTable,
		DB:    db,
	}
}

func (*Liquidity) TableName() string {
	return LiquidityTable
}

func (m *defaultLiquidityModel) CreateLiquidityTable() error {
	return m.DB.AutoMigrate(Liquidity{})
}

func (m *defaultLiquidityModel) DropLiquidityTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultLiquidityModel) GetLiquidityByIndex(pairIndex int64) (entity *Liquidity, err error) {
	dbTx := m.DB.Table(m.table).Where("pair_index = ?", pairIndex).Find(&entity)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return entity, nil
}

func (m *defaultLiquidityModel) GetAllLiquidity() (liquidityList []*Liquidity, err error) {
	dbTx := m.DB.Table(m.table).Order("id").Find(&liquidityList)
	if dbTx.Error != nil {
		return liquidityList, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return liquidityList, nil
}

func (m *defaultLiquidityModel) CreateLiquidityInTransact(tx *gorm.DB, liquidity []*Liquidity) error {
	dbTx := tx.Table(m.table).CreateInBatches(liquidity, len(liquidity))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(liquidity)) {
		return types.DbErrFailToCreateLiquidity
	}
	return nil
}

func (m *defaultLiquidityModel) UpdateLiquidityInTransact(tx *gorm.DB, liquidity []*Liquidity) error {
	for _, pendingLiquidity := range liquidity {
		dbTx := tx.Table(m.table).Where("pair_index = ?", pendingLiquidity.PairIndex).
			Select("*").
			Updates(&pendingLiquidity)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToUpdateLiquidity
		}
	}
	return nil
}

func (n *Liquidity) ToStdLiquidity() *std.Liquidity {
	kLast, _ := new(big.Int).SetString(n.KLast, 10)
	assetA, _ := new(big.Int).SetString(n.AssetA, 10)
	assetB, _ := new(big.Int).SetString(n.AssetB, 10)
	lpAmount, _ := new(big.Int).SetString(n.LpAmount, 10)

	return &std.Liquidity{
		PairIndex:            n.PairIndex,
		AssetAId:             n.AssetAId,
		AssetA:               assetA,
		AssetBId:             n.AssetBId,
		AssetB:               assetB,
		LpAmount:             lpAmount,
		KLast:                kLast,
		FeeRate:              n.FeeRate,
		TreasuryAccountIndex: n.TreasuryAccountIndex,
		TreasuryRate:         n.TreasuryRate,
	}
}
