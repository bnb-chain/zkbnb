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

package mempool

import (
	"errors"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/types"
)

const (
	MempoolTableName = `mempool_tx`
)

const (
	PendingTxStatus = iota
	ExecutedTxStatus
	SuccessTxStatus
	FailTxStatus
)

type (
	MempoolModel interface {
		CreateMempoolTxTable() error
		DropMempoolTxTable() error
		GetMempoolTxsList(limit int64, offset int64) (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsTotalCount() (count int64, err error)
		GetMempoolTxByTxHash(hash string) (mempoolTxs *MempoolTx, err error)
		GetMempoolTxsByStatus(status int) (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, mempoolTxs []*MempoolTx, err error)
		CreateBatchedMempoolTxs(mempoolTxs []*MempoolTx) error
		GetPendingMempoolTxsByAccountIndex(accountIndex int64) (mempoolTxs []*MempoolTx, err error)
		GetMaxNonceByAccountIndex(accountIndex int64) (nonce int64, err error)
		UpdateMempoolTxs(pendingUpdateMempoolTxs []*MempoolTx, pendingDeleteMempoolTxs []*MempoolTx) error
	}

	defaultMempoolModel struct {
		table string
		DB    *gorm.DB
	}

	MempoolTx struct {
		gorm.Model
		TxHash        string `gorm:"uniqueIndex"`
		TxType        int64
		GasFeeAssetId int64
		GasFee        string
		NftIndex      int64
		PairIndex     int64
		AssetId       int64
		TxAmount      string
		NativeAddress string
		TxInfo        string
		ExtraInfo     string
		Memo          string
		AccountIndex  int64
		Nonce         int64
		ExpiredAt     int64
		L2BlockHeight int64
		Status        int `gorm:"index"` // 0: pending tx; 1: committed tx; 2: verified tx;
	}
)

func NewMempoolModel(db *gorm.DB) MempoolModel {
	return &defaultMempoolModel{
		table: MempoolTableName,
		DB:    db,
	}
}

func (*MempoolTx) TableName() string {
	return MempoolTableName
}

func (m *defaultMempoolModel) CreateMempoolTxTable() error {
	return m.DB.AutoMigrate(MempoolTx{})
}

func (m *defaultMempoolModel) DropMempoolTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultMempoolModel) GetMempoolTxsList(limit int64, offset int64) (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", PendingTxStatus).Limit(int(limit)).Offset(int(offset)).Order("created_at desc, id desc").Find(&mempoolTxs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetMempoolTxsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height = ?", l2BlockHeight).Find(&mempoolTxs)
	if dbTx.Error != nil {
		return 0, nil, types.DbErrSqlOperation
	}
	return dbTx.RowsAffected, mempoolTxs, nil
}

func (m *defaultMempoolModel) GetMempoolTxsByStatus(status int) (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", status).Order("created_at, id").Find(&mempoolTxs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetMempoolTxsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and deleted_at is NULL", PendingTxStatus).Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultMempoolModel) GetMempoolTxByTxHash(hash string) (mempoolTx *MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and tx_hash = ?", PendingTxStatus, hash).Find(&mempoolTx)
	if dbTx.Error != nil {
		if dbTx.Error == types.DbErrNotFound {
			return mempoolTx, dbTx.Error
		} else {
			return nil, types.DbErrSqlOperation
		}
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return mempoolTx, nil
}

func (m *defaultMempoolModel) CreateBatchedMempoolTxs(mempoolTxs []*MempoolTx) error {
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

func (m *defaultMempoolModel) GetMempoolTxsListByL2BlockHeight(blockHeight int64) (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and l2_block_height <= ?", SuccessTxStatus, blockHeight).Find(&mempoolTxs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}

	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetPendingMempoolTxsByAccountIndex(accountIndex int64) (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? AND account_index = ?", PendingTxStatus, accountIndex).
		Order("created_at, id").Find(&mempoolTxs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return mempoolTxs, nil
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

func (m *defaultMempoolModel) UpdateMempoolTxs(pendingUpdateMempoolTxs []*MempoolTx, pendingDeleteMempoolTxs []*MempoolTx) (err error) {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact

		// update mempool
		for _, mempoolTx := range pendingUpdateMempoolTxs {
			dbTx := tx.Table(MempoolTableName).Where("id = ?", mempoolTx.ID).
				Select("*").
				Updates(&mempoolTx)
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errors.New("no new mempoolTx")
			}
		}
		for _, pendingDeleteMempoolTx := range pendingDeleteMempoolTxs {
			dbTx := tx.Table(MempoolTableName).Where("id = ?", pendingDeleteMempoolTx.ID).Delete(&pendingDeleteMempoolTx)
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errors.New("delete invalid mempool tx")
			}
		}

		return nil
	})
}
