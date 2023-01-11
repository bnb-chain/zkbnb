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

package rollback

import (
	"github.com/bnb-chain/zkbnb/types"
	"gorm.io/gorm"
)

const (
	RollbackTableName = `rollback`
)

type (
	RollbackModel interface {
		CreateRollbackTable() error
		DropRollbackTable() error
		Get(height int64, limit int, offset int64) (rollbacks []*Rollback, err error)
		GetCount(height int64, limit int, offset int64) (count int64, err error)
		CreateInTransact(tx *gorm.DB, rollback *Rollback) (err error)
	}

	defaultRollbackModel struct {
		table string
		DB    *gorm.DB
	}

	Rollback struct {
		gorm.Model
		FromBlockHeight int64 `gorm:"index"`
		FromPoolTxId    uint  `gorm:"index"`
		FromTxHash      string
		PoolTxIds       string
		BlockHeights    string
		AccountIndexes  string
		NftIndexes      string
	}
)

func NewRollbackModel(db *gorm.DB) RollbackModel {
	return &defaultRollbackModel{
		table: RollbackTableName,
		DB:    db,
	}
}

func (*Rollback) TableName() string {
	return RollbackTableName
}

func (m *defaultRollbackModel) CreateRollbackTable() error {
	return m.DB.AutoMigrate(Rollback{})
}

func (m *defaultRollbackModel) DropRollbackTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultRollbackModel) Get(height int64, limit int, offset int64) (rollbacks []*Rollback, err error) {
	dbTx := m.DB.Table(m.table).Select("from_block_height,from_tx_hash,id,created_at").Where("from_block_height >= ? and from_pool_tx_id!=0", height).Limit(limit).Offset(int(offset)).Order("id asc").Find(&rollbacks)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return rollbacks, nil
}
func (m *defaultRollbackModel) GetCount(height int64, limit int, offset int64) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("from_block_height >= ? and from_pool_tx_id!=0", height).Count(&count)
	if dbTx.Error != nil {
		if dbTx.Error == types.DbErrNotFound {
			return 0, nil
		}
		return 0, types.DbErrSqlOperation
	}
	return count, nil
}
func (m *defaultRollbackModel) CreateInTransact(tx *gorm.DB, rollback *Rollback) (err error) {
	dbTx := tx.Table(m.table).Create(rollback)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreateRollback
	}
	return nil
}
