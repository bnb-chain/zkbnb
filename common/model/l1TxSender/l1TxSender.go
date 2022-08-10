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

package l1TxSender

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	L1TxSenderModel interface {
		CreateL1TxSenderTable() error
		DropL1TxSenderTable() error
		CreateL1TxSender(tx *L1TxSender) (bool, error)
		GetLatestHandledBlock(txType int64) (txSender *L1TxSender, err error)
		GetLatestPendingBlock(txType int64) (txSender *L1TxSender, err error)
		GetL1TxSendersByTxStatus(txStatus int) (txs []*L1TxSender, err error)
		DeleteL1TxSender(sender *L1TxSender) error
		UpdateSentTxs(
			pendingUpdateSenders []*L1TxSender,
			pendingUpdateProofSenderStatus map[int64]int,
		) (err error)
	}

	defaultL1TxSenderModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L1TxSender struct {
		gorm.Model
		// txVerification hash
		L1TxHash string
		// txVerification status, 1 - pending, 2 - handled
		TxStatus int
		// txVerification type: commit / verify
		TxType uint8
		// layer-2 block height
		L2BlockHeight int64
	}
)

func (*L1TxSender) TableName() string {
	return TableName
}

func NewL1TxSenderModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L1TxSenderModel {
	return &defaultL1TxSenderModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TableName,
		DB:         db,
	}
}

/*
	Func: CreateL1TxSenderTable
	Params:
	Return: err error
	Description: create l2 txVerification event monitor table
*/
func (m *defaultL1TxSenderModel) CreateL1TxSenderTable() error {
	return m.DB.AutoMigrate(L1TxSender{})
}

/*
	Func: DropL1TxSenderTable
	Params:
	Return: err error
	Description: drop l2 txVerification event monitor table
*/
func (m *defaultL1TxSenderModel) DropL1TxSenderTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateL1TxSender
	Params: asset *L1TxSender
	Return: bool, error
	Description: create L1TxSender txVerification
*/
func (m *defaultL1TxSenderModel) CreateL1TxSender(tx *L1TxSender) (bool, error) {
	dbTx := m.DB.Table(m.table).Create(tx)
	if dbTx.Error != nil {
		logx.Errorf("create l1 tx sender error, err: %s", dbTx.Error.Error())
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return false, errors.New("invalid l1TxSender")
	}
	return true, nil
}

/*
	Func: GetL1TxSendersByTxStatus
	Return: txVerification []*L1TxSender, err error
	Description: get L1TxSender by txVerification status
*/
func (m *defaultL1TxSenderModel) GetL1TxSendersByTxStatus(txStatus int) (txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ?", txStatus).Order("l2_block_height, tx_type").Find(&txs)
	if dbTx.Error != nil {
		logx.Errorf("get l1 tx senders by status error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultL1TxSenderModel) DeleteL1TxSender(sender *L1TxSender) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Where("id = ?", sender.ID).Delete(&sender)
		if dbTx.Error != nil {
			logx.Errorf("delete l1 tx sender error, err: %s", dbTx.Error.Error())
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return errors.New("delete invalid sender")
		}
		return nil
	})
}

func (m *defaultL1TxSenderModel) UpdateSentTxs(
	pendingUpdateSenders []*L1TxSender,
	pendingUpdateProofSenderStatus map[int64]int,
) (err error) {
	const (
		Txs = "Txs"
	)
	err = m.DB.Transaction(func(tx *gorm.DB) error {
		// update sender
		for _, pendingUpdateSender := range pendingUpdateSenders {
			dbTx := tx.Table(TableName).Where("id = ?", pendingUpdateSender.ID).
				Select("*").
				Updates(&pendingUpdateSender)
			if dbTx.Error != nil {
				logx.Errorf("update tx sender error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				senderInfo, err := json.Marshal(pendingUpdateSender)
				if err != nil {
					logx.Errorf("marshal tx sender error, err: %s", err.Error())
					return err
				}
				logx.Errorf("invalid sender:  %s", string(senderInfo))
				return errors.New("invalid sender")
			}
		}
		// modify proofSender Status
		for blockHeight, newStatus := range pendingUpdateProofSenderStatus {
			var row *proofSender.ProofSender
			dbTx := tx.Table(proofSender.TableName).Where("block_number = ?", blockHeight).Find(&row)
			if dbTx.Error != nil {
				logx.Errorf("update proof sender error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return fmt.Errorf("no such proof. height: %d", blockHeight)
			}
			dbTx = tx.Model(&row).
				Select("status").
				Updates(&proofSender.ProofSender{Status: int64(newStatus)})
			if dbTx.Error != nil {
				logx.Errorf("update proof sender error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return fmt.Errorf("update no proof: %d", row.BlockNumber)
			}
		}
		return nil
	})
	return err
}

func (m *defaultL1TxSenderModel) GetLatestHandledBlock(txType int64) (txSender *L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_type = ? AND tx_status = ?", txType, HandledStatus).Order("l2_block_height desc").Find(&txSender)
	if dbTx.Error != nil {
		logx.Errorf("unable to get latest handled block: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return txSender, nil
}

func (m *defaultL1TxSenderModel) GetLatestPendingBlock(txType int64) (txSender *L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_type = ? AND tx_status = ?", txType, PendingStatus).Find(&txSender)
	if dbTx.Error != nil {
		logx.Errorf("unable to get latest pending block: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return txSender, nil
}
