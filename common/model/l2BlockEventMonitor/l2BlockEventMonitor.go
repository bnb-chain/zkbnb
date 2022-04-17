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

package l2BlockEventMonitor

import (
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	L2BlockEventMonitorModel interface {
		CreateL2BlockEventMonitorTable() error
		DropL2BlockEventMonitorTable() error
		CreateL2BlockEventMonitor(tx *L2BlockEventMonitor) (bool, error)
		CreateL2BlockEventMonitorsInBatches(l2TxEventMonitors []*L2BlockEventMonitor) (rowsAffected int64, err error)
		GetL2BlockEventMonitors() (events []*L2BlockEventMonitor, err error)
		GetL2BlockEventMonitorsByEventType(blockEventType uint8) (events []*L2BlockEventMonitor, err error)
		GetL2BlockEventMonitorsByEventTypeAndStatus(eventType uint8, status int) (events []*L2BlockEventMonitor, err error)
		GetL2BlockEventMonitorsByStatus(status int) (rowsAffected int64, events []*L2BlockEventMonitor, err error)
		GetL2BlockEventMonitorsByTxType(txType uint8) (events []*L2BlockEventMonitor, err error)
		GetL2BlockEventMonitorsByL2BlockHeight(l2BlockHeight int64) (events []*L2BlockEventMonitor, err error)
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
		// layer-1 tx hash
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
	Description: create l2 tx event monitor table
*/
func (m *defaultL2BlockEventMonitorModel) CreateL2BlockEventMonitorTable() error {
	return m.DB.AutoMigrate(L2BlockEventMonitor{})
}

/*
	Func: DropL2BlockEventMonitorTable
	Params:
	Return: err error
	Description: drop l2 tx event monitor table
*/
func (m *defaultL2BlockEventMonitorModel) DropL2BlockEventMonitorTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateL2BlockEventMonitor
	Params: asset *L2BlockEventMonitor
	Return: bool, error
	Description: create L2BlockEventMonitor tx
*/
func (m *defaultL2BlockEventMonitorModel) CreateL2BlockEventMonitor(tx *L2BlockEventMonitor) (bool, error) {
	dbTx := m.DB.Table(m.table).Create(tx)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.CreateL2BlockEventMonitor] %s", dbTx.Error)
		logx.Error(err)
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		ErrInvalidL2BlockEventMonitor := errors.New("invalid l2BlockEventMonitor.go")
		err := fmt.Sprintf("[l2BlockEventMonitor.go.CreateL2BlockEventMonitor] %s", ErrInvalidL2BlockEventMonitor)
		logx.Error(err)
		return false, ErrInvalidL2BlockEventMonitor
	}
	return true, nil
}

/*
	Func: CreateL2BlockEventMonitorsInBatches
	Params: []*L2BlockEventMonitor
	Return: rowsAffected int64, err error
	Description: create L2BlockEventMonitor batches
*/
func (m *defaultL2BlockEventMonitorModel) CreateL2BlockEventMonitorsInBatches(l2TxEventMonitors []*L2BlockEventMonitor) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(l2TxEventMonitors, len(l2TxEventMonitors))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.CreateL1AssetsMonitorInBatches] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

/*
	GetL2BlockEventMonitors: get all L2BlockEventMonitors
*/
func (m *defaultL2BlockEventMonitorModel) GetL2BlockEventMonitors() (events []*L2BlockEventMonitor, err error) {
	dbTx := m.DB.Table(m.table).Find(&events).Order("l2_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitors] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitors] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return events, dbTx.Error
}

/*
	Func: GetL2BlockEventMonitorsByEventType
	Return: events []*L2BlockEventMonitor, err error
	Description: get l2TxEventMonitors by event type
*/
func (m *defaultL2BlockEventMonitorModel) GetL2BlockEventMonitorsByEventType(blockEventType uint8) (events []*L2BlockEventMonitor, err error) {
	dbTx := m.DB.Table(m.table).Where("block_event_type = ?", blockEventType).Find(&events).Order("l2_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitorsByEventType] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitorsByEventType] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return events, nil
}

/*
	Func: GetL2BlockEventMonitorsByEventTypeAndStatus
	Return: events []*L2BlockEventMonitor, err error
	Description: get l2TxEventMonitors by event type and status
*/
func (m *defaultL2BlockEventMonitorModel) GetL2BlockEventMonitorsByEventTypeAndStatus(eventType uint8, status int) (events []*L2BlockEventMonitor, err error) {
	dbTx := m.DB.Table(m.table).Where("block_event_type = ? AND status = ?", eventType, status).Find(&events).Order("l2_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitorsByEventTypeAndStatus] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitorsByEventTypeAndStatus] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return events, nil
}

/*
	Func: GetL2BlockEventMonitorsByEventTypeAndStatus
	Return: events []*L2BlockEventMonitor, err error
	Description: get l2TxEventMonitors by event type and status
*/
func (m *defaultL2BlockEventMonitorModel) GetL2BlockEventMonitorsByStatus(status int) (rowsAffected int64, events []*L2BlockEventMonitor, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", status).Find(&events).Order("block_event_type AND l2_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitorsByEventTypeAndStatus] %s", dbTx.Error)
		logx.Error(err)
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, events, nil
}

/*
	Func: GetL2BlockEventMonitorsByTxType
	Return: events []*L2BlockEventMonitor, err error
	Description: get l2TxEventMonitors by tx type
*/
func (m *defaultL2BlockEventMonitorModel) GetL2BlockEventMonitorsByTxType(txType uint8) (events []*L2BlockEventMonitor, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_type = ?", txType).Find(&events).Order("l2_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitorsByTxType] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitorsByTxType] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return events, nil
}

/*
	Func: GetL2BlockEventMonitorsByL2BlockHeight
	Return: events []*L2BlockEventMonitor, err error
	Description: get l2TxEventMonitors by l2 block height
*/
func (m *defaultL2BlockEventMonitorModel) GetL2BlockEventMonitorsByL2BlockHeight(l2BlockHeight int64) (events []*L2BlockEventMonitor, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height = ?", l2BlockHeight).Find(&events).Order("l2_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitorsByL2BlockHeight] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2BlockEventMonitor.go.GetL2BlockEventMonitorsByL2BlockHeight] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return events, nil
}
