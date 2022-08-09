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

package l2TxEventMonitor

import (
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	L2TxEventMonitorModel interface {
		CreateL2TxEventMonitorTable() error
		DropL2TxEventMonitorTable() error
		GetL2TxEventMonitorsByStatus(status int) (txs []*L2TxEventMonitor, err error)
		CreateMempoolTxsAndUpdateL2Events(pendingNewMempoolTxs []*mempool.MempoolTx, pendingUpdateL2Events []*L2TxEventMonitor) (err error)
		GetLastHandledRequestId() (requestId int64, err error)
	}

	defaultL2TxEventMonitorModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2TxEventMonitor struct {
		gorm.Model
		// related txVerification hash
		L1TxHash string
		// related block height
		L1BlockHeight int64
		// sender
		SenderAddress string
		// request id
		RequestId int64
		// tx type
		TxType int64
		// pub data
		Pubdata string
		// expirationBlock
		ExpirationBlock int64
		// status
		Status int
	}
)

func (*L2TxEventMonitor) TableName() string {
	return TableName
}

func NewL2TxEventMonitorModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2TxEventMonitorModel {
	return &defaultL2TxEventMonitorModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TableName,
		DB:         db,
	}
}

/*
	Func: CreateL2TxEventMonitorTable
	Params:
	Return: err error
	Description: create l2 txVerification event monitor table
*/
func (m *defaultL2TxEventMonitorModel) CreateL2TxEventMonitorTable() error {
	return m.DB.AutoMigrate(L2TxEventMonitor{})
}

/*
	Func: DropL2TxEventMonitorTable
	Params:
	Return: err error
	Description: drop l2 txVerification event monitor table
*/
func (m *defaultL2TxEventMonitorModel) DropL2TxEventMonitorTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	GetL2TxEventMonitors: get all L2TxEventMonitors
*/
func (m *defaultL2TxEventMonitorModel) GetL2TxEventMonitors() (txs []*L2TxEventMonitor, err error) {
	dbTx := m.DB.Table(m.table).Find(&txs).Order("l1_block_height")
	if dbTx.Error != nil {
		logx.Errorf("find l2 tx events error,  err=%s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return txs, dbTx.Error
}

/*
	Func: GetPendingL2TxEventMonitors
	Return: txVerification []*L2TxEventMonitor, err error
	Description: get pending l2TxEventMonitors
*/
func (m *defaultL2TxEventMonitorModel) GetL2TxEventMonitorsByStatus(status int) (txs []*L2TxEventMonitor, err error) {
	// todo order id
	dbTx := m.DB.Table(m.table).Where("status = ?", status).Order("request_id").Find(&txs)
	if dbTx.Error != nil {
		logx.Errorf("find l2 tx events error,  err=%s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultL2TxEventMonitorModel) CreateMempoolTxsAndUpdateL2Events(newMempoolTxs []*mempool.MempoolTx, toUpdateL2Events []*L2TxEventMonitor) (err error) {
	err = m.DB.Transaction(
		func(tx *gorm.DB) error {
			dbTx := tx.Table(mempool.MempoolTableName).CreateInBatches(newMempoolTxs, len(newMempoolTxs))
			if dbTx.Error != nil {
				logx.Errorf("unable to create pending new mempool txs: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(newMempoolTxs)) {
				logx.Errorf("create mempool txs error, rowsToCreate=%d, rowsCreated=%d",
					len(newMempoolTxs), dbTx.RowsAffected)
				return errors.New("create mempool txs error")
			}

			eventIds := make([]uint, 0, len(toUpdateL2Events))
			for _, l2Event := range toUpdateL2Events {
				eventIds = append(eventIds, l2Event.ID)
			}
			dbTx = tx.Table(m.table).Where("id in ?", eventIds).Update("status", HandledStatus)
			if dbTx.Error != nil {
				logx.Errorf("unable to update l2 tx event: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(eventIds)) {
				logx.Errorf("update l2 events error, rowsToUpdate=%d, rowsUpdated=%d",
					len(eventIds), dbTx.RowsAffected)
				return errors.New("update l2 events error")
			}
			return nil
		})
	return err
}

func (m *defaultL2TxEventMonitorModel) GetLastHandledRequestId() (requestId int64, err error) {
	var event *L2TxEventMonitor
	dbTx := m.DB.Table(m.table).Where("status = ?", HandledStatus).Order("request_id desc").Find(&event)
	if dbTx.Error != nil {
		logx.Errorf("unable to get last handled request id: %s", dbTx.Error.Error())
		return -1, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return -1, nil
	}
	return event.RequestId, nil
}
