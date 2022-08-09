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
		GetLiquidityByPairIndex(pairIndex int64) (entity *Liquidity, err error)
		GetAllLiquidityAssets() (liquidityList []*Liquidity, err error)
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
	Func: GetAccountLiquidityByPairIndex
	Params: pairIndex int64
	Return: entities []*Liquidity, err error
	Description: get account liquidity entities by account index
*/
func (m *defaultLiquidityModel) GetLiquidityByPairIndex(pairIndex int64) (entity *Liquidity, err error) {
	dbTx := m.DB.Table(m.table).Where("pair_index = ?", pairIndex).Find(&entity)
	if dbTx.Error != nil {
		logx.Errorf("get liquidity by pair index error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return entity, nil
}

func (m *defaultLiquidityModel) GetAllLiquidityAssets() (liquidityList []*Liquidity, err error) {
	dbTx := m.DB.Table(m.table).Order("id").Find(&liquidityList)
	if dbTx.Error != nil {
		return liquidityList, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return liquidityList, nil
}
