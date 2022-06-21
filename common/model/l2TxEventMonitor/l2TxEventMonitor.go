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

package l2TxEventMonitor

import (
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
)

type (
	L2TxEventMonitorModel interface {
		CreateL2TxEventMonitorTable() error
		DropL2TxEventMonitorTable() error
		CreateL2TxEventMonitor(tx *L2TxEventMonitor) (bool, error)
		CreateL2TxEventMonitorsInBatches(l2TxEventMonitors []*L2TxEventMonitor) (rowsAffected int64, err error)
		GetL2TxEventMonitorsByStatus(status int) (txs []*L2TxEventMonitor, err error)
		GetL2TxEventMonitorsBySenderAddress(senderAddr string) (txs []*L2TxEventMonitor, err error)
		GetL2TxEventMonitorsByTxType(txType uint8) (txs []*L2TxEventMonitor, err error)
		CreateMempoolAndActiveAccount(
			pendingNewAccount []*account.Account,
			pendingNewMempoolTxs []*mempool.MempoolTx,
			pendingNewLiquidityInfos []*liquidity.Liquidity,
			pendingNewNfts []*nft.L2Nft,
			pendingUpdateL2Events []*L2TxEventMonitor) (err error)
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
	Func: CreateL2TxEventMonitor
	Params: asset *L2TxEventMonitor
	Return: bool, error
	Description: create L2TxEventMonitor txVerification
*/
func (m *defaultL2TxEventMonitorModel) CreateL2TxEventMonitor(tx *L2TxEventMonitor) (bool, error) {
	dbTx := m.DB.Table(m.table).Create(tx)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.CreateL2TxEventMonitor] %s", dbTx.Error)
		logx.Error(err)
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		ErrInvalidL2TxEventMonitor := errors.New("invalid l2TxEventMonitor")
		err := fmt.Sprintf("[l2TxEventMonitor.CreateL2TxEventMonitor] %s", ErrInvalidL2TxEventMonitor)
		logx.Error(err)
		return false, ErrInvalidL2TxEventMonitor
	}
	return true, nil
}

/*
	Func: CreateL2TxEventMonitorsInBatches
	Params: []*L2TxEventMonitor
	Return: rowsAffected int64, err error
	Description: create L2TxEventMonitor batches
*/
func (m *defaultL2TxEventMonitorModel) CreateL2TxEventMonitorsInBatches(l2TxEventMonitors []*L2TxEventMonitor) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(l2TxEventMonitors, len(l2TxEventMonitors))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.CreateL1AssetsMonitorInBatches] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

/*
	GetL2TxEventMonitors: get all L2TxEventMonitors
*/
func (m *defaultL2TxEventMonitorModel) GetL2TxEventMonitors() (txs []*L2TxEventMonitor, err error) {
	dbTx := m.DB.Table(m.table).Find(&txs).Order("l1_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitors] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitors] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
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
	dbTx := m.DB.Table(m.table).Where("status = ?", status).Find(&txs).Order("request_id")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByStatus] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByStatus] %s", ErrNotFound)
		logx.Info(err)
		return nil, ErrNotFound
	}
	return txs, nil
}

/*
	Func: GetL2TxEventMonitorsByAccountName
	Return: txVerification []*L2TxEventMonitor, err error
	Description: get l2TxEventMonitors by account name
*/
func (m *defaultL2TxEventMonitorModel) GetL2TxEventMonitorsBySenderAddress(senderAddr string) (txs []*L2TxEventMonitor, err error) {
	// todo order id
	dbTx := m.DB.Table(m.table).Where("sender_address = ?", senderAddr).Find(&txs).Order("l1_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsBySenderAddress] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsBySenderAddress] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, nil
}

/*
	Func: GetL2TxEventMonitorsByTxType
	Return: txVerification []*L2TxEventMonitor, err error
	Description: get l2TxEventMonitors by txVerification type
*/
func (m *defaultL2TxEventMonitorModel) GetL2TxEventMonitorsByTxType(txType uint8) (txs []*L2TxEventMonitor, err error) {
	// todo order id
	dbTx := m.DB.Table(m.table).Where("tx_type = ?", txType).Find(&txs).Order("l1_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByTxType] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByTxType] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, nil
}

func (m *defaultL2TxEventMonitorModel) CreateMempoolAndActiveAccount(
	pendingNewAccount []*account.Account,
	pendingNewMempoolTxs []*mempool.MempoolTx,
	pendingNewLiquidityInfos []*liquidity.Liquidity,
	pendingNewNfts []*nft.L2Nft,
	pendingUpdateL2Events []*L2TxEventMonitor,
) (err error) {
	err = m.DB.Transaction(
		func(tx *gorm.DB) error { //transact
			dbTx := tx.Table(account.TableName).CreateInBatches(pendingNewAccount, len(pendingNewAccount))
			if dbTx.Error != nil {
				logx.Errorf("[CreateMempoolAndActiveAccount] unable to create pending new account: %s", err.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(pendingNewAccount)) {
				logx.Errorf("[CreateMempoolAndActiveAccount] invalid new account")
				return errors.New("[CreateMempoolAndActiveAccount] invalid new account")
			}
			dbTx = tx.Table(mempool.TableName).CreateInBatches(pendingNewMempoolTxs, len(pendingNewMempoolTxs))
			if dbTx.Error != nil {
				logx.Errorf("[CreateMempoolAndActiveAccount] unable to create pending new mempool txs: %s", err.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(pendingNewMempoolTxs)) {
				logx.Errorf("[CreateMempoolAndActiveAccount] invalid new mempool txs")
				return errors.New("[CreateMempoolAndActiveAccount] invalid new mempool txs")
			}
			if len(pendingNewLiquidityInfos) != 0 {
				dbTx = tx.Table(liquidity.TableName).CreateInBatches(pendingNewLiquidityInfos, len(pendingNewLiquidityInfos))
				if dbTx.Error != nil {
					logx.Errorf("[CreateMempoolAndActiveAccount] unable to create pending new liquidity infos: %s", err.Error())
					return dbTx.Error
				}
				if dbTx.RowsAffected != int64(len(pendingNewLiquidityInfos)) {
					logx.Errorf("[CreateMempoolAndActiveAccount] invalid new liquidity infos")
					return errors.New("[CreateMempoolAndActiveAccount] invalid new liquidity infos")
				}
			}
			if len(pendingNewNfts) != 0 {
				dbTx = tx.Table(nft.TableName).CreateInBatches(pendingNewNfts, len(pendingNewNfts))
				if dbTx.Error != nil {
					logx.Errorf("[CreateMempoolAndActiveAccount] unable to create pending new nft infos: %s", err.Error())
					return dbTx.Error
				}
				if dbTx.RowsAffected != int64(len(pendingNewNfts)) {
					logx.Errorf("[CreateMempoolAndActiveAccount] invalid new nft infos")
					return errors.New("[CreateMempoolAndActiveAccount] invalid new nft infos")
				}
			}
			for _, pendingUpdateL2Event := range pendingUpdateL2Events {
				dbTx = tx.Table(m.table).Where("id = ?", pendingUpdateL2Event.ID).Select("*").Updates(&pendingUpdateL2Event)
				if dbTx.Error != nil {
					logx.Errorf("[CreateMempoolAndActiveAccount] unable to update l2 tx event: %s", err.Error())
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					logx.Errorf("[CreateMempoolAndActiveAccount] invalid l2 tx event")
					return errors.New("[CreateMempoolAndActiveAccount] invalid l2 tx event")
				}
			}
			return nil
		})
	return err
}

func (m *defaultL2TxEventMonitorModel) GetLastHandledRequestId() (requestId int64, err error) {
	var event *L2TxEventMonitor
	dbTx := m.DB.Table(m.table).Where("status = ?", HandledStatus).Order("request_id desc").Find(&event)
	if dbTx.Error != nil {
		logx.Errorf("[GetLastHandledRequestId] unable to get last handled request id: %s", dbTx.Error)
		return -1, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return -1, nil
	}
	return event.RequestId, nil
}
