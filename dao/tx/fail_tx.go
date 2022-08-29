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
	"errors"

	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	FailTxTableName = `fail_tx`
)

type (
	FailTxModel interface {
		CreateFailTxTable() error
		DropFailTxTable() error
		CreateFailTx(tx *Tx) error
	}

	defaultFailTxModel struct {
		table string
		DB    *gorm.DB
	}
)

func NewFailTxModel(db *gorm.DB) FailTxModel {
	return &defaultFailTxModel{
		table: FailTxTableName,
		DB:    db,
	}
}

func (m *defaultFailTxModel) CreateFailTxTable() error {
	return m.DB.AutoMigrate(Tx{})
}

func (m *defaultFailTxModel) DropFailTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultFailTxModel) CreateFailTx(tx *Tx) error {
	if tx.TxStatus != StatusFailed {
		return errors.New("tx status is not failed")
	}

	dbTx := m.DB.Table(m.table).Create(tx)
	if dbTx.Error != nil {
		return types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateFailTx
	}
	return nil
}
