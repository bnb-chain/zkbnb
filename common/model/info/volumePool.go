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
	cacheZkbasVolumePoolIdPrefix   = "cache:zkbas:volume:id:"
	cacheZkbasVolumePoolDatePrefix = "cache:zkbas:volume:date:"
)

type (
	VolumePoolModel interface {
		CreateVolumePoolTable() error
		DropVolumePoolTable() error
		CreateVolumePool(volume *VolumePool) error
		GetPoolVolumeSumBetweenDate(date1 time.Time, date2 time.Time) (result []*ResultVolumePoolSum, err error)
	}

	defaultVolumePoolModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	VolumePool struct {
		gorm.Model
		PoolId       int64 `gorm:"index"`
		VolumeDeltaA int64
		VolumeDeltaB int64
		Date         time.Time `gorm:"index"` //days:hour
	}
)

func NewVolumePoolModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) VolumePoolModel {
	return &defaultVolumePoolModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      `volume_pool`,
		DB:         db,
	}
}

func (*VolumePool) TableName() string {
	return `volume_pool`
}

/*
	Func: CreateVolumePoolTable
	Params:
	Return: err error
	Description: create VolumePool table
*/
func (m *defaultVolumePoolModel) CreateVolumePoolTable() error {
	return m.DB.AutoMigrate(VolumePool{})
}

/*
	Func: DropVolumePoolTable
	Params:
	Return: err error
	Description: drop VolumePool table
*/
func (m *defaultVolumePoolModel) DropVolumePoolTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateVolumePool
	Params: volumePool *VolumePool
	Return: error
	Description: Insert New VolumePool
*/

func (m *defaultVolumePoolModel) CreateVolumePool(volumePool *VolumePool) error {
	dbTx := m.DB.Table(m.table).Create(volumePool)
	if dbTx.Error != nil {
		logx.Errorf("[volumePool.CreateVolumePool] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Errorf("[volumePool.CreateVolumePool] Create VolumePool Error")
		return errorcode.DbErrFailToCreateVolume
	}
	return nil
}

/*
	Func: GetLockVolumeSum
	Params: tvls []*TVL
	Return: error
	Description: Insert New TVLs in Batch
*/
type ResultVolumePoolSum struct {
	PoolId int64
	TotalA int64
	TotalB int64
}

func (m *defaultVolumePoolModel) GetPoolVolumeSumBetweenDate(date1 time.Time, date2 time.Time) (result []*ResultVolumePoolSum, err error) {
	dbTx := m.DB.Table(m.table).Select("pool_id, sum(volume_delta_a) as total_a, sum(volume_delta_b) as total_b").Where("date <= ? and date > ?", date1, date2).Group("pool_id").Order("pool_id").Find(&result)
	if dbTx.Error != nil {
		logx.Errorf("[tvl.GetPoolVolumeSum] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[volume.GetPoolVolumeSum] no result in tvl pool table")
		return nil, errorcode.DbErrNotFound
	}

	return result, nil
}
