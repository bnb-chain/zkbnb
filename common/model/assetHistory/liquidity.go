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

package assetHistory

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheAccountLiquidityHistoryIdPrefix                  = "cache::AccountLiquidityHistory:id:"
	cacheAccountLiquidityHistoryPairAndAccountIndexPrefix = "cache::AccountLiquidityHistory:pairAndAccountIndex:"
)

type (
	AccountLiquidityHistoryModel interface {
		CreateAccountLiquidityHistoryTable() error
		DropAccountLiquidityHistoryTable() error
		CreateAccountLiquidityHistory(liquidity *AccountLiquidityHistory) error
		CreateAccountLiquidityHistoryInBatches(entities []*AccountLiquidityHistory) error
		GetAccountLiquidityHistoryByAccountIndex(accountIndex uint32) (entities []*AccountLiquidityHistory, err error)
		GetLiquidityByAccountIndexAndPairIndex(accountIndex uint32, pairIndex uint32) (AccountLiquidityHistory *AccountLiquidityHistory, err error)
		UpdateAccountLiquidityHistory(liquidity *AccountLiquidityHistory) (bool, error)
		UpdateAccountLiquidityHistoryInBatches(entities []*AccountLiquidityHistory) error
		GetAllLiquidityAssets() (AccountLiquidityHistory []*AccountLiquidityHistory, err error)
		GetLatestAccountLiquidityAssetsByBlockHeight(height int64) (
			rowsAffected int64, accountAssets []*AccountLiquidityHistory, err error,
		)
		GetLatestLiquidityAsset(
			accountIndex uint32, pairIndex uint32) (
			rowsAffected int64, liquidityAsset *AccountLiquidityHistory, err error)
		GetLiquidityAssetsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, liquidityAssets []*AccountLiquidityHistory, err error)
	}

	defaultAccountLiquidityHistoryModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	AccountLiquidityHistory struct {
		gorm.Model
		AccountIndex  int64 `gorm:"index"`
		PairIndex     int64
		AssetAId      int64
		AssetA        string
		AssetBId      int64
		AssetB        string
		LpAmount      string
		L2BlockHeight int64
	}
)

func NewAccountLiquidityHistoryModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AccountLiquidityHistoryModel {
	return &defaultAccountLiquidityHistoryModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      LiquidityAssetHistoryTable,
		DB:         db,
	}
}

func (*AccountLiquidityHistory) TableName() string {
	return LiquidityAssetHistoryTable
}

/*
	Func: CreateAccountLiquidityHistoryTable
	Params:
	Return: err error
	Description: create account liquidity table
*/
func (m *defaultAccountLiquidityHistoryModel) CreateAccountLiquidityHistoryTable() error {
	return m.DB.AutoMigrate(AccountLiquidityHistory{})
}

/*
	Func: DropAccountLiquidityHistoryTable
	Params:
	Return: err error
	Description: drop account liquidity table
*/
func (m *defaultAccountLiquidityHistoryModel) DropAccountLiquidityHistoryTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateAccountLiquidityHistory
	Params: liquidity *AccountLiquidityHistory
	Return: err error
	Description: create account liquidity entity
*/
func (m *defaultAccountLiquidityHistoryModel) CreateAccountLiquidityHistory(liquidity *AccountLiquidityHistory) error {
	dbTx := m.DB.Table(m.table).Create(liquidity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateAccountLiquidityHistory] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.CreateAccountLiquidityHistory] %s", ErrInvalidAccountLiquidityHistoryInput)
		logx.Error(err)
		return ErrInvalidAccountLiquidityHistoryInput
	}
	return nil
}

/*
	Func: CreateAccountLiquidityHistoryInBatches
	Params: entities []*AccountLiquidityHistory
	Return: err error
	Description: create account liquidity entities
*/
func (m *defaultAccountLiquidityHistoryModel) CreateAccountLiquidityHistoryInBatches(entities []*AccountLiquidityHistory) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(entities, len(entities))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateAccountLiquidityHistoryInBatches] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.CreateAccountLiquidityHistoryInBatches] %s", ErrInvalidAccountLiquidityHistoryInput)
		logx.Error(err)
		return ErrInvalidAccountLiquidityHistoryInput
	}
	return nil
}

/*
	Func: GetAccountLiquidityHistoryByAccountIndex
	Params: accountIndex uint32
	Return: entities []*AccountLiquidityHistory, err error
	Description: get account liquidity entities by account index
*/
func (m *defaultAccountLiquidityHistoryModel) GetAccountLiquidityHistoryByAccountIndex(accountIndex uint32) (entities []*AccountLiquidityHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&entities)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetAccountLiquidityHistoryByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetAccountLiquidityHistoryByAccountIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return entities, nil
}

/*
	Func: GetLiquidityByAccountIndexandPairIndex
	Params: accountIndex uint32, pairIndex uint32
	Return: AccountLiquidityHistory *AccountLiquidityHistory, err error
	Description: get account liquidity entities by account index and pair index
*/
func (m *defaultAccountLiquidityHistoryModel) GetLiquidityByAccountIndexAndPairIndex(accountIndex uint32, pairIndex uint32) (AccountLiquidityHistory *AccountLiquidityHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? AND pair_index = ?", accountIndex, pairIndex).Find(&AccountLiquidityHistory)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetLiquidityByAccountIndexandPairIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetLiquidityByAccountIndexandPairIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return AccountLiquidityHistory, nil
}

/*
	Func: UpdateAccountLiquidityHistory
	Params: liquidity *AccountLiquidityHistory
	Return: err error
	Description: update account liquidity entity
*/
func (m *defaultAccountLiquidityHistoryModel) UpdateAccountLiquidityHistory(liquidity *AccountLiquidityHistory) (bool, error) {
	dbTx := m.DB.Table(m.table).Where("id = ?", liquidity.ID).
		Select("*").
		Updates(liquidity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.UpdateAccountLiquidityHistory] %s", dbTx.Error)
		logx.Error(err)
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.UpdateAccountLiquidityHistory] %s", ErrInvalidAccountLiquidityHistoryInput)
		logx.Error(err)
		return false, ErrInvalidAccountLiquidityHistoryInput
	}
	return true, dbTx.Error
}

/*
	Func: UpdateAccountLiquidityHistoryInBatches
	Params: entities []*AccountLiquidityHistory
	Return: err error
	Description: update account liquidity entities
*/
func (m *defaultAccountLiquidityHistoryModel) UpdateAccountLiquidityHistoryInBatches(entities []*AccountLiquidityHistory) error {
	err := m.DB.Table(m.table).Transaction(
		func(tx *gorm.DB) error { // transact
			for _, entity := range entities {
				dbTx := tx.Table(m.table).Where("id = ?", entity.ID).
					Select("*").
					Updates(entity)
				if dbTx.Error != nil {
					err := fmt.Sprintf("[liquidity.UpdateAccountLiquidityHistoryInBatches] %s", dbTx.Error)
					logx.Error(err)
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					accountAssetInfo, err := json.Marshal(entity)
					if err != nil {
						res := fmt.Sprintf("[liquidity.UpdateAccountLiquidityHistoryInBatches] %s", err)
						logx.Error(res)
						return err
					}
					logx.Error("[liquidity.UpdateAccountLiquidityHistoryInBatches]" + "invalid liquidity, " + string(accountAssetInfo))
					return errors.New("[liquidity storage] err: invalid liquidity, " + string(accountAssetInfo))
				}
			}
			return nil
		})
	return err
}

/*
	Func: GetAllLiquidityAssets
	Params:
	Return: AccountLiquidityHistory *AccountLiquidityHistory, err error
	Description: used for constructing MPT
*/
func (m *defaultAccountLiquidityHistoryModel) GetAllLiquidityAssets() (AccountLiquidityHistory []*AccountLiquidityHistory, err error) {
	dbTx := m.DB.Table(m.table).Order("account_index, pair_index").Find(&AccountLiquidityHistory)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetAllLiquidityAssets] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetAllLiquidityAssets] %s", ErrNotFound)
		logx.Error(err)
		return AccountLiquidityHistory, ErrNotFound
	}
	return AccountLiquidityHistory, nil
}

func (m *defaultAccountLiquidityHistoryModel) GetLatestAccountLiquidityAssetsByBlockHeight(height int64) (
	rowsAffected int64, accountAssets []*AccountLiquidityHistory, err error,
) {
	dbTx := m.DB.Table(m.table).
		Raw("SELECT a.* FROM account_liquidity_history a WHERE NOT EXISTS"+
			"(SELECT * FROM account_liquidity_history WHERE account_index = a.account_index AND pair_index = a.pair_index AND l2_block_height <= ? AND l2_block_height > a.l2_block_height) "+
			"AND l2_block_height <= ? ORDER BY account_index, pair_index", height, height).
		Find(&accountAssets)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestAccountLiquidityAssetsByBlockHeight] unable to get related assets: %s", dbTx.Error.Error())
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, accountAssets, nil
}

func (m *defaultAccountLiquidityHistoryModel) GetLatestLiquidityAsset(
	accountIndex uint32, pairIndex uint32) (
	rowsAffected int64, liquidityAsset *AccountLiquidityHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? AND pair_index = ?", accountIndex, pairIndex).Order("l2_block_height desc").Find(&liquidityAsset)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestLiquidityAsset] unable to get related assets: %s", dbTx.Error.Error())
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, liquidityAsset, nil
}

func (m *defaultAccountLiquidityHistoryModel) GetLiquidityAssetsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, liquidityAssets []*AccountLiquidityHistory, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height = ?", l2BlockHeight).Find(&liquidityAssets)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[GetLiquidityAssetsByBlockHeight] unable to get related assets: %s", err.Error())
		logx.Error(errInfo)
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, liquidityAssets, nil
}
