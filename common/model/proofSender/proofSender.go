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
 */

package proofSender

import (
	"errors"
	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type (
	ProofSenderModel interface {
		CreateProofSenderTable() error
		DropProofSenderTable() error
		CreateProof(row *ProofSender) error
		GetProofsByBlockRange(start int64, end int64, status int, maxBlocksCount int) (proofs []*ProofSender, err error)
		GetProofStartBlockNumber() (num int64, err error)
	}

	defaultProofSenderModel struct {
		table string
		DB    *gorm.DB
	}

	ProofSender struct {
		gorm.Model
		ProofInfo      string
		BlockNumber    int64 `gorm:"index"`
		OnchainOpsRoot []byte
		AccountRoot    []byte
		Commitment     []byte
		Timestamp      int64
		Status         int64
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
		return ErrInvalidProof
	}
	return nil
}

/*
	Func: GetProof
	Params:
	Return: err error
	Description: insert proof and block info in proofSender Table
*/

func (m *defaultProofSenderModel) GetProofsByBlockRange(start int64, end int64, status int, maxBlocksCount int) (proofs []*ProofSender, err error) {

	var blocks []*block.Block

	if end != -1 {
		dbTx := m.DB.Table(block.BlockTableName).Where("block_status = ? AND block_height >= ? AND block_height <= ?", status, start, end).
			Order("block_height").
			Limit(maxBlocksCount).
			Find(&blocks)
		if dbTx.Error != nil {
			logx.Error("[proofSender.GetProofsByBlockRange] %s", dbTx.Error)
			return proofs, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			logx.Errorf("[proofSender.GetProofsByBlockRange] not found")
			return proofs, ErrNotFound
		}
	} else {
		dbTx := m.DB.Table(block.BlockTableName).Where("block_height >= ? AND block_status = ?", start, status).
			Order("block_height").
			Limit(maxBlocksCount).
			Find(&blocks)
		if dbTx.Error != nil {
			logx.Error("[proofSender.GetProofsByBlockRange] %s", dbTx.Error)
			return proofs, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			logx.Errorf("[proofSender.GetProofsByBlockRange] not found")
			return proofs, ErrNotFound
		}
	}
	/*
		needs to check:
		1. if start from "start height"
		2. if is it continuous
		3. if there is corresponding proofs
	*/
	var (
		proofStart  = blocks[0].BlockHeight
		proofEnd    = blocks[len(blocks)-1].BlockHeight
		proofLength = int64(len(blocks))
	)
	logx.Infof("proofStart/start/proofEnd/proofLength : %d/%d/%d/%d", proofStart, start, proofEnd, proofLength)
	if (proofStart == start) && (proofEnd-proofStart+1 == proofLength) {
		dbTx := m.DB.Debug().Table(m.table).Where("block_number >= ? AND block_number <= ? AND status = ?",
			proofStart,
			proofEnd,
			Pending).
			Order("block_number").
			Limit(int(proofLength)).
			Find(&proofs)
		if dbTx.Error != nil {
			logx.Error("[proofSender.GetProofsByBlockRange] %s", dbTx.Error)
			return proofs, dbTx.Error
		} else if dbTx.RowsAffected != proofLength {
			logx.Errorf("[proofSender.GetProofsByBlockRange] proof length cannot correspond to blocks")
			return proofs, errors.New("[proofSender.GetProofsByBlockRange] proof length cannot correspond to blocks")
		}
	}

	return proofs, err
}

/*
	Func: GetStartProofBlockNumber
	Params:
	Return: err error
	Description: get latest proof block number, it used to support prover hub to handle crypto blocks, the result will determine the start range.
*/

func (m *defaultProofSenderModel) GetProofStartBlockNumber() (num int64, err error) {
	var row *ProofSender
	dbTx := m.DB.Table(m.table).Order("block_number desc").Limit(1).Find(&row)
	if dbTx.Error != nil {
		logx.Error("[proofSender.GetProofStartBlockNumber] %s", dbTx.Error)
		return num, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[proofSender.GetProofStartBlockNumber] not found")
		return num, ErrNotFound
	} else {
		return row.BlockNumber, nil
	}
}
