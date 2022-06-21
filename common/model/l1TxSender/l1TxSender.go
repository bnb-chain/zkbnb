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

package l1TxSender

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/proofSender"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	L1TxSenderModel interface {
		CreateL1TxSenderTable() error
		DropL1TxSenderTable() error
		CreateL1TxSender(tx *L1TxSender) (bool, error)
		CreateL1TxSendersInBatches(txs []*L1TxSender) (rowsAffected int64, err error)
		GetL1TxSenders() (txs []*L1TxSender, err error)
		GetLatestHandledBlock(txType int64) (txSender *L1TxSender, err error)
		GetLatestL1TxSender() (blockInfo *L1TxSender, err error)
		GetLatestPendingBlock(txType int64) (txSender *L1TxSender, err error)
		GetL1TxSendersByTxStatus(txStatus int) (txs []*L1TxSender, err error)
		GetL1TxSendersByTxTypeAndStatus(txType uint8, txStatus int) (rowsAffected int64, txs []*L1TxSender, err error)
		GetL1TxSendersByTxHashAndTxType(txHash string, txType uint8) (rowsAffected int64, txs []*L1TxSender, err error)
		DeleteL1TxSender(sender *L1TxSender) error
		UpdateRelatedEventsAndResetRelatedAssetsAndTxs(
			pendingUpdateBlocks []*block.Block,
			pendingUpdateSenders []*L1TxSender,
			pendingUpdateMempoolTxs []*mempool.MempoolTx,
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
		err := fmt.Sprintf("[l1TxSender.CreateL1TxSender] %s", dbTx.Error)
		logx.Error(err)
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		ErrInvalidL1TxSender := errors.New("invalid l1TxSender")
		err := fmt.Sprintf("[l1TxSender.CreateL1TxSender] %s", ErrInvalidL1TxSender)
		logx.Error(err)
		return false, ErrInvalidL1TxSender
	}
	return true, nil
}

/*
	Func: CreateL1TxSendersInBatches
	Params: []*L1TxSender
	Return: rowsAffected int64, err error
	Description: create L1TxSender batches
*/
func (m *defaultL1TxSenderModel) CreateL1TxSendersInBatches(txs []*L1TxSender) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(txs, len(txs))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.CreateL1TxSendersInBatches] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

/*
	GetL1TxSenders: get all L1TxSenders
*/
func (m *defaultL1TxSenderModel) GetL1TxSenders() (txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Find(&txs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSenders] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSenders] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, dbTx.Error
}

/*
	Func: GetLatestL1TxSender
	Return: txVerification []*L1TxSender, err error
	Description: get latest l1 block monitor info
*/
func (m *defaultL1TxSenderModel) GetLatestL1TxSender() (blockInfo *L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).First(&blockInfo)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetLatestL1TxSender] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1TxSender.GetLatestL1TxSender] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return blockInfo, nil
}

/*
	Func: GetL1TxSendersByTxStatus
	Return: txVerification []*L1TxSender, err error
	Description: get L1TxSender by txVerification status
*/
func (m *defaultL1TxSenderModel) GetL1TxSendersByTxStatus(txStatus int) (txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ?", txStatus).Order("l2_block_height, tx_type").Find(&txs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByTxStatus] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByTxStatus] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, nil
}

/*
	Func: GetL1TxSendersByTxStatus
	Return: txVerification []*L1TxSender, err error
	Description: get L1TxSender by txVerification type and status
*/
func (m *defaultL1TxSenderModel) GetL1TxSendersByTxTypeAndStatus(txType uint8, txStatus int) (rowsAffected int64, txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_type = ? AND tx_status = ?", txType, txStatus).Find(&txs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByTxTypeAndStatus] %s", dbTx.Error)
		logx.Error(err)
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, txs, nil
}

func (m *defaultL1TxSenderModel) GetL1TxSendersByTxHashAndTxType(txHash string, txType uint8) (rowsAffected int64, txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("l1_tx_hash = ? AND tx_type = ?", txHash, txType).Find(&txs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByTxHashAndTxType] %s", dbTx.Error)
		logx.Error(err)
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, txs, nil
}

func (m *defaultL1TxSenderModel) DeleteL1TxSender(sender *L1TxSender) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Where("id = ?", sender.ID).Delete(&sender)
		if dbTx.Error != nil {
			logx.Error("[l1TxSender.DeleteL1TxSender] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Error("[l1TxSender.DeleteL1TxSender] Delete invalid sender")
			return errors.New("[l1TxSender.DeleteL1TxSender] delete invalid sender")
		}
		return nil
	})
}

func (m *defaultL1TxSenderModel) UpdateRelatedEventsAndResetRelatedAssetsAndTxs(
	pendingUpdateBlocks []*block.Block,
	pendingUpdateSenders []*L1TxSender,
	pendingUpdateMempoolTxs []*mempool.MempoolTx,
	pendingUpdateProofSenderStatus map[int64]int,
) (err error) {
	const (
		Txs = "Txs"
	)
	err = m.DB.Transaction(func(tx *gorm.DB) error {
		// update blocks
		for _, pendingUpdateBlock := range pendingUpdateBlocks {
			dbTx := tx.Table(block.BlockTableName).Where("id = ?", pendingUpdateBlock.ID).
				Omit(Txs).
				Select("*").
				Updates(&pendingUpdateBlock)
			if dbTx.Error != nil {
				err := fmt.Sprintf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", dbTx.Error)
				logx.Error(err)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				blocksInfo, err := json.Marshal(pendingUpdateBlocks)
				if err != nil {
					res := fmt.Sprintf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", err)
					logx.Error(res)
					return err
				}
				logx.Error("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s" + "Invalid block:  " + string(blocksInfo))
				return errors.New("Invalid blocks:  " + string(blocksInfo))
			}
		}
		// update sender
		for _, pendingUpdateSender := range pendingUpdateSenders {
			dbTx := tx.Table(TableName).Where("id = ?", pendingUpdateSender.ID).
				Select("*").
				Updates(&pendingUpdateSender)
			if dbTx.Error != nil {
				err := fmt.Sprintf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", dbTx.Error)
				logx.Error(err)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				senderInfo, err := json.Marshal(pendingUpdateSender)
				if err != nil {
					res := fmt.Sprintf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", err)
					logx.Error(res)
					return err
				}
				logx.Error("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s" + "Invalid sender:  " + string(senderInfo))
				return errors.New("Invalid sender:  " + string(senderInfo))
			}
		}
		// delete mempool txs
		for _, pendingDeleteMempoolTx := range pendingUpdateMempoolTxs {
			for _, detail := range pendingDeleteMempoolTx.MempoolDetails {
				dbTx := tx.Table(mempool.DetailTableName).Where("id = ?", detail.ID).Delete(&detail)
				if dbTx.Error != nil {
					logx.Errorf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", dbTx.Error)
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					logx.Errorf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Delete Invalid Mempool Tx")
					return errors.New("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Delete Invalid Mempool Tx")
				}
			}
			dbTx := tx.Table(mempool.MempoolTableName).Where("id = ?", pendingDeleteMempoolTx.ID).Delete(&pendingDeleteMempoolTx)
			if dbTx.Error != nil {
				logx.Errorf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Error("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Delete Invalid Mempool Tx")
				return errors.New("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Delete Invalid Mempool Tx")
			}
		}
		// modify proofSender Status
		for blockHeight, newStatus := range pendingUpdateProofSenderStatus {
			var row *proofSender.ProofSender
			dbTx := tx.Table(proofSender.TableName).Where("block_number = ?", blockHeight).Find(&row)
			if dbTx.Error != nil {
				logx.Errorf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Error(fmt.Sprintf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] No such proof. Height: %d", blockHeight))
				return errors.New(fmt.Sprintf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] No such proof. Height: %d", blockHeight))
			}
			dbTx = tx.Model(&row).
				Select("status").
				Updates(&proofSender.ProofSender{Status: int64(newStatus)})
			if dbTx.Error != nil {
				logx.Errorf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Error(fmt.Sprintf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Update No Proof: %d", row.BlockNumber))
				return errors.New(fmt.Sprintf("[UpdateRelatedEventsAndResetRelatedAssetsAndTxs] Update No Proof: %d", row.BlockNumber))
			}
		}
		return nil
	})
	return err
}

func (m *defaultL1TxSenderModel) GetLatestHandledBlock(txType int64) (txSender *L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_type = ? AND tx_status = ?", txType, HandledStatus).Order("l2_block_height desc").Find(&txSender)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestHandledBlock] unable to get latest handled block: %s", err.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return txSender, nil
}

func (m *defaultL1TxSenderModel) GetLatestPendingBlock(txType int64) (txSender *L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_type = ? AND tx_status = ?", txType, PendingStatus).Find(&txSender)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestHandledBlock] unable to get latest pending block: %s", err.Error())
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return txSender, nil
}
