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

package tx

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	MempoolTableName = `mempool_tx`
)

type (
	MempoolModel interface {
		CreateMempoolTxTable() error
		DropMempoolTxTable() error
		GetMempoolTxs(limit int64, offset int64) (mempoolTxs []*Tx, err error)
		GetMempoolTxsTotalCount() (count int64, err error)
		GetMempoolTxByTxHash(hash string) (mempoolTxs *Tx, err error)
		GetMempoolTxsByStatus(status int) (mempoolTxs []*Tx, err error)
		CreateMempoolTxs(mempoolTxs []*Tx) error
		GetPendingMempoolTxsByAccountIndex(accountIndex int64) (mempoolTxs []*Tx, err error)
		GetMaxNonceByAccountIndex(accountIndex int64) (nonce int64, err error)
		UpdateMempoolTxs(pendingUpdateMempoolTxs []*Tx, pendingDeleteMempoolTxs []*Tx) error
		CreateMempoolTxsInTransact(tx *gorm.DB, mempoolTxs []*Tx) error
		UpdateMempoolTxsInTransact(tx *gorm.DB, mempoolTxs []*Tx) error
		DeleteMempoolTxsInTransact(tx *gorm.DB, mempoolTxs []*Tx) error
	}

	defaultMempoolModel struct {
		table string
		DB    *gorm.DB
	}
)

func NewMempoolModel(db *gorm.DB) MempoolModel {
	return &defaultMempoolModel{
		table: MempoolTableName,
		DB:    db,
	}
}

func (m *defaultMempoolModel) CreateMempoolTxTable() error {
	return m.DB.AutoMigrate(Tx{})
}

func (m *defaultMempoolModel) DropMempoolTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultMempoolModel) GetMempoolTxs(limit int64, offset int64) (mempoolTxs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", StatusPending).Limit(int(limit)).Offset(int(offset)).Order("created_at desc, id desc").Find(&mempoolTxs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetMempoolTxsByStatus(status int) (txs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", status).Order("created_at, id").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return txs, nil
}

func (m *defaultMempoolModel) GetMempoolTxsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and deleted_at is NULL", StatusPending).Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultMempoolModel) GetMempoolTxByTxHash(hash string) (tx *Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_hash = ?", hash).Find(&tx)
	if dbTx.Error != nil {
		if dbTx.Error == types.DbErrNotFound {
			return tx, dbTx.Error
		} else {
			return nil, types.DbErrSqlOperation
		}
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultMempoolModel) CreateMempoolTxs(mempoolTxs []*Tx) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(mempoolTxs)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToCreateMempoolTx
		}
		return nil
	})
}

func (m *defaultMempoolModel) GetPendingMempoolTxsByAccountIndex(accountIndex int64) (txs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? AND account_index = ?", StatusPending, accountIndex).
		Order("created_at, id").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultMempoolModel) GetMaxNonceByAccountIndex(accountIndex int64) (nonce int64, err error) {
	dbTx := m.DB.Table(m.table).Select("nonce").Where("deleted_at is null and account_index = ?", accountIndex).Order("nonce desc").Limit(1).Find(&nonce)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return nonce, nil
}

func (m *defaultMempoolModel) UpdateMempoolTxs(pendingUpdateTxs []*Tx, pendingDeleteTxs []*Tx) (err error) {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact

		// update mempool
		for _, mempoolTx := range pendingUpdateTxs {
			dbTx := tx.Table(MempoolTableName).Where("id = ?", mempoolTx.ID).
				Select("*").
				Updates(&mempoolTx)
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return types.DbErrFailToUpdateMempoolTx
			}
		}
		for _, pendingDeleteMempoolTx := range pendingDeleteTxs {
			dbTx := tx.Table(MempoolTableName).Where("id = ?", pendingDeleteMempoolTx.ID).Delete(&pendingDeleteMempoolTx)
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return types.DbErrFailToDeleteMempoolTx
			}
		}

		return nil
	})
}

func (m *defaultMempoolModel) CreateMempoolTxsInTransact(tx *gorm.DB, mempoolTxs []*Tx) error {
	dbTx := tx.Table(m.table).CreateInBatches(mempoolTxs, len(mempoolTxs))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateMempoolTx
	}
	return nil
}

func (m *defaultMempoolModel) UpdateMempoolTxsInTransact(tx *gorm.DB, mempoolTxs []*Tx) error {
	for _, mempoolTx := range mempoolTxs {
		dbTx := tx.Table(m.table).Where("id = ?", mempoolTx.ID).
			Select("*").
			Updates(&mempoolTx)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToUpdateMempoolTx
		}
	}
	return nil
}

func (m *defaultMempoolModel) DeleteMempoolTxsInTransact(tx *gorm.DB, mempoolTxs []*Tx) error {
	for _, mempoolTx := range mempoolTxs {
		dbTx := tx.Table(m.table).Where("id = ?", mempoolTx.ID).Delete(&mempoolTx)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToDeleteMempoolTx
		}
	}
	return nil
}
