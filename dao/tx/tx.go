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

package tx

import (
	"sort"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/types"
)

type (
	TxModel interface {
		CreateTxTable() error
		DropTxTable() error
		GetTxsTotalCount() (count int64, err error)
		GetTxsList(limit int64, offset int64) (txList []*Tx, err error)
		GetTxsListByAccountIndex(accountIndex int64, limit int64, offset int64) (txList []*Tx, err error)
		GetTxsCountByAccountIndex(accountIndex int64) (count int64, err error)
		GetTxsListByAccountIndexTxType(accountIndex int64, txType int64, limit int64, offset int64) (txList []*Tx, err error)
		GetTxsCountByAccountIndexTxType(accountIndex int64, txType int64) (count int64, err error)
		GetTxByHash(txHash string) (tx *Tx, err error)
		GetTxById(id int64) (tx *Tx, err error)
		GetTxsTotalCountBetween(from, to time.Time) (count int64, err error)
		GetDistinctAccountsCountBetween(from, to time.Time) (count int64, err error)
	}

	defaultTxModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	Tx struct {
		gorm.Model
		TxHash        string `gorm:"uniqueIndex"`
		TxType        int64
		GasFee        string
		GasFeeAssetId int64
		TxStatus      int64
		BlockHeight   int64 `gorm:"index"`
		BlockId       int64 `gorm:"index"`
		StateRoot     string
		NftIndex      int64
		PairIndex     int64
		AssetId       int64
		TxAmount      string
		NativeAddress string
		TxInfo        string
		TxDetails     []*TxDetail `gorm:"foreignKey:TxId"`
		ExtraInfo     string
		Memo          string
		AccountIndex  int64
		Nonce         int64
		ExpiredAt     int64
		TxIndex       int64
	}
)

func NewTxModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) TxModel {
	return &defaultTxModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TxTableName,
		DB:         db,
	}
}

func (*Tx) TableName() string {
	return TxTableName
}

func (m *defaultTxModel) CreateTxTable() error {
	return m.DB.AutoMigrate(Tx{})
}

func (m *defaultTxModel) DropTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultTxModel) GetTxsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		if dbTx.Error == types.DbErrNotFound {
			return 0, nil
		}
		logx.Errorf("get tx total count error, err: %s", dbTx.Error.Error())
		return 0, types.DbErrSqlOperation
	}
	return count, nil
}

func (m *defaultTxModel) GetTxsList(limit int64, offset int64) (txList []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("created_at desc").Find(&txList)
	if dbTx.Error != nil {
		logx.Errorf("fail to get txs, offset: %d, limit: %d, error: %s", offset, limit, dbTx.Error.Error())
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txList, nil
}

func (m *defaultTxModel) GetTxsListByAccountIndex(accountIndex int64, limit int64, offset int64) (txList []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Limit(int(limit)).Offset(int(offset)).Order("created_at desc").Find(&txList)
	if dbTx.Error != nil {
		logx.Errorf("fail to get txs by account index: %d, offset: %d, limit: %d, error: %s", accountIndex, offset, limit, dbTx.Error.Error())
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txList, nil
}

func (m *defaultTxModel) GetTxsCountByAccountIndex(accountIndex int64) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("fail to get tx count by account: %d, error: %s", accountIndex, dbTx.Error.Error())
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultTxModel) GetTxsListByAccountIndexTxType(accountIndex int64, txType int64, limit int64, offset int64) (txList []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and tx_type = ?", accountIndex, txType).Limit(int(limit)).Offset(int(offset)).Order("created_at desc").Find(&txList)
	if dbTx.Error != nil {
		logx.Errorf("fail to get txs by account index: %d, tx type:%d, offset: %d, limit: %d, error: %s", accountIndex, txType, offset, limit, dbTx.Error.Error())
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txList, nil
}

func (m *defaultTxModel) GetTxsCountByAccountIndexTxType(accountIndex int64, txType int64) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and tx_type = ?", accountIndex, txType).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("fail to get tx count by account: %d, tx type: %d, error: %s", accountIndex, txType, dbTx.Error.Error())
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultTxModel) GetTxByHash(txHash string) (tx *Tx, err error) {
	var txForeignKeyColumn = `TxDetails`

	dbTx := m.DB.Table(m.table).Where("tx_hash = ?", txHash).Find(&tx)
	if dbTx.Error != nil {
		logx.Errorf("get tx by hash error, err: %s", dbTx.Error.Error())
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	err = m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
	if err != nil {
		logx.Errorf("get associate tx details error, err: %s", err.Error())
		return nil, err
	}
	// re-order tx details
	sort.SliceStable(tx.TxDetails, func(i, j int) bool {
		return tx.TxDetails[i].Order < tx.TxDetails[j].Order
	})

	return tx, nil
}

func (m *defaultTxModel) GetTxById(id int64) (tx *Tx, err error) {
	var txForeignKeyColumn = `TxDetails`

	dbTx := m.DB.Table(m.table).Where("id = ?", id).Find(&tx)
	if dbTx.Error != nil {
		logx.Errorf("get tx by id error, err: %s", dbTx.Error.Error())
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	err = m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
	if err != nil {
		logx.Errorf("get associate tx details error, err: %s", err.Error())
		return nil, err
	}
	// re-order tx details
	sort.SliceStable(tx.TxDetails, func(i, j int) bool {
		return tx.TxDetails[i].Order < tx.TxDetails[j].Order
	})

	return tx, nil
}

func (m *defaultTxModel) GetTxsTotalCountBetween(from, to time.Time) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("created_at BETWEEN ? AND ?", from, to).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("fail to get tx by time range: %d-%d, error: %s", from.Unix(), to.Unix(), dbTx.Error.Error())
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultTxModel) GetDistinctAccountsCountBetween(from, to time.Time) (count int64, err error) {
	dbTx := m.DB.Raw("SELECT account_index FROM tx WHERE created_at BETWEEN ? AND ? AND account_index != -1 GROUP BY account_index", from, to).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("fail to get dau by time range: %d-%d, error: %s", from.Unix(), to.Unix(), dbTx.Error.Error())
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}
