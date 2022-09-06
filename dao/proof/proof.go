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
 */

package proof

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	TableName = "proof"
)

const (
	NotSent = iota
	NotConfirmed
	Confirmed
)

type (
	ProofModel interface {
		CreateProofTable() error
		DropProofTable() error
		CreateProof(row *Proof) error
		GetProofsBetween(start int64, end int64) (proofs []*Proof, err error)
		GetLatestConfirmedProof() (p *Proof, err error)
		GetProofByBlockNumber(num int64) (p *Proof, err error)
	}

	defaultProofModel struct {
		table string
		DB    *gorm.DB
	}

	Proof struct {
		gorm.Model
		ProofInfo   string
		BlockNumber int64 `gorm:"index:idx_number,unique"`
		Status      int64
	}
)

func (*Proof) TableName() string {
	return TableName
}

func NewProofModel(db *gorm.DB) ProofModel {
	return &defaultProofModel{
		table: TableName,
		DB:    db,
	}
}

func (m *defaultProofModel) CreateProofTable() error {
	return m.DB.AutoMigrate(Proof{})
}

func (m *defaultProofModel) DropProofTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultProofModel) CreateProof(row *Proof) error {
	dbTx := m.DB.Table(m.table).Create(row)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateProof
	}
	return nil
}

func (m *defaultProofModel) GetProofsBetween(start int64, end int64) (proofs []*Proof, err error) {
	dbTx := m.DB.Debug().Table(m.table).Where("block_number >= ? AND block_number <= ? AND status = ?",
		start,
		end,
		NotSent).
		Order("block_number").
		Find(&proofs)

	if dbTx.Error != nil {
		return proofs, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}

	return proofs, err
}

func (m *defaultProofModel) GetLatestConfirmedProof() (p *Proof, err error) {
	var row *Proof
	dbTx := m.DB.Table(m.table).Where("status >= ?", NotConfirmed).Order("block_number desc").Limit(1).Find(&row)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	} else {
		return row, nil
	}
}

func (m *defaultProofModel) GetProofByBlockNumber(num int64) (p *Proof, err error) {
	var row *Proof
	dbTx := m.DB.Table(m.table).Where("block_number = ?", num).Find(&row)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	} else {
		return row, nil
	}
}
