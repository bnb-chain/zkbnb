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

package info

import (
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

var (
	cacheZkbasTVLPoolPoolIdPrefix = "cache:zkbas:tvlpool:id:"
	cacheZkbasTVLPoolDatePrefix   = "cache:zkbas:tvlpool:date:"
)

type (
	TVLPoolModel interface {
		CreateTVLPoolTable() error
		DropTVLPoolTable() error
		CreateTVLPool(tvlpool *TVLPool) error
		CreateTVLPoolsInBatch(tvlPools []*TVLPool) error
		GetPoolAmountSum(date time.Time) (result []*ResultTvlPoolSum, err error)
	}

	defaultTVLPoolModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	TVLPool struct {
		gorm.Model
		PoolId       int64 `gorm:"index"`
		AmountDeltaA int64
		AmountDeltaB int64
		Date         time.Time `gorm:"index"` //days:hour
	}
)

func NewTVLPoolModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) TVLPoolModel {
	return &defaultTVLPoolModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      `tvl_pool`,
		DB:         db,
	}
}

func (*TVLPool) TableName() string {
	return `tvl_pool`
}

/*
	Func: CreateTVLPoolTable
	Params:
	Return: err error
	Description: create TVLPool table
*/
func (m *defaultTVLPoolModel) CreateTVLPoolTable() error {
	return m.DB.AutoMigrate(TVLPool{})
}

/*
	Func: DropTVLPoolTable
	Params:
	Return: err error
	Description: drop TVLPool table
*/
func (m *defaultTVLPoolModel) DropTVLPoolTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateTVLPool
	Params: tvlpool *TVLPool
	Return: error
	Description: Insert New TVLPool
*/

func (m *defaultTVLPoolModel) CreateTVLPool(tvlpool *TVLPool) error {
	dbTx := m.DB.Table(m.table).Create(tvlpool)
	if dbTx.Error != nil {
		logx.Errorf("[tvlpool.CreateTVLPool] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Errorf("[tvlpool.CreateTVLPool] Delete Invalid Mempool Tx")
		return errorcode.DbErrFailToCreateTVL
	}
	return nil
}

/*
	Func: CreateTVLPoolsInBatch
	Params: tvlPools []*TVLPool
	Return: error
	Description: Insert New TVLPools in Batch
*/
func (m *defaultTVLPoolModel) CreateTVLPoolsInBatch(tvlPools []*TVLPool) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(tvlPools, len(tvlPools))
	if dbTx.Error != nil {
		logx.Errorf("[tvlpool.CreateTVLPoolsInBatch] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Errorf("[tvlpool.CreateTVLPoolsInBatch] Create TVLPool Error")
		return errorcode.DbErrFailToCreateTVL
	}
	return nil
}

/*
	Func: GetLockAmountSum
	Params: tvls []*TVL
	Return: error
	Description: Insert New TVLs in Batch
*/
type ResultTvlPoolSum struct {
	PoolId int64
	TotalA int64
	TotalB int64
}

func (m *defaultTVLPoolModel) GetPoolAmountSum(date time.Time) (result []*ResultTvlPoolSum, err error) {
	dbTx := m.DB.Table(m.table).Select("pool_id, sum(amount_delta_a) as total_a, sum(amount_delta_b) as total_b").Where("date <= ?", date).Group("pool_id").Order("pool_id").Find(&result)
	if dbTx.Error != nil {
		logx.Errorf("[tvl.GetPoolAmountSum] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[volume.GetPoolAmountSum] no result in tvl pool table")
		return nil, errorcode.DbErrNotFound
	}

	return result, nil
}
