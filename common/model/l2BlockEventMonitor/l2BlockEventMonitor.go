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

package l2BlockEventMonitor

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	L2BlockEventMonitorModel interface {
		CreateL2BlockEventMonitorTable() error
		DropL2BlockEventMonitorTable() error
	}

	defaultL2BlockEventMonitorModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2BlockEventMonitor struct {
		gorm.Model
		// event type, 1 - Committed, 2 - Verified, 3 - Reverted
		BlockEventType uint8 `gorm:"index"`
		// layer-1 block height
		L1BlockHeight int64
		// layer-1 txVerification hash
		L1TxHash string
		// layer-2 block height
		L2BlockHeight int64 `gorm:"index"`
		// status
		Status int
	}
)

func (*L2BlockEventMonitor) TableName() string {
	return TableName
}

func NewL2BlockEventMonitorModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2BlockEventMonitorModel {
	return &defaultL2BlockEventMonitorModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TableName,
		DB:         db,
	}
}

/*
	Func: CreateL2BlockEventMonitorTable
	Params:
	Return: err error
	Description: create l2 txVerification event monitor table
*/
func (m *defaultL2BlockEventMonitorModel) CreateL2BlockEventMonitorTable() error {
	return m.DB.AutoMigrate(L2BlockEventMonitor{})
}

/*
	Func: DropL2BlockEventMonitorTable
	Params:
	Return: err error
	Description: drop l2 txVerification event monitor table
*/
func (m *defaultL2BlockEventMonitorModel) DropL2BlockEventMonitorTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
