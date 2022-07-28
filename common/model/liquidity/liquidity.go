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

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	LiquidityModel interface {
		CreateLiquidityTable() error
		DropLiquidityTable() error
		CreateLiquidity(liquidity *Liquidity) error
		CreateLiquidityInBatches(entities []*Liquidity) error
		GetLiquidityByPairIndex(pairIndex int64) (entity *Liquidity, err error)
	}

	defaultLiquidityModel struct {
		sqlc.CachedConn
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

func NewLiquidityModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) LiquidityModel {
	return &defaultLiquidityModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      LiquidityTable,
		DB:         db,
	}
}

func (*Liquidity) TableName() string {
	return LiquidityTable
}

/*
	Func: CreateAccountLiquidityTable
	Params:
	Return: err error
	Description: create account liquidity table
*/
func (m *defaultLiquidityModel) CreateLiquidityTable() error {
	return m.DB.AutoMigrate(Liquidity{})
}

/*
	Func: DropAccountLiquidityTable
	Params:
	Return: err error
	Description: drop account liquidity table
*/
func (m *defaultLiquidityModel) DropLiquidityTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateAccountLiquidity
	Params: liquidity *Liquidity
	Return: err error
	Description: create account liquidity entity
*/
func (m *defaultLiquidityModel) CreateLiquidity(liquidity *Liquidity) error {
	dbTx := m.DB.Table(m.table).Create(liquidity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateLiquidity] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.CreateLiquidity] %s", errorcode.DbErrFailToCreateLiquidity)
		logx.Error(err)
		return errorcode.DbErrFailToCreateLiquidity
	}
	return nil
}

/*
	Func: CreateAccountLiquidityInBatches
	Params: entities []*Liquidity
	Return: err error
	Description: create account liquidity entities
*/
func (m *defaultLiquidityModel) CreateLiquidityInBatches(entities []*Liquidity) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(entities, len(entities))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateLiquidityInBatches] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.CreateLiquidityInBatches] %s", errorcode.DbErrFailToCreateLiquidity)
		logx.Error(err)
		return errorcode.DbErrFailToCreateLiquidity
	}
	return nil
}

/*
	Func: GetAccountLiquidityByPairIndex
	Params: pairIndex int64
	Return: entities []*Liquidity, err error
	Description: get account liquidity entities by account index
*/
func (m *defaultLiquidityModel) GetLiquidityByPairIndex(pairIndex int64) (entity *Liquidity, err error) {
	dbTx := m.DB.Table(m.table).Where("pair_index = ?", pairIndex).Find(&entity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetLiquidityByPairIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetLiquidityByPairIndex] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return entity, nil
}
