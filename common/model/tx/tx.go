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

package tx

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

var (
	cacheZkbasTxIdPrefix = "cache:zkbas:txVerification:id:"

	cacheZkbasTxTxCountPrefix = "cache:zkbas:txVerification:txCount"
)

type (
	TxModel interface {
		CreateTxTable() error
		DropTxTable() error
		GetTxsTotalCount() (count int64, err error)
		GetTxsList(limit int64, offset int64) (txList []*Tx, err error)
		GetTxByTxHash(txHash string) (tx *Tx, err error)
		GetTxByTxId(id int64) (tx *Tx, err error)
		GetTxsTotalCountBetween(from, to time.Time) (count int64, err error)
		GetDistinctAccountCountBetween(from, to time.Time) (count int64, err error)
	}

	defaultTxModel struct {
		sqlc.CachedConn
		table     string
		DB        *gorm.DB
		RedisConn *redis.Redis
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

func NewTxModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB, redisConn *redis.Redis) TxModel {
	return &defaultTxModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TxTableName,
		DB:         db,
		RedisConn:  redisConn,
	}
}

func (*Tx) TableName() string {
	return TxTableName
}

/*
	Func: CreateTxTable
	Params:
	Return: err error
	Description: create txVerification table
*/
func (m *defaultTxModel) CreateTxTable() error {
	return m.DB.AutoMigrate(Tx{})
}

/*
	Func: DropTxTable
	Params:
	Return: err error
	Description: drop txVerification table
*/
func (m *defaultTxModel) DropTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetTxsTotalCount
	Params:
	Return: count int64, err error
	Description: used for counting total transactions for explorer dashboard
*/
func (m *defaultTxModel) GetTxsTotalCount() (count int64, err error) {

	key := fmt.Sprintf("%s", cacheZkbasTxTxCountPrefix)
	val, err := m.RedisConn.Get(key)
	if err != nil {
		logx.Errorf("get redis error: %s, key:%s", err.Error(), key)
		return 0, err

	} else if val == "" {
		dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
		if dbTx.Error != nil {
			if dbTx.Error == errorcode.DbErrNotFound {
				return 0, nil
			}
			logx.Errorf("get Tx count error, err: %s", dbTx.Error.Error())
			return 0, dbTx.Error
		}

		err = m.RedisConn.Setex(key, strconv.FormatInt(count, 10), 120)
		if err != nil {
			logx.Errorf("redis set error: %s", err.Error())
			return 0, err
		}
	} else {
		count, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			logx.Errorf("strconv.ParseInt error: %s, value : %s", err.Error(), val)
			return 0, err
		}
	}

	return count, nil
}

/*
	Func: GetTxsList
	Params:
	Return: list of txs, err error
	Description: used for showing transactions for explorer
*/

func (m *defaultTxModel) GetTxsList(limit int64, offset int64) (txList []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("created_at desc").Find(&txList)
	if dbTx.Error != nil {
		logx.Errorf("fail to get txs offset: %d, limit: %d, error: %s", offset, limit, dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return txList, nil
}

/*
	Func: GetTxByTxHash
	Params: txHash string
	Return: txVerification Tx, err error
	Description: used for /api/v1/txVerification/getTxByHash
*/
func (m *defaultTxModel) GetTxByTxHash(txHash string) (tx *Tx, err error) {
	var txForeignKeyColumn = `TxDetails`

	dbTx := m.DB.Table(m.table).Where("tx_hash = ?", txHash).Find(&tx)
	if dbTx.Error != nil {
		logx.Errorf("get tx by hash error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
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

func (m *defaultTxModel) GetTxByTxId(id int64) (tx *Tx, err error) {
	var txForeignKeyColumn = `TxDetails`

	dbTx := m.DB.Table(m.table).Where("id = ?", id).Find(&tx)
	if dbTx.Error != nil {
		logx.Errorf("get tx by id error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
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
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultTxModel) GetDistinctAccountCountBetween(from, to time.Time) (count int64, err error) {
	dbTx := m.DB.Raw("SELECT account_index FROM tx WHERE created_at BETWEEN ? AND ? AND account_index != -1 GROUP BY account_index", from, to).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("fail to get dau by time range: %d-%d, error: %s", from.Unix(), to.Unix(), dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}
