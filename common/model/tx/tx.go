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
 *
 */

package tx

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheZecreyTxIdPrefix      = "cache:zecrey:tx:id:"
	cacheZecreyTxTxHashPrefix  = "cache:zecrey:tx:txHash:"
	cacheZecreyTxTxCountPrefix = "cache:zecrey:tx:txCount"
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
		GasFee        int64
		GasFeeAssetId int64
		TxStatus      int64
		BlockHeight   int64 `gorm:"index"`
		BlockId       int64 `gorm:"index"`
		AccountRoot   string
		AssetAId      int64
		AssetBId      int64
		TxAmount      int64
		NativeAddress string
		TxInfo        string
		TxDetails     []*TxDetail `gorm:"foreignkey:TxId"`
		ExtraInfo     string
		Memo          string
		// block detail pk
		BlockDetailPk uint
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
	Description: create tx table
*/
func (m *defaultTxModel) CreateTxTable() error {
	return m.DB.AutoMigrate(Tx{})
}

/*
	Func: DropTxTable
	Params:
	Return: err error
	Description: drop tx table
*/
func (m *defaultTxModel) DropTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetTxsListByBlockHeight
	Params: blockHeight int64, limit int64, offset int64
	Return: txs []*Tx, err error
	Description: used for getTxsListByBlockHeight API
*/

func (m *defaultTxModel) GetTxsListByBlockHeight(blockHeight int64, limit int, offset int) (txs []*Tx, err error) {
	var txForeignKeyColumn = `TxDetails`
	// todo cache optimize
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Order("created_at desc, id desc").Offset(offset).Limit(limit).Find(&txs)
	if dbTx.Error != nil {
		logx.Error("[tx.GetTxsListByBlockHeight] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[tx.GetTxsListByBlockHeight] Get Txs Error")
		return nil, ErrNotFound
	}

	for _, tx := range txs {
		key := fmt.Sprintf("%s%v", cacheZecreyTxIdPrefix, tx.ID)
		val, err := m.RedisConn.Get(key)
		if err != nil {
			errInfo := fmt.Sprintf("[tx.GetTxsListByBlockHeight] Get Redis Error: %s, key:%s", err.Error(), key)
			logx.Errorf(errInfo)
			return nil, err
		} else if val == "" {
			err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
			if err != nil {
				logx.Error("[tx.GetTxsListByBlockHeight] Get Associate TxDetails Error")
				return nil, err
			}

			// json string
			jsonString, err := json.Marshal(tx.TxDetails)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByBlockHeight] json.Marshal Error: %s, value: %v", tx.TxDetails)
				return nil, err
			}
			// todo
			err = m.RedisConn.Setex(key, string(jsonString), 60*10+rand.Intn(60*3))
			if err != nil {
				logx.Errorf("[tx.GetTxsListByBlockHeight] redis set error: %s", err.Error())
				return nil, err
			}
		} else {
			// json string unmarshal
			var (
				nTxDetails []*TxDetail
			)
			err = json.Unmarshal([]byte(val), &nTxDetails)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByBlockHeight] json.Unmarshal error: %s, value : %s", err.Error(), val)
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
	Return: txs []*Tx, err error
	Description: used for getTxsListByAccountIndex API, return all txs related to accountIndex.
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
		logx.Error("[tx.GetTxsListByAccountIndex] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[info.GetTxsListByAccountIndex] Get TxIds Error")
		return nil, ErrNotFound
	}
	dbTx = m.DB.Table(m.table).Order("created_at desc, id desc").Offset(offset).Limit(limit).Find(&txs, txIds)
	if dbTx.Error != nil {
		logx.Error("[tx.GetTxsListByAccountIndex] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[tx.GetTxsListByAccountIndex] Get Txs Error")
		return nil, ErrNotFound
	}
	//TODO: cache operation
	for _, tx := range txs {
		key := fmt.Sprintf("%s%v", cacheZecreyTxIdPrefix, tx.ID)
		val, err := m.RedisConn.Get(key)
		if err != nil {
			errInfo := fmt.Sprintf("[tx.GetTxsListByAccountIndex] Get Redis Error: %s, key:%s", err.Error(), key)
			logx.Errorf(errInfo)
			return nil, err

		} else if val == "" {
			err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
			if err != nil {
				logx.Error("[tx.GetTxsListByAccountIndex] Get Associate TxDetails Error")
				return nil, err
			}

			// json string
			jsonString, err := json.Marshal(tx.TxDetails)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByAccountIndex] json.Marshal Error: %s, value: %v", tx.TxDetails)
				return nil, err
			}
			// todo
			err = m.RedisConn.Setex(key, string(jsonString), 60*10+rand.Intn(60*3))
			if err != nil {
				logx.Errorf("[tx.GetTxsListByAccountIndex] redis set error: %s", err.Error())
				return nil, err
			}
		} else {
			// json string unmarshal
			var (
				nTxDetails []*TxDetail
			)
			err = json.Unmarshal([]byte(val), &nTxDetails)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByAccountIndex] json.Unmarshal error: %s, value : %s", err.Error(), val)
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
	Return: txs []*Tx, err error
	Description: used for getTxsListByAccountIndex API, return all txs related to accountIndex and txType.
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
		logx.Error("[tx.GetTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[info.GetTxsListByAccountIndexAndTxType] Get TxIds Error")
		return nil, ErrNotFound
	}
	dbTx = m.DB.Table(m.table).Order("created_at desc").Where("tx_type = ?", txType).Offset(offset).Limit(limit).Find(&txs, txIds)
	if dbTx.Error != nil {
		logx.Error("[tx.GetTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[tx.GetTxsListByAccountIndexAndTxType] Get Txs Error")
		return nil, ErrNotFound
	}
	//TODO: cache operation
	for _, tx := range txs {
		key := fmt.Sprintf("%s%v:txType:%v", cacheZecreyTxIdPrefix, tx.ID, txType)
		val, err := m.RedisConn.Get(key)
		if err != nil {
			errInfo := fmt.Sprintf("[tx.GetTxsListByAccountIndexAndTxType] Get Redis Error: %s, key:%s", err.Error(), key)
			logx.Errorf(errInfo)
			return nil, err

		} else if val == "" {
			err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
			if err != nil {
				logx.Error("[tx.GetTxsListByAccountIndexAndTxType] Get Associate TxDetails Error")
				return nil, err
			}

			// json string
			jsonString, err := json.Marshal(tx.TxDetails)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByAccountIndexAndTxType] json.Marshal Error: %s, value: %v", tx.TxDetails)
				return nil, err
			}
			// todo
			err = m.RedisConn.Setex(key, string(jsonString), 30)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByAccountIndexAndTxType] redis set error: %s", err.Error())
				return nil, err
			}
		} else {
			// json string unmarshal
			var (
				nTxDetails []*TxDetail
			)
			err = json.Unmarshal([]byte(val), &nTxDetails)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByAccountIndexAndTxType] json.Unmarshal error: %s, value : %s", err.Error(), val)
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
	Return: txs []*Tx, err error
	Description: used for getTxsListByAccountIndex API, return all txs related to accountIndex and txTypeArray.
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
		logx.Error("[tx.GetTxsListByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[info.GetTxsListByAccountIndexAndTxTypeArray] Get TxIds Error")
		return nil, ErrNotFound
	}
	dbTx = m.DB.Table(m.table).Order("created_at desc").Where("tx_type in (?)", txTypeArray).Offset(offset).Limit(limit).Find(&txs, txIds)
	if dbTx.Error != nil {
		logx.Error("[tx.GetTxsListByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[GetTxsListByAccountIndexAndTxTypeArray] Get Txs Error")
		return nil, ErrNotFound
	}
	//TODO: cache operation
	for _, tx := range txs {
		key := fmt.Sprintf("%s%v:txTypeArray:%s", cacheZecreyTxIdPrefix, tx.ID, txTypeArray)
		val, err := m.RedisConn.Get(key)
		if err != nil {
			errInfo := fmt.Sprintf("[tx.GetTxsListByAccountIndexAndTxTypeArray] Get Redis Error: %s, key:%s", err.Error(), key)
			logx.Errorf(errInfo)
			return nil, err

		} else if val == "" {
			err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
			if err != nil {
				logx.Error("[tx.GetTxsListByAccountIndexAndTxTypeArray] Get Associate TxDetails Error")
				return nil, err
			}

			// json string
			jsonString, err := json.Marshal(tx.TxDetails)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByAccountIndexAndTxTypeArray] json.Marshal Error: %s, value: %v", tx.TxDetails)
				return nil, err
			}
			// todo
			err = m.RedisConn.Setex(key, string(jsonString), 30)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByAccountIndexAndTxTypeArray] redis set error: %s", err.Error())
				return nil, err
			}
		} else {
			// json string unmarshal
			var (
				nTxDetails []*TxDetail
			)
			err = json.Unmarshal([]byte(val), &nTxDetails)
			if err != nil {
				logx.Errorf("[tx.GetTxsListByAccountIndexAndTxTypeArray] json.Unmarshal error: %s, value : %s", err.Error(), val)
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
	Return: txs []*Tx, err error
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
		logx.Error("[tx.GetTxsListByAccountName] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[tx.GetTxsListByAccountName] Get TxIds Error")
		return nil, ErrNotFound
	}
	dbTx = m.DB.Table(m.table).Order("created_at desc, id desc").Offset(offset).Limit(limit).Find(&txs, txIds)
	if dbTx.Error != nil {
		logx.Error("[tx.GetTxsListByAccountName] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[tx.GetTxsListByAccountName] Get Txs Error")
		return nil, ErrNotFound
	}
	//TODO: cache operation
	for _, tx := range txs {
		err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
		if err != nil {
			logx.Error("[tx.GetTxsListByAccountName] Get Associate TxDetails Error")
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

	key := fmt.Sprintf("%s", cacheZecreyTxTxCountPrefix)
	val, err := m.RedisConn.Get(key)
	if err != nil {
		errInfo := fmt.Sprintf("[tx.GetTxsTotalCount] Get Redis Error: %s, key:%s", err.Error(), key)
		logx.Errorf(errInfo)
		return 0, err

	} else if val == "" {
		dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
		if dbTx.Error != nil {
			if dbTx.Error == ErrNotFound {
				return 0, nil
			}
			logx.Error("[tx.GetTxsTotalCount] Get Tx Count Error")
			return 0, err
		}

		err = m.RedisConn.Setex(key, strconv.FormatInt(count, 10), 120)
		if err != nil {
			logx.Errorf("[tx.GetTxsTotalCount] redis set error: %s", err.Error())
			return 0, err
		}
	} else {
		count, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			logx.Errorf("[tx.GetTxsListByAccountIndex] strconv.ParseInt error: %s, value : %s", err.Error(), val)
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
		logx.Error("[tx.GetTxsTotalCountByAccountIndex] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[tx.GetTxsTotalCountByAccountIndex] No Txs of account index %d in Tx Table", accountIndex)
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
		logx.Error("[tx.GetTxsTotalCountByAccountIndexAndTxType] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[tx.GetTxsTotalCountByAccountIndexAndTxType] No Txs of account index %d  and tx type %d in Tx Table", accountIndex, txType)
		return 0, nil
	}
	dbTx = m.DB.Table(m.table).Where("id in (?) and deleted_at is NULL and tx_type = ?", txIds, txType).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[tx.GetTxsTotalCountByAccountIndexAndTxTypee] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[tx.GetTxsTotalCountByAccountIndexAndTxType] no txs of account index %d and tx type = %d in mempool", accountIndex, txType)
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
		logx.Error("[tx.GetTxsTotalCountByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[tx.GetTxsTotalCountByAccountIndexAndTxTypeArray] No Txs of account index %d  and tx type %v in Tx Table", accountIndex, txTypeArray)
		return 0, nil
	}
	dbTx = m.DB.Table(m.table).Where("id in (?) and deleted_at is NULL and tx_type in (?)", txIds, txTypeArray).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[tx.GetTxsTotalCountByAccountIndexAndTxTypeArray] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[tx.GetTxsTotalCountByAccountIndexAndTxTypeArray] no txs of account index %d and tx type = %v in mempool", accountIndex, txTypeArray)
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
		logx.Error("[tx.GetTxsTotalCountByBlockHeight] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[tx.GetTxsTotalCountByBlockHeight] No Txs of block height %d in Tx Table", blockHeight)
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetTxByTxHash
	Params: txHash string
	Return: tx Tx, err error
	Description: used for /api/v1/tx/getTxByHash
*/
func (m *defaultTxModel) GetTxByTxHash(txHash string) (tx *Tx, err error) {
	var txForeignKeyColumn = `TxDetails`

	key := fmt.Sprintf("%s%v", cacheZecreyTxTxHashPrefix, txHash)
	val, err := m.RedisConn.Get(key)

	if err != nil {
		errInfo := fmt.Sprintf("[tx.GetTxByTxHash] Get Redis Error: %s, key:%s", err.Error(), key)
		logx.Errorf(errInfo)
		return nil, err

	} else if val == "" {

		dbTx := m.DB.Table(m.table).Where("tx_hash = ?", txHash).Find(&tx)
		if dbTx.Error != nil {
			logx.Error("[tx.GetTxByTxHash] %s", dbTx.Error)
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			logx.Error("[tx.GetTxByTxHash] No such Tx with txHash: %s", txHash)
			return nil, ErrNotFound
		}
		err = m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
		if err != nil {
			logx.Error("[tx.GetTxByTxHash] Get Associate TxDetails Error")
			return nil, err
		}

		// json string
		jsonString, err := json.Marshal(tx)
		if err != nil {
			logx.Errorf("[tx.GetTxByTxHash] json.Marshal Error: %s, value: %v", tx)
			return nil, err
		}

		err = m.RedisConn.Setex(key, string(jsonString), 60*10+rand.Intn(180))
		if err != nil {
			logx.Errorf("[tx.GetTxByTxHash] redis set error: %s", err.Error())
			return nil, err
		}
	} else {
		// json string unmarshal
		var (
			nTx *Tx
		)
		err = json.Unmarshal([]byte(val), &nTx)
		if err != nil {
			logx.Errorf("[tx.GetTxByTxHash] json.Unmarshal error: %s, value : %s", err.Error(), val)
			return nil, err
		}
		tx = nTx
	}

	return tx, nil
}

/*
	Func: GetTxsListGreaterThanBlockHeight
	Params: blockHeight int64
	Return: txs []*Tx, err error
	Description: used for info service
*/

func (m *defaultTxModel) GetTxsListGreaterThanBlockHeight(blockHeight int64) (txs []*Tx, err error) {
	var (
		txForeignKeyColumn = `TxDetails`
	)

	dbTx := m.DB.Table(m.table).Where("block_height >= ? and block_height < ?", blockHeight, blockHeight+maxBlocks).Order("created_at, id").Find(&txs)

	if dbTx.Error != nil {
		logx.Error("[tx.GetTxsListGreaterThanBlockHeight] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[tx.GetTxsListGreaterThanBlockHeight] No txs blockHeight greater than %d", blockHeight)
		return nil, nil
	}

	for _, tx := range txs {
		err := m.DB.Model(&tx).Association(txForeignKeyColumn).Find(&tx.TxDetails)
		if err != nil {
			logx.Error("[tx.GetTxsListGreaterThanBlockHeight] Get Associate TxDetails Error")
			return nil, err
		}
	}
	return txs, nil
}
