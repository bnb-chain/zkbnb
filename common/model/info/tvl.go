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

package info

import (
	"time"

	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
)

type (
	TVLModel interface {
		CreateTVLTable() error
		DropTVLTable() error
		CreateTVL(tvl *TVL) error
		CreateTVLsInBatch(tvls []*TVL) error
		GetLockAmountSum(date time.Time) (result []*ResultTvlSum, err error)
		GetLockAmountSumGroupByDays() (result []*ResultTvlDaySum, err error)
	}

	defaultTVLModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	TVL struct {
		gorm.Model
		AssetId         int64 `gorm:"index"`
		LockAmountDelta int64
		BlockHeight     int64
		Date            time.Time `gorm:"index"` //days:hour
	}
)

func (*TVL) TableName() string {
	return `tvl`
}

/*
	Func: CreateTVLTable
	Params:
	Return: err error
	Description: create TVL table
*/
func (m *defaultTVLModel) CreateTVLTable() error {
	return m.DB.AutoMigrate(TVL{})
}

/*
	Func: DropTVLTable
	Params:
	Return: err error
	Description: drop TVL table
*/
func (m *defaultTVLModel) DropTVLTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateTVL
	Params: tvl *TVL
	Return: error
	Description: Insert New TVL
*/

func (m *defaultTVLModel) CreateTVL(tvl *TVL) error {
	dbTx := m.DB.Table(m.table).Create(tvl)
	if dbTx.Error != nil {
		logx.Errorf("[tvl.CreateTVL] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Errorf("[tvl.CreateTVL] Delete Invalid Mempool Tx")
		return ErrInvalidTVL
	}
	return nil
}

/*
	Func: CreateTVLsInBatch
	Params: tvls []*TVL
	Return: error
	Description: Insert New TVLs in Batch
*/
func (m *defaultTVLModel) CreateTVLsInBatch(tvls []*TVL) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(tvls, len(tvls))
	if dbTx.Error != nil {
		logx.Errorf("[tvl.CreateTVLsInBatch] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Errorf("[tvl.CreateTVLsInBatch] Create TVL Error")
		return ErrInvalidTVL
	}
	return nil
}

/*
	Func: GetLockAmountSum
	Params: tvls []*TVL
	Return: error
	Description: Insert New TVLs in Batch
*/
type ResultTvlSum struct {
	AssetId int64
	Total   int64
}

func (m *defaultTVLModel) GetLockAmountSum(date time.Time) (result []*ResultTvlSum, err error) {
	dbTx := m.DB.Table(m.table).Select("asset_id, sum(lock_amount_delta) as total").Where("date <= ?", date).Group("asset_id").Order("asset_id").Find(&result)
	if dbTx.Error != nil {
		logx.Errorf("[tvl.CreateTVLsInBatch] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[volume.GetLockAmountSum] no result in tvl table")
		return nil, ErrNotFound
	}

	return result, nil
}

type ResultTvlDaySum struct {
	Total   int64
	AssetId int64
	Date    time.Time
}

func (m *defaultTVLModel) GetLockAmountSumGroupByDays() (result []*ResultTvlDaySum, err error) {
	// SELECT SUM( lock_amount_delta ), asset_id, date_trunc( 'day', DATE ) FROM tvl GROUP BY date_trunc( 'day', DATE ), asset_id ORDER BY date_trunc( 'day', DATE ), asset_id
	dbTx := m.DB.Table(m.table).Debug().Select("sum(lock_amount_delta) as total, asset_id, date_trunc( 'day', DATE )::date as date").Group("date_trunc( 'day', DATE ), asset_id").Order("date_trunc( 'day', DATE ), asset_id").Find(&result)
	if dbTx.Error != nil {
		logx.Errorf("[tvl.GetLockAmountSumGroupByDays] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[volume.GetLockAmountSumGroupByDays] no result in tvl table")
		return nil, ErrNotFound
	}

	return result, nil
}
