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

package l1rolluptx

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	TableName = "l1_rollup_tx"

	StatusPending = 1
	StatusHandled = 2

	TxTypeCommit           = 1
	TxTypeVerifyAndExecute = 2
)

type (
	L1RollupTxModel interface {
		CreateL1RollupTxTable() error
		DropL1RollupTxTable() error
		CreateL1RollupTx(tx *L1RollupTx) error
		GetLatestHandledTx(txType int64) (tx *L1RollupTx, err error)
		GetLatestPendingTx(txType int64) (tx *L1RollupTx, err error)
		GetL1RollupTxsByStatus(txStatus int) (txs []*L1RollupTx, err error)
		GetL1RollupTxsByHash(hash string) (txs []*L1RollupTx, err error)
		DeleteL1RollupTx(tx *L1RollupTx) error
		DeleteGreaterOrEqualToHeight(height int64, txType uint8) error
		UpdateL1RollupTxsStatusInTransact(tx *gorm.DB, txs []*L1RollupTx) error
		GetLatestByNonce(l1Nonce int64, txType int64) (tx *L1RollupTx, err error)
		GetRecentById(id uint, txType int64) (tx *L1RollupTx, err error)
		GetRecent2Transact(txType int64) (txs []*L1RollupTx, err error)
	}

	defaultL1RollupTxModel struct {
		table string
		DB    *gorm.DB
	}

	L1RollupTx struct {
		gorm.Model
		// txVerification hash
		L1TxHash string `gorm:"index"`
		// txVerification status, 1 - pending, 2 - handled
		TxStatus int `gorm:"index:idx_tx_status"`
		// txVerification type: commit / verify
		TxType uint8 `gorm:"index:idx_tx_status"`
		// layer-2 block height
		L2BlockHeight int64 `gorm:"index:l2_block_height"`
		// gas price
		GasPrice int64
		// gas used
		GasUsed uint64
		//l1 nonce
		L1Nonce int64 `gorm:"index:idx_l1_nonce"`
	}
)

func (*L1RollupTx) TableName() string {
	return TableName
}

func NewL1RollupTxModel(db *gorm.DB) L1RollupTxModel {
	return &defaultL1RollupTxModel{
		table: TableName,
		DB:    db,
	}
}

func (m *defaultL1RollupTxModel) CreateL1RollupTxTable() error {
	return m.DB.AutoMigrate(L1RollupTx{})
}

func (m *defaultL1RollupTxModel) DropL1RollupTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultL1RollupTxModel) CreateL1RollupTx(tx *L1RollupTx) error {
	dbTx := m.DB.Table(m.table).Create(tx)
	if dbTx.Error != nil {
		return dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateL1RollupTx
	}
	return nil
}

func (m *defaultL1RollupTxModel) GetL1RollupTxsByStatus(txStatus int) (txs []*L1RollupTx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ?", txStatus).Order("l2_block_height, tx_type").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultL1RollupTxModel) GetL1RollupTxsByHash(hash string) (txs []*L1RollupTx, err error) {
	dbTx := m.DB.Table(m.table).Where("l1_tx_hash = ?", hash).Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultL1RollupTxModel) DeleteL1RollupTx(rollupTx *L1RollupTx) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		dbTx := tx.Table(m.table).Where("id = ?", rollupTx.ID).Delete(&rollupTx)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToDeleteL1RollupTx
		}
		return nil
	})
}

func (m *defaultL1RollupTxModel) DeleteGreaterOrEqualToHeight(height int64, txType uint8) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		dbTx := tx.Table(m.table).Unscoped().Where("l2_block_height >= ? and tx_type= ?", height, txType).Delete(&L1RollupTx{})
		if dbTx.Error != nil {
			return dbTx.Error
		}
		return nil
	})
}

func (m *defaultL1RollupTxModel) GetLatestHandledTx(txType int64) (tx *L1RollupTx, err error) {
	tx = &L1RollupTx{}

	dbTx := m.DB.Table(m.table).Where("tx_type = ? AND tx_status = ?", txType, StatusHandled).Order("l2_block_height desc").Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultL1RollupTxModel) GetLatestPendingTx(txType int64) (tx *L1RollupTx, err error) {
	tx = &L1RollupTx{}

	dbTx := m.DB.Table(m.table).Where("tx_type = ? AND tx_status = ?", txType, StatusPending).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultL1RollupTxModel) UpdateL1RollupTxsStatusInTransact(tx *gorm.DB, txs []*L1RollupTx) error {
	for _, rollupTx := range txs {
		dbTx := tx.Table(TableName).Where("id = ?", rollupTx.ID).
			Updates(map[string]interface{}{
				"tx_status": rollupTx.TxStatus,
				"gas_used":  rollupTx.GasUsed,
			})
		if dbTx.Error != nil {
			return dbTx.Error
		}
	}
	return nil
}

func (m *defaultL1RollupTxModel) GetLatestByNonce(l1Nonce int64, txType int64) (tx *L1RollupTx, err error) {
	tx = &L1RollupTx{}

	dbTx := m.DB.Table(m.table).Unscoped().Where("tx_type = ? AND l1_nonce = ?", txType, l1Nonce).Order("id desc").Limit(1).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultL1RollupTxModel) GetRecentById(id uint, txType int64) (tx *L1RollupTx, err error) {
	dbTx := m.DB.Table(m.table).Unscoped().Where("tx_type = ? AND id > ?", txType, id).
		Order("id asc").Limit(1).Find(&tx)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultL1RollupTxModel) GetRecent2Transact(txType int64) (txs []*L1RollupTx, err error) {
	dbTx := m.DB.Table(m.table).Unscoped().Where("tx_type = ?", txType).
		Order("id desc").Limit(2).Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}
