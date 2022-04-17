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
	"errors"
	"fmt"

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
		GetLatestL1TxSender() (blockInfo *L1TxSender, err error)
		GetL1TxSendersByChainId(chainId uint8) (txs []*L1TxSender, err error)
		GetL1TxSendersByChainIdAndL2BlockHeight(chainId uint8, l2BlockHeight int64) (txs []*L1TxSender, err error)
		GetL1TxSendersByTxStatus(txStatus int) (txs []*L1TxSender, err error)
		GetL1TxSendersByChainIdAndTxStatus(chainId int64, txStatus int) (txs []*L1TxSender, err error)
		GetLatestHandledBlock(chainId int64, txType uint8) (tx *L1TxSender, err error)
		GetLatestPendingBlocks(chainId int64, txType uint8) (rowsAffected int64, txs []*L1TxSender, err error)
		GetL1TxSendersByTxTypeAndStatus(txType uint8, txStatus int) (rowsAffected int64, txs []*L1TxSender, err error)
		GetL1TxSendersByChainIdAndTxType(chainId uint8, txType uint8) (txs []*L1TxSender, err error)
		GetL1TxSendersByTxHashAndTxType(txHash string, txType uint8) (rowsAffected int64, txs []*L1TxSender, err error)
		DeleteL1TxSender(sender *L1TxSender) error
	}

	defaultL1TxSenderModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L1TxSender struct {
		gorm.Model
		// tx hash
		L1TxHash string
		// tx status, 1 - pending, 2 - handled
		TxStatus int
		// tx type: commit / verify / execute
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
	Description: create l2 tx event monitor table
*/
func (m *defaultL1TxSenderModel) CreateL1TxSenderTable() error {
	return m.DB.AutoMigrate(L1TxSender{})
}

/*
	Func: DropL1TxSenderTable
	Params:
	Return: err error
	Description: drop l2 tx event monitor table
*/
func (m *defaultL1TxSenderModel) DropL1TxSenderTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateL1TxSender
	Params: asset *L1TxSender
	Return: bool, error
	Description: create L1TxSender tx
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
	Return: txs []*L1TxSender, err error
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
	Func: GetL1TxSendersByChainId
	Return: txs []*L1TxSender, err error
	Description: get L1TxSender by chain id
*/
func (m *defaultL1TxSenderModel) GetL1TxSendersByChainId(chainId uint8) (txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ?", chainId).Find(&txs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByChainId] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByChainId] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, nil
}

/*
	Func: GetL1TxSendersByChainIdAndL2BlockHeight
	Return: txs []*L1TxSender, err error
	Description: get L1TxSender by chain id and l2 block height
*/
func (m *defaultL1TxSenderModel) GetL1TxSendersByChainIdAndL2BlockHeight(chainId uint8, l2BlockHeight int64) (txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ? AND l2_block_height = ?", chainId, l2BlockHeight).Find(&txs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByChainId] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByChainId] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, nil
}

/*
	Func: GetL1TxSendersByTxStatus
	Return: txs []*L1TxSender, err error
	Description: get L1TxSender by tx status
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
	Func: GetL1TxSendersByChainIdAndTxStatus
	Return: txs []*L1TxSender, err error
	Description: get L1TxSender by tx status
*/
func (m *defaultL1TxSenderModel) GetL1TxSendersByChainIdAndTxStatus(chainId int64, txStatus int) (txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ? AND tx_status = ?", chainId, txStatus).Find(&txs)
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
	Return: txs []*L1TxSender, err error
	Description: get L1TxSender by tx type and status
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

/*
	Func: GetL1TxSendersByChainIdAndTxTypeAndStatus
	Return: txs []*L1TxSender, err error
	Description: get L1TxSender by tx type and status
*/
func (m *defaultL1TxSenderModel) GetLatestHandledBlock(
	chainId int64, txType uint8,
) (
	tx *L1TxSender,
	err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ? AND tx_type = ? AND tx_status = ?", chainId, txType, HandledStatus).
		Order("l2_block_height desc").Find(&tx)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetLatestHandledBlock] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Info("[l1TxSender.GetLatestHandledBlock] no block info")
		return nil, nil
	}
	return tx, nil
}

func (m *defaultL1TxSenderModel) GetLatestPendingBlocks(chainId int64, txType uint8) (rowsAffected int64, txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ? AND tx_type = ? AND tx_status = ?", chainId, txType, PendingStatus).
		Order("l2_block_height desc").Find(&txs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetLatestPendingBlocks] %s", dbTx.Error)
		logx.Error(err)
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, txs, nil
}

/*
	Func: GetL1TxSendersByChainIdAndTxTypeAndStatusAndHigherThanHeight
	Return: txs []*L1TxSender, err error
	Description: get L1TxSender by tx type and status
*/
func (m *defaultL1TxSenderModel) GetL1TxSendersByChainIdAndTxTypeAndStatusAndHigherThanHeight(
	chainId int64, txType uint8, txStatus int, l2Height int64, maxSize int,
) (
	rowsAffected int64,
	txs []*L1TxSender,
	err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ? AND tx_type = ? AND tx_status = ? AND l2_block_height >= ?", chainId, txType, txStatus, l2Height).
		Order("l2_block_height").Limit(maxSize).Find(&txs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByTxTypeAndStatus] %s", dbTx.Error)
		logx.Error(err)
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, txs, nil
}

/*
	Func: GetL1TxSendersByChainIdAndTxType
	Return: txs []*L1TxSender, err error
	Description: get L1TxSender by chain id and tx type
*/
func (m *defaultL1TxSenderModel) GetL1TxSendersByChainIdAndTxType(chainId uint8, txType uint8) (txs []*L1TxSender, err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ? AND tx_type = ?", chainId, txType).Find(&txs)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByTxTypeAndStatus] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1TxSender.GetL1TxSendersByTxTypeAndStatus] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, nil
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
