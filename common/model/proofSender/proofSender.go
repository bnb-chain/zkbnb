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
 */

package proofSender

import (
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	ProofSenderModel interface {
		CreateProofSenderTable() error
		DropProofSenderTable() error
		CreateProof(row *ProofSender) error
		GetProofsByBlockRange(start int64, end int64, maxProofsCount int) (proofs []*ProofSender, err error)
		GetProofStartBlockNumber() (num int64, err error)
		GetLatestConfirmedProof() (p *ProofSender, err error)
		GetProofByBlockNumber(num int64) (p *ProofSender, err error)
	}

	defaultProofSenderModel struct {
		table string
		DB    *gorm.DB
	}

	ProofSender struct {
		gorm.Model
		ProofInfo   string
		BlockNumber int64 `gorm:"index:idx_number,unique"`
		Status      int64
	}
)

func (*ProofSender) TableName() string {
	return TableName
}

func NewProofSenderModel(db *gorm.DB) ProofSenderModel {
	return &defaultProofSenderModel{
		table: TableName,
		DB:    db,
	}
}

/*
	Func: CreateProofSenderTable
	Params:
	Return: err error
	Description: create proofSender table
*/

func (m *defaultProofSenderModel) CreateProofSenderTable() error {
	return m.DB.AutoMigrate(ProofSender{})
}

/*
	Func: DropProofSenderTable
	Params:
	Return: err error
	Description: drop proofSender table
*/

func (m *defaultProofSenderModel) DropProofSenderTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: InsertProof
	Params:
	Return: err error
	Description: insert proof and block info in proofSender Table
*/

func (m *defaultProofSenderModel) CreateProof(row *ProofSender) error {
	dbTx := m.DB.Table(m.table).Create(row)
	if dbTx.Error != nil {
		logx.Errorf("[proofSender.CreateProof] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Errorf("[proofSender.CreateProof] Create Invalid Proof")
		return errorcode.DbErrFailToCreateProof
	}
	return nil
}

/*
	Func: GetProof
	Params:
	Return: err error
	Description: getProofsByBlockRange
*/

func (m *defaultProofSenderModel) GetProofsByBlockRange(start int64, end int64, maxProofsCount int) (proofs []*ProofSender, err error) {

	dbTx := m.DB.Debug().Table(m.table).Where("block_number >= ? AND block_number <= ? AND status = ?",
		start,
		end,
		NotSent).
		Order("block_number").
		Limit(maxProofsCount).
		Find(&proofs)

	if dbTx.Error != nil {
		logx.Error("[proofSender.GetProofsByBlockRange] %s", dbTx.Error)
		return proofs, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[proofSender.GetProofsByBlockRange] error not found")
		return proofs, errorcode.DbErrNotFound
	}

	return proofs, err
}

/*
	Func: GetStartProofBlockNumber
	Params:
	Return: err error
	Description: Get the latest proof block number. It is used to support the prover hub to handle crypto blocks; the result will determine the start range.
*/

func (m *defaultProofSenderModel) GetProofStartBlockNumber() (num int64, err error) {
	var row *ProofSender
	dbTx := m.DB.Table(m.table).Order("block_number desc").Limit(1).Find(&row)
	if dbTx.Error != nil {
		logx.Error("[proofSender.GetProofStartBlockNumber] %s", dbTx.Error)
		return num, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[proofSender.GetProofStartBlockNumber] not found")
		return num, errorcode.DbErrNotFound
	} else {
		return row.BlockNumber, nil
	}
}

/*
	Func: GetLatestSentProof
	Params:
	Return: p *ProofSender, err error
	Description: get latest sent proof block number,
		it used to support prover hub to init merkle tree.
*/
func (m *defaultProofSenderModel) GetLatestConfirmedProof() (p *ProofSender, err error) {
	var row *ProofSender
	dbTx := m.DB.Table(m.table).Where("status >= ?", NotConfirmed).Order("block_number desc").Limit(1).Find(&row)
	if dbTx.Error != nil {
		logx.Errorf("[proofSender.GetLatestSentProof] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[proofSender.GetLatestSentProof] not found")
		return nil, errorcode.DbErrNotFound
	} else {
		return row, nil
	}
}

/*
	Func: GetProofByBlockNumber
	Params:
	Return: p *ProofSender, err error
	Description: get certain blockNumber proof
		it used to support prover hub to init unproved block.
*/
func (m *defaultProofSenderModel) GetProofByBlockNumber(num int64) (p *ProofSender, err error) {
	var row *ProofSender
	dbTx := m.DB.Table(m.table).Where("block_number = ?", num).Find(&row)
	if dbTx.Error != nil {
		logx.Errorf("[proofSender.GetProofByBlockNumber] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[proofSender.GetProofByBlockNumber] not found")
		return nil, errorcode.DbErrNotFound
	} else {
		return row, nil
	}
}
