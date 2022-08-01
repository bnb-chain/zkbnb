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
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

var (
	cacheZkbasTxIdPrefix      = "cache:zkbas:txVerification:id:"
	cacheZkbasTxTxHashPrefix  = "cache:zkbas:txVerification:txHash:"
	cacheZkbasTxTxCountPrefix = "cache:zkbas:txVerification:txCount"
)

type (
	TxModel interface {
		CreateTxTable() error
		DropTxTable() error
		GetTxsListByBlockHeight(blockHeight int64, limit int, offset int) (txs []*Tx, err error)
		GetTxsListByAccountIndex(accountIndex int64, limit int, offset int) (txs []*Tx, err error)
		GetTxsListByAccountIndexAndTxType(accountIndex int64, txType uint8, limit int, offset int) (txs []*Tx, err error)
		GetTxsListByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8, limit int, offset int) (txs []*Tx, err error)
		GetTxsListByAccountName(accountName string, limit int, offset int) (txs []*Tx, err error)
		GetTxsTotalCount() (count int64, err error)
		GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
		GetTxsTotalCountByAccountIndexAndTxType(accountIndex int64, txType uint8) (count int64, err error)
		GetTxsTotalCountByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8) (count int64, err error)
		GetTxsTotalCountByBlockHeight(blockHeight int64) (count int64, err error)
		GetTxByTxHash(txHash string) (tx *Tx, err error)
		GetTxByTxId(id uint) (tx *Tx, err error)
		GetTxsListGreaterThanBlockHeight(blockHeight int64) (txs []*Tx, err error)
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
	Func: GetTxsListByBlockHeight
	Params: blockHeight int64, limit int64, offset int64
	Return: txVerification []*Tx, err error
	Description: used for getTxsListByBlockHeight API
*/

func (m *defaultTxModel) GetTxsListByBlockHeight(blockHeight int64, limit int, offset int) (txs []*Tx, err error) {
	var txForeignKeyColumn = `TxDetails`
	// todo cache optimize
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Order("created_at desc, id desc").Offset(offset).Limit(limit).Find(&txs)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByBlockHeight] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[txVerification.GetTxsListByBlockHeight] Get Txs Error")
		return nil, errorcode.DbErrNotFound
	}

	for _, tx := range txs {
		key := fmt.Sprintf("%s%v", cacheZkbasTxIdPrefix, tx.ID)
		val, err := m.RedisConn.Get(key)
		if err != nil {
			errInfo := fmt.Sprintf("[txVerification.GetTxsListByBlockHeight] Get Redis Error: %s, key:%s", err.Error(), key)
			logx.Errorf(errInfo)
			return nil, err
		} else if val == "" {
			err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
			if err != nil {
				logx.Error("[txVerification.GetTxsListByBlockHeight] Get Associate TxDetails Error")
				return nil, err
			}

			// json string
			jsonString, err := json.Marshal(tx.TxDetails)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByBlockHeight] json.Marshal Error: %s, value: %v", err.Error(), tx.TxDetails)
				return nil, err
			}
			// todo
			err = m.RedisConn.Setex(key, string(jsonString), 60*10+rand.Intn(60*3))
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByBlockHeight] redis set error: %s", err.Error())
				return nil, err
			}
		} else {
			// json string unmarshal
			var (
				nTxDetails []*TxDetail
			)
			err = json.Unmarshal([]byte(val), &nTxDetails)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByBlockHeight] json.Unmarshal error: %s, value : %s", err.Error(), val)
				return nil, err
			}
			tx.TxDetails = nTxDetails
		}

	}
	return txs, nil
}

/*
	Func: GetTxsListByAccountIndex
	Params: accountIndex int64, limit int64, offset int64
	Return: txVerification []*Tx, err error
	Description: used for getTxsListByAccountIndex API, return all txVerification related to accountIndex.
				Because there are many accountIndex in
				 sorted by created_time
				 Associate With TxDetail Table
*/

func (m *defaultTxModel) GetTxsListByAccountIndex(accountIndex int64, limit int, offset int) (txs []*Tx, err error) {
	var (
		txDetailTable      = `tx_detail`
		txIds              []int64
		txForeignKeyColumn = `TxDetails`
	)
	dbTx := m.DB.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountIndex] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[info.GetTxsListByAccountIndex] Get TxIds Error")
		return nil, errorcode.DbErrNotFound
	}
	dbTx = m.DB.Table(m.table).Order("created_at desc, id desc").Offset(offset).Limit(limit).Find(&txs, txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountIndex] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[txVerification.GetTxsListByAccountIndex] Get Txs Error")
		return nil, errorcode.DbErrNotFound
	}
	//TODO: cache operation
	for _, tx := range txs {
		key := fmt.Sprintf("%s%v", cacheZkbasTxIdPrefix, tx.ID)
		val, err := m.RedisConn.Get(key)
		if err != nil {
			errInfo := fmt.Sprintf("[txVerification.GetTxsListByAccountIndex] Get Redis Error: %s, key:%s", err.Error(), key)
			logx.Errorf(errInfo)
			return nil, err

		} else if val == "" {
			err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
			if err != nil {
				logx.Error("[txVerification.GetTxsListByAccountIndex] Get Associate TxDetails Error")
				return nil, err
			}

			// json string
			jsonString, err := json.Marshal(tx.TxDetails)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByAccountIndex] json.Marshal Error: %s, value: %v", err.Error(), tx.TxDetails)
				return nil, err
			}
			// todo
			err = m.RedisConn.Setex(key, string(jsonString), 60*10+rand.Intn(60*3))
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByAccountIndex] redis set error: %s", err.Error())
				return nil, err
			}
		} else {
			// json string unmarshal
			var (
				nTxDetails []*TxDetail
			)
			err = json.Unmarshal([]byte(val), &nTxDetails)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByAccountIndex] json.Unmarshal error: %s, value : %s", err.Error(), val)
				return nil, err
			}
			tx.TxDetails = nTxDetails
		}
	}
	return txs, nil
}

/*
	Func: GetTxsListByAccountIndexAndTxType
	Params: accountIndex int64, txType uint8,limit int, offset int
	Return: txVerification []*Tx, err error
	Description: used for getTxsListByAccountIndex API, return all txVerification related to accountIndex and txType.
				Because there are many accountIndex in
				 sorted by created_time
				 Associate With TxDetail Table
*/

func (m *defaultTxModel) GetTxsListByAccountIndexAndTxType(accountIndex int64, txType uint8, limit int, offset int) (txs []*Tx, err error) {
	var (
		txDetailTable      = `tx_detail`
		txIds              []int64
		txForeignKeyColumn = `TxDetails`
	)
	dbTx := m.DB.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[info.GetTxsListByAccountIndexAndTxType] Get TxIds Error")
		return nil, errorcode.DbErrNotFound
	}
	dbTx = m.DB.Table(m.table).Order("created_at desc").Where("tx_type = ?", txType).Offset(offset).Limit(limit).Find(&txs, txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[txVerification.GetTxsListByAccountIndexAndTxType] Get Txs Error")
		return nil, errorcode.DbErrNotFound
	}
	//TODO: cache operation
	for _, tx := range txs {
		key := fmt.Sprintf("%s%v:txType:%v", cacheZkbasTxIdPrefix, tx.ID, txType)
		val, err := m.RedisConn.Get(key)
		if err != nil {
			errInfo := fmt.Sprintf("[txVerification.GetTxsListByAccountIndexAndTxType] Get Redis Error: %s, key:%s", err.Error(), key)
			logx.Errorf(errInfo)
			return nil, err

		} else if val == "" {
			err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
			if err != nil {
				logx.Error("[txVerification.GetTxsListByAccountIndexAndTxType] Get Associate TxDetails Error")
				return nil, err
			}

			// json string
			jsonString, err := json.Marshal(tx.TxDetails)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByAccountIndexAndTxType] json.Marshal Error: %s, value: %v", err.Error(), tx.TxDetails)
				return nil, err
			}
			// todo
			err = m.RedisConn.Setex(key, string(jsonString), 30)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByAccountIndexAndTxType] redis set error: %s", err.Error())
				return nil, err
			}
		} else {
			// json string unmarshal
			var (
				nTxDetails []*TxDetail
			)
			err = json.Unmarshal([]byte(val), &nTxDetails)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByAccountIndexAndTxType] json.Unmarshal error: %s, value : %s", err.Error(), val)
				return nil, err
			}
			tx.TxDetails = nTxDetails
		}
	}
	return txs, nil
}

/*
	Func: GetTxsListByAccountIndexAndTxTypeArray
	Params: accountIndex int64, txTypeArray []uint8, limit int, offset int
	Return: txVerification []*Tx, err error
	Description: used for getTxsListByAccountIndex API, return all txVerification related to accountIndex and txTypeArray.
				Because there are many accountIndex in
				 sorted by created_time
				 Associate With TxDetail Table
*/

func (m *defaultTxModel) GetTxsListByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8, limit int, offset int) (txs []*Tx, err error) {
	var (
		txDetailTable      = `tx_detail`
		txIds              []int64
		txForeignKeyColumn = `TxDetails`
	)
	dbTx := m.DB.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[info.GetTxsListByAccountIndexAndTxTypeArray] Get TxIds Error")
		return nil, errorcode.DbErrNotFound
	}
	dbTx = m.DB.Table(m.table).Order("created_at desc").Where("tx_type in (?)", txTypeArray).Offset(offset).Limit(limit).Find(&txs, txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[GetTxsListByAccountIndexAndTxTypeArray] Get Txs Error")
		return nil, errorcode.DbErrNotFound
	}
	//TODO: cache operation
	for _, tx := range txs {
		key := fmt.Sprintf("%s%v:txTypeArray:%s", cacheZkbasTxIdPrefix, tx.ID, txTypeArray)
		val, err := m.RedisConn.Get(key)
		if err != nil {
			errInfo := fmt.Sprintf("[txVerification.GetTxsListByAccountIndexAndTxTypeArray] Get Redis Error: %s, key:%s", err.Error(), key)
			logx.Errorf(errInfo)
			return nil, err

		} else if val == "" {
			err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
			if err != nil {
				logx.Error("[txVerification.GetTxsListByAccountIndexAndTxTypeArray] Get Associate TxDetails Error")
				return nil, err
			}

			// json string
			jsonString, err := json.Marshal(tx.TxDetails)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByAccountIndexAndTxTypeArray] json.Marshal Error: %s, value: %v", tx.TxDetails)
				return nil, err
			}
			// todo
			err = m.RedisConn.Setex(key, string(jsonString), 30)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByAccountIndexAndTxTypeArray] redis set error: %s", err.Error())
				return nil, err
			}
		} else {
			// json string unmarshal
			var (
				nTxDetails []*TxDetail
			)
			err = json.Unmarshal([]byte(val), &nTxDetails)
			if err != nil {
				logx.Errorf("[txVerification.GetTxsListByAccountIndexAndTxTypeArray] json.Unmarshal error: %s, value : %s", err.Error(), val)
				return nil, err
			}
			tx.TxDetails = nTxDetails
		}
	}
	return txs, nil
}

/*
	Func: GetTxsListByAccountName
	Params: accountName string, limit int64, offset int64
	Return: txVerification []*Tx, err error
	Description: used for getTxsListByAccountName API
				 sorted by created_time
				 Associate With TxDetail Table
*/
func (m *defaultTxModel) GetTxsListByAccountName(accountName string, limit int, offset int) (txs []*Tx, err error) {
	var (
		txDetailTable      = `tx_detail`
		txIds              []int64
		txForeignKeyColumn = `TxDetails`
	)
	dbTx := m.DB.Table(txDetailTable).Select("tx_id").Where("account_name = ? and deleted_at is NULL", accountName).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountName] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[txVerification.GetTxsListByAccountName] Get TxIds Error")
		return nil, errorcode.DbErrNotFound
	}
	dbTx = m.DB.Table(m.table).Order("created_at desc, id desc").Offset(offset).Limit(limit).Find(&txs, txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListByAccountName] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[txVerification.GetTxsListByAccountName] Get Txs Error")
		return nil, errorcode.DbErrNotFound
	}
	//TODO: cache operation
	for _, tx := range txs {
		err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
		if err != nil {
			logx.Error("[txVerification.GetTxsListByAccountName] Get Associate TxDetails Error")
			return nil, err
		}
	}
	return txs, nil
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
		errInfo := fmt.Sprintf("[txVerification.GetTxsTotalCount] Get Redis Error: %s, key:%s", err.Error(), key)
		logx.Errorf(errInfo)
		return 0, err

	} else if val == "" {
		dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
		if dbTx.Error != nil {
			if dbTx.Error == errorcode.DbErrNotFound {
				return 0, nil
			}
			logx.Error("[txVerification.GetTxsTotalCount] Get Tx Count Error")
			return 0, err
		}

		err = m.RedisConn.Setex(key, strconv.FormatInt(count, 10), 120)
		if err != nil {
			logx.Errorf("[txVerification.GetTxsTotalCount] redis set error: %s", err.Error())
			return 0, err
		}
	} else {
		count, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			logx.Errorf("[txVerification.GetTxsListByAccountIndex] strconv.ParseInt error: %s, value : %s", err.Error(), val)
			return 0, err
		}
	}

	return count, nil
}

/*
	Func: GetTxsTotalCount
	Params: accountIndex int64
	Return: count int64, err error
	Description: used for counting total transactions for explorer dashboard
*/
func (m *defaultTxModel) GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error) {
	var (
		txDetailTable = `tx_detail`
	)
	dbTx := m.DB.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Count(&count)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsTotalCountByAccountIndex] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[txVerification.GetTxsTotalCountByAccountIndex] No Txs of account index %d in Tx Table", accountIndex)
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetTxsTotalCountByAccountIndexAndTxType
	Params: accountIndex int64, txType uint8
	Return: count int64, err error
	Description: used for counting total transactions for explorer dashboard
*/
func (m *defaultTxModel) GetTxsTotalCountByAccountIndexAndTxType(accountIndex int64, txType uint8) (count int64, err error) {
	var (
		txDetailTable = `tx_detail`
		txIds         []int64
	)
	dbTx := m.DB.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsTotalCountByAccountIndexAndTxType] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[txVerification.GetTxsTotalCountByAccountIndexAndTxType] No Txs of account index %d  and txVerification type %d in Tx Table", accountIndex, txType)
		return 0, nil
	}
	dbTx = m.DB.Table(m.table).Where("id in (?) and deleted_at is NULL and tx_type = ?", txIds, txType).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[txVerification.GetTxsTotalCountByAccountIndexAndTxTypee] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[txVerification.GetTxsTotalCountByAccountIndexAndTxType] no txVerification of account index %d and txVerification type = %d in mempool", accountIndex, txType)
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetTxsTotalCountByAccountIndexAndTxTypeArray
	Params: accountIndex int64, txTypeArray []uint8
	Return: count int64, err error
	Description: used for counting total transactions for explorer dashboard
*/
func (m *defaultTxModel) GetTxsTotalCountByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8) (count int64, err error) {
	var (
		txDetailTable = `tx_detail`
		txIds         []int64
	)
	dbTx := m.DB.Table(txDetailTable).Select("tx_id").Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		logx.Errorf("[txVerification.GetTxsTotalCountByAccountIndexAndTxTypeArray] %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[txVerification.GetTxsTotalCountByAccountIndexAndTxTypeArray] No Txs of account index %d  and txVerification type %v in Tx Table", accountIndex, txTypeArray)
		return 0, nil
	}
	dbTx = m.DB.Table(m.table).Where("id in (?) and deleted_at is NULL and tx_type in (?)", txIds, txTypeArray).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[txVerification.GetTxsTotalCountByAccountIndexAndTxTypeArray] %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[txVerification.GetTxsTotalCountByAccountIndexAndTxTypeArray] no txVerification of account index %d and txVerification type = %v in mempool", accountIndex, txTypeArray)
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetTxsTotalCountByBlockHeight
	Params: blockHeight int64
	Return: count int64, err error
	Description: used for counting total transactions for explorer dashboard
*/
func (m *defaultTxModel) GetTxsTotalCountByBlockHeight(blockHeight int64) (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height = ? and deleted_at is NULL", blockHeight).Count(&count)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsTotalCountByBlockHeight] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[txVerification.GetTxsTotalCountByBlockHeight] No Txs of block height %d in Tx Table", blockHeight)
		return 0, nil
	}
	return count, nil
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
		logx.Error("[txVerification.GetTxByTxHash] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[txVerification.GetTxByTxHash] No such Tx with txHash: %s", txHash)
		return nil, errorcode.DbErrNotFound
	}
	err = m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
	if err != nil {
		logx.Error("[txVerification.GetTxByTxHash] Get Associate TxDetails Error")
		return nil, err
	}
	// re-order tx details
	sort.SliceStable(tx.TxDetails, func(i, j int) bool {
		return tx.TxDetails[i].Order < tx.TxDetails[j].Order
	})

	return tx, nil
}

func (m *defaultTxModel) GetTxByTxId(id uint) (tx *Tx, err error) {
	var txForeignKeyColumn = `TxDetails`

	dbTx := m.DB.Table(m.table).Where("id = ?", id).Find(&tx)
	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxByTxId] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[txVerification.GetTxByTxId] No such Tx with tx id: %s", id)
		return nil, errorcode.DbErrNotFound
	}
	err = m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
	if err != nil {
		logx.Error("[txVerification.GetTxByTxId] Get Associate TxDetails Error")
		return nil, err
	}
	// re-order tx details
	sort.SliceStable(tx.TxDetails, func(i, j int) bool {
		return tx.TxDetails[i].Order < tx.TxDetails[j].Order
	})

	return tx, nil
}

/*
	Func: GetTxsListGreaterThanBlockHeight
	Params: blockHeight int64
	Return: txVerification []*Tx, err error
	Description: used for info service
*/

func (m *defaultTxModel) GetTxsListGreaterThanBlockHeight(blockHeight int64) (txs []*Tx, err error) {
	var (
		txForeignKeyColumn = `TxDetails`
	)

	dbTx := m.DB.Table(m.table).Where("block_height >= ? and block_height < ?", blockHeight, blockHeight+maxBlocks).Order("created_at, id").Find(&txs)

	if dbTx.Error != nil {
		logx.Error("[txVerification.GetTxsListGreaterThanBlockHeight] %s", dbTx.Error)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[txVerification.GetTxsListGreaterThanBlockHeight] No txVerification blockHeight greater than %d", blockHeight)
		return nil, nil
	}

	for _, tx := range txs {
		err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
		if err != nil {
			logx.Error("[txVerification.GetTxsListGreaterThanBlockHeight] Get Associate TxDetails Error")
			return nil, err
		}
	}
	return txs, nil
}
