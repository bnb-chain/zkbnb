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
	cacheZkbasVolumeIdPrefix   = "cache:zkbas:volume:id:"
	cacheZkbasVolumeDatePrefix = "cache:zkbas:volume:date:"
)

type (
	VolumeModel interface {
		CreateVolumeTable() error
		DropVolumeTable() error
		CreateVolume(volume *Volume) error
		CreateVolumesInBatch(volumes []*Volume) error
		CreateVolumesAndTVLsInBatch(volumes []*Volume, tvls []*TVL, volumesPool []*VolumePool, tvlsPool []*TVLPool) error
		GetLatestBlockHeight() (blockHeight int64, err error)
		GetVolumeSumBetweenDate(date1 time.Time, date2 time.Time) (result []*ResultVolumeSum, err error)
		GetVolumeSumGroupByDays() (result []*ResultVolumeDaySum, err error)
	}

	defaultVolumeModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	Volume struct {
		gorm.Model
		AssetId     int64 `gorm:"index"`
		VolumeDelta int64
		BlockHeight int64
		Date        time.Time `gorm:"index"` //days:hour
	}
)

func NewVolumeModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) VolumeModel {
	return &defaultVolumeModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      `volume`,
		DB:         db,
	}
}

func (*Volume) TableName() string {
	return `volume`
}

/*
	Func: CreateVolumeTable
	Params:
	Return: err error
	Description: create Volume table
*/
func (m *defaultVolumeModel) CreateVolumeTable() error {
	return m.DB.AutoMigrate(Volume{})
}

/*
	Func: DropVolumeTable
	Params:
	Return: err error
	Description: drop Volume table
*/
func (m *defaultVolumeModel) DropVolumeTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateVolume
	Params: volume *Volume
	Return: error
	Description: Insert New Volume
*/

func (m *defaultVolumeModel) CreateVolume(volume *Volume) error {
	dbTx := m.DB.Table(m.table).Create(volume)
	if dbTx.Error != nil {
		logx.Errorf("[volume.CreateVolume] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Errorf("[volume.CreateVolume] Create Volume Error")
		return errorcode.DbErrFailToCreateVolume
	}
	return nil
}

/*
	Func: CreateVolumesInBatch
	Params: volumes []*Volume
	Return: error
	Description: Insert New Volumes in Batch
*/

func (m *defaultVolumeModel) CreateVolumesInBatch(volumes []*Volume) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(volumes, len(volumes))
	if dbTx.Error != nil {
		logx.Errorf("[volume.CreateVolumesInBatch] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Errorf("[volume.CreateVolumesInBatch] Create Volume Error")
		return errorcode.DbErrFailToCreateVolume
	}
	return nil
}

/*
	Func: CreateVolumesAndTVLInBatch
	Params: volumes []*Volume
	Return: error
	Description: Insert New Volumes in Batch
*/

func (m *defaultVolumeModel) CreateVolumesAndTVLsInBatch(volumes []*Volume, tvls []*TVL, volumesPool []*VolumePool, tvlsPool []*TVLPool) error {
	var (
		tvlTableName   = "tvl"
		volumePoolName = "volume_pool"
		tvlPoolName    = "tvl_pool"
	)
	err := m.DB.Transaction(func(tx *gorm.DB) error {
		dbTx := tx.Table(m.table).CreateInBatches(volumes, len(volumes))
		if dbTx.Error != nil {
			logx.Errorf("[volume.CreateVolumesAndTVLInBatch] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[volume.CreateVolumesAndTVLInBatch] Create Volume Error")
			return errorcode.DbErrFailToCreateVolume
		}

		if len(tvls) != 0 {
			dbTx = tx.Table(tvlTableName).CreateInBatches(tvls, len(tvls))
			if dbTx.Error != nil {
				logx.Errorf("[tvl.CreateVolumesAndTVLInBatch] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[tvl.CreateVolumesAndTVLInBatch] Create TVL Error")
				return errorcode.DbErrFailToCreateTVL
			}
		}

		if len(volumesPool) != 0 {
			dbTx = tx.Table(volumePoolName).CreateInBatches(volumesPool, len(volumesPool))
			if dbTx.Error != nil {
				logx.Errorf("[tvl.CreateVolumesAndTVLInBatch] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[tvl.CreateVolumesAndTVLInBatch] Create TVL Error")
				return errorcode.DbErrFailToCreateTVL
			}
		}

		if len(tvlsPool) != 0 {
			dbTx = tx.Table(tvlPoolName).CreateInBatches(tvlsPool, len(tvlsPool))
			if dbTx.Error != nil {
				logx.Errorf("[tvl.CreateVolumesAndTVLInBatch] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[tvl.CreateVolumesAndTVLInBatch] Create TVL Error")
				return errorcode.DbErrFailToCreateTVL
			}
		}

		return nil
	})

	return err
}

/*
	Func: CreateVolume
	Params: volume *Volume
	Return: error
	Description: Insert New Volume
*/

func (m *defaultVolumeModel) GetLatestBlockHeight() (blockHeight int64, err error) {
	dbTx := m.DB.Table(m.table).Select("block_height").Order("block_height desc").Limit(1).Find(&blockHeight)
	if dbTx.Error != nil {
		logx.Errorf("[volume.CreateVolume] %s", dbTx.Error)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Info("[volume.CreateVolume] no result in volume table")
		return 0, errorcode.DbErrNotFound
	}
	return blockHeight, nil
}

/*
	Func: GetVolumeSum
	Params: tvls []*TVL
	Return: error
	Description: Insert New TVLs in Batch
*/
type ResultVolumeSum struct {
	AssetId int64
	Total   int64
}

func (m *defaultVolumeModel) GetVolumeSumBetweenDate(date1 time.Time, date2 time.Time) (result []*ResultVolumeSum, err error) {
	dbTx := m.DB.Table(m.table).Select("asset_id, sum(volume_delta) as total").Where("date <= ? and date > ?", date1, date2).Group("asset_id").Order("asset_id").Find(&result)
	if dbTx.Error != nil {
		logx.Errorf("[tvl.GetVolumeSum] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[volume.CreateVolume] no result in volume table")
		return nil, errorcode.DbErrNotFound
	}

	return result, nil
}

type ResultVolumeDaySum struct {
	Total   int64
	AssetId int64
	Date    time.Time
}

func (m *defaultVolumeModel) GetVolumeSumGroupByDays() (result []*ResultVolumeDaySum, err error) {
	// SELECT SUM( lock_amount_delta ), asset_id, date_trunc( 'day', DATE ) FROM tvl GROUP BY date_trunc( 'day', DATE ), asset_id ORDER BY date_trunc( 'day', DATE ), asset_id
	dbTx := m.DB.Table(m.table).Debug().Select("sum(volume_delta) as total, asset_id, date_trunc( 'day', DATE )::date as date").Group("date_trunc( 'day', DATE ), asset_id").Order("date_trunc( 'day', DATE ), asset_id").Find(&result)
	if dbTx.Error != nil {
		logx.Errorf("[tvl.GetVolumeSumGroupByDays] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[volume.GetVolumeSumGroupByDays] no result in tvl table")
		return nil, errorcode.DbErrNotFound
	}

	return result, nil
}
