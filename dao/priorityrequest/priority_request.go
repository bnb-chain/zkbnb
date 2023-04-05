/*
 * Copyright © 2021 ZkBNB Protocol
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
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	TableName = "priority_request"

	PendingStatus = 1
	HandledStatus = 2
)

type (
	PriorityRequestModel interface {
		CreatePriorityRequestTable() error
		DropPriorityRequestTable() error
		GetPriorityRequestsByStatus(status int) (txs []*PriorityRequest, err error)
		GetLatestHandledRequestId() (requestId int64, err error)
		UpdateHandledPriorityRequestsInTransact(tx *gorm.DB, requests []*PriorityRequest) (err error)
		CreatePriorityRequestsInTransact(tx *gorm.DB, requests []*PriorityRequest) (err error)
		GetPriorityRequestsByL2TxHash(txHash string) (tx *PriorityRequest, err error)
		GetPriorityRequestsByTxTypes(requestId int64, txTypes []int64) (txs []*PriorityRequest, err error)
	}

	defaultPriorityRequestModel struct {
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
		// L2TxHash for the relation to tx table
		L2TxHash string `gorm:"index"`
	}
)

func (*PriorityRequest) TableName() string {
	return TableName
}

func NewPriorityRequestModel(db *gorm.DB) PriorityRequestModel {
	return &defaultPriorityRequestModel{
		table: TableName,
		DB:    db,
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
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, dbTx.Error
}

func (m *defaultPriorityRequestModel) GetPriorityRequestsByStatus(status int) (txs []*PriorityRequest, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", status).Order("request_id").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultPriorityRequestModel) GetLatestHandledRequestId() (requestId int64, err error) {
	var event *PriorityRequest
	dbTx := m.DB.Table(m.table).Where("status = ?", HandledStatus).Order("request_id desc").Find(&event)
	if dbTx.Error != nil {
		return -1, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return -1, nil
	}
	return event.RequestId, nil
}

func (m *defaultPriorityRequestModel) UpdateHandledPriorityRequestsInTransact(tx *gorm.DB, requests []*PriorityRequest) (err error) {
	for _, request := range requests {
		dbTx := tx.Model(&PriorityRequest{}).Where("id = ?", request.ID).Updates(
			map[string]interface{}{
				"status":     HandledStatus,
				"l2_tx_hash": request.L2TxHash,
			},
		)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected != 1 {
			return types.DbErrFailToUpdatePriorityRequest
		}
	}
	return nil
}

func (m *defaultPriorityRequestModel) CreatePriorityRequestsInTransact(tx *gorm.DB, requests []*PriorityRequest) (err error) {
	dbTx := tx.Table(m.table).CreateInBatches(requests, len(requests))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(requests)) {
		return types.DbErrFailToCreatePriorityRequest
	}
	return nil
}

func (m *defaultPriorityRequestModel) GetPriorityRequestsByL2TxHash(txHash string) (tx *PriorityRequest, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_tx_hash = ?", txHash).Limit(1).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}

	return tx, nil
}

func (m *defaultPriorityRequestModel) GetPriorityRequestsByTxTypes(requestId int64, txTypes []int64) (txs []*PriorityRequest, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_type in ? and request_id>=?", txTypes, requestId).Limit(100).Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}
