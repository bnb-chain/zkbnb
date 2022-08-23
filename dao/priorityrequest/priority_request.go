/*
 * Copyright © 2021 ZkBAS Protocol
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

package priorityrequest

import (
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/types"
)

type (
	PriorityRequestModel interface {
		CreatePriorityRequestTable() error
		DropPriorityRequestTable() error
		GetPriorityRequestsByStatus(status int) (txs []*PriorityRequest, err error)
		CreateMempoolTxsAndUpdateRequests(pendingNewMempoolTxs []*mempool.MempoolTx, pendingUpdateRequests []*PriorityRequest) (err error)
		GetLatestHandledRequestId() (requestId int64, err error)
	}

	defaultPriorityRequestModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	PriorityRequest struct {
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

func (*PriorityRequest) TableName() string {
	return TableName
}

func NewPriorityRequestModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) PriorityRequestModel {
	return &defaultPriorityRequestModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TableName,
		DB:         db,
	}
}

func (m *defaultPriorityRequestModel) CreatePriorityRequestTable() error {
	return m.DB.AutoMigrate(PriorityRequest{})
}

func (m *defaultPriorityRequestModel) DropPriorityRequestTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultPriorityRequestModel) GetL2TxEventMonitors() (txs []*PriorityRequest, err error) {
	dbTx := m.DB.Table(m.table).Find(&txs).Order("l1_block_height")
	if dbTx.Error != nil {
		logx.Errorf("find l2 tx events error,  err=%s", dbTx.Error.Error())
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, dbTx.Error
}

func (m *defaultPriorityRequestModel) GetPriorityRequestsByStatus(status int) (txs []*PriorityRequest, err error) {
	// todo order id
	dbTx := m.DB.Table(m.table).Where("status = ?", status).Order("request_id").Find(&txs)
	if dbTx.Error != nil {
		logx.Errorf("find l2 tx events error,  err=%s", dbTx.Error.Error())
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultPriorityRequestModel) CreateMempoolTxsAndUpdateRequests(newMempoolTxs []*mempool.MempoolTx, toUpdateL2Events []*PriorityRequest) (err error) {
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

func (m *defaultPriorityRequestModel) GetLatestHandledRequestId() (requestId int64, err error) {
	var event *PriorityRequest
	dbTx := m.DB.Table(m.table).Where("status = ?", HandledStatus).Order("request_id desc").Find(&event)
	if dbTx.Error != nil {
		logx.Errorf("unable to get latest handled request id: %s", dbTx.Error.Error())
		return -1, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return -1, nil
	}
	return event.RequestId, nil
}
