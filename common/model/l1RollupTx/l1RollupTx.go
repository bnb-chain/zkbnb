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

package l1RollupTx

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/proof"
	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	L1RollupTxModel interface {
		CreateL1RollupTxTable() error
		DropL1RollupTxTable() error
		CreateL1RollupTx(tx *L1RollupTx) (bool, error)
		GetLatestHandledTx(txType int64) (tx *L1RollupTx, err error)
		GetLatestPendingTx(txType int64) (tx *L1RollupTx, err error)
		GetL1RollupTxsByStatus(txStatus int) (txs []*L1RollupTx, err error)
		DeleteL1RollupTx(tx *L1RollupTx) error
		UpdateL1RollupTxs(
			pendingUpdateTxs []*L1RollupTx,
			pendingUpdateProofStatus map[int64]int,
		) (err error)
	}

	defaultL1RollupTxModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L1RollupTx struct {
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

func (*L1RollupTx) TableName() string {
	return TableName
}

func NewL1RollupTxModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L1RollupTxModel {
	return &defaultL1RollupTxModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TableName,
		DB:         db,
	}
}

func (m *defaultL1RollupTxModel) CreateL1RollupTxTable() error {
	return m.DB.AutoMigrate(L1RollupTx{})
}

func (m *defaultL1RollupTxModel) DropL1RollupTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultL1RollupTxModel) CreateL1RollupTx(tx *L1RollupTx) (bool, error) {
	dbTx := m.DB.Table(m.table).Create(tx)
	if dbTx.Error != nil {
		logx.Errorf("create l1 rollup tx error, err: %s", dbTx.Error.Error())
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return false, errors.New("invalid rollup tx")
	}
	return true, nil
}

func (m *defaultL1RollupTxModel) GetL1RollupTxsByStatus(txStatus int) (txs []*L1RollupTx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ?", txStatus).Order("l2_block_height, tx_type").Find(&txs)
	if dbTx.Error != nil {
		logx.Errorf("get l1 rollup txs by status error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultL1RollupTxModel) DeleteL1RollupTx(rollupTx *L1RollupTx) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		dbTx := tx.Table(m.table).Where("id = ?", rollupTx.ID).Delete(&rollupTx)
		if dbTx.Error != nil {
			logx.Errorf("delete l1 rollup tx error, err: %s", dbTx.Error.Error())
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return errors.New("delete invalid rollupTx")
		}
		return nil
	})
}

func (m *defaultL1RollupTxModel) UpdateL1RollupTxs(
	pendingUpdateTxs []*L1RollupTx,
	pendingUpdateProofStatus map[int64]int,
) (err error) {
	err = m.DB.Transaction(func(tx *gorm.DB) error {
		for _, pendingUpdateTx := range pendingUpdateTxs {
			dbTx := tx.Table(TableName).Where("id = ?", pendingUpdateTx.ID).
				Select("*").
				Updates(&pendingUpdateTx)
			if dbTx.Error != nil {
				logx.Errorf("update rollup tx error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				txInfo, err := json.Marshal(pendingUpdateTx)
				if err != nil {
					logx.Errorf("marshal rollup tx error, err: %s", err.Error())
					return err
				}
				logx.Errorf("invalid rollup tx:  %s", string(txInfo))
				return errors.New("invalid rollup tx")
			}
		}

		for blockHeight, newStatus := range pendingUpdateProofStatus {
			var row *proof.Proof
			dbTx := tx.Table(proof.TableName).Where("block_number = ?", blockHeight).Find(&row)
			if dbTx.Error != nil {
				logx.Errorf("update proof error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return fmt.Errorf("no such proof. height: %d", blockHeight)
			}
			dbTx = tx.Model(&row).
				Select("status").
				Updates(&proof.Proof{Status: int64(newStatus)})
			if dbTx.Error != nil {
				logx.Errorf("update proof error, err: %s", dbTx.Error.Error())
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

func (m *defaultL1RollupTxModel) GetLatestHandledTx(txType int64) (tx *L1RollupTx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_type = ? AND tx_status = ?", txType, StatusHandled).Order("l2_block_height desc").Find(&tx)
	if dbTx.Error != nil {
		logx.Errorf("unable to get latest handled tx: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultL1RollupTxModel) GetLatestPendingTx(txType int64) (tx *L1RollupTx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_type = ? AND tx_status = ?", txType, StatusPending).Find(&tx)
	if dbTx.Error != nil {
		logx.Errorf("unable to get latest pending tx: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return tx, nil
}
