package accountliquidity

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

var (
	cacheAccountLiquidityIdPrefix                  = "cache::accountLiquidity:id:"
	cacheAccountLiquidityPairAndAccountIndexPrefix = "cache::accountLiquidity:pairAndAccountIndex:"
)

type accountLiquidity struct {
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	redisConn  *redis.Redis
	cache      multcache.MultCache
}

/*
	Func: CreateAccountLiquidity
	Params: liquidity *AccountLiquidity
	Return: err error
	Description: create account liquidity entity
*/
func (m *accountLiquidity) CreateAccountLiquidity(liquidity *AccountLiquidityInfo) error {
	dbTx := m.db.Table(m.table).Create(liquidity)
	if dbTx.Error != nil {
		return dbTx.Error
	}
	// TODO: ensure if logic branch is necessary
	if dbTx.RowsAffected == 0 {
		return ErrIllegalParam
	}
	return nil
}

/*
	Func: CreateAccountLiquidityInBatches
	Params: entities []*AccountLiquidity
	Return: err error
	Description: create account liquidity entities
*/
func (m *accountLiquidity) CreateAccountLiquidityInBatches(entities []*AccountLiquidityInfo) error {
	dbTx := m.db.Table(m.table).CreateInBatches(entities, len(entities))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return ErrIllegalParam
	}
	return nil
}

/*
	Func: GetAccountLiquidityByAccountIndex
	Params: accountIndex uint32
	Return: entities []*AccountLiquidity, err error
	Description: get account liquidity entities by account index
*/
func (m *accountLiquidity) GetAccountLiquidityByAccountIndex(accountIndex uint32) (entities []*AccountLiquidityInfo, err error) {
	dbTx := m.db.Table(m.table).Where("account_index = ?", accountIndex).Find(&entities)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotExistInSql
	}
	return entities, nil
}

/*
	Func: GetLiquidityByAccountIndexandPairIndex
	Params: accountIndex uint32, pairIndex uint32
	Return: accountLiquidity *AccountLiquidity, err error
	Description: get account liquidity entities by account index and pair index
*/
func (m *accountLiquidity) GetLiquidityByAccountIndexAndPairIndex(accountIndex uint32, pairIndex uint32) (accountLiquidity *AccountLiquidityInfo, err error) {
	dbTx := m.db.Table(m.table).Where("account_index = ? AND pair_index = ?", accountIndex, pairIndex).Find(&accountLiquidity)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotExistInSql
	}
	return accountLiquidity, nil
}

/*
	Func: UpdateAccountLiquidity
	Params: liquidity *AccountLiquidity
	Return: err error
	Description: update account liquidity entity
*/
func (m *accountLiquidity) UpdateAccountLiquidity(liquidity *AccountLiquidityInfo) (bool, error) {
	dbTx := m.db.Table(m.table).Where("id = ?", liquidity.ID).Select("*").Updates(liquidity)
	if dbTx.Error != nil {
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return false, ErrNotExistInSql
	}
	return true, dbTx.Error
}

/*
	Func: UpdateAccountLiquidityInBatches
	Params: entities []*AccountLiquidity
	Return: err error
	Description: update account liquidity entities
*/
func (m *accountLiquidity) UpdateAccountLiquidityInBatches(entities []*AccountLiquidityInfo) error {
	err := m.db.Table(m.table).Transaction(
		func(tx *gorm.DB) error { // transact
			for _, entity := range entities {
				dbTx := tx.Table(m.table).Where("id = ?", entity.ID).Select("*").Updates(entity)
				if dbTx.Error != nil {
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					accountAssetInfo, err := json.Marshal(entity)
					if err != nil {
						return err
					}
					logx.Error("[liquidity.UpdateAccountLiquidityInBatches]" + "invalid liquidity, " + string(accountAssetInfo))
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
	Return: accountLiquidity *AccountLiquidity, err error
	Description: used for constructing MPT
*/
func (m *accountLiquidity) GetAllLiquidityAssets() (accountLiquidity []*AccountLiquidityInfo, err error) {
	dbTx := m.db.Table(m.table).Order("account_index, pair_index").Find(&accountLiquidity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetAllLiquidityAssets] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetAllLiquidityAssets] %s", ErrNotExistInSql)
		logx.Error(err)
		return accountLiquidity, ErrNotExistInSql
	}
	return accountLiquidity, nil
}
