package asset

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
	cacheAccountLiquidityIdPrefix                  = "cache::accountLiquidity:id:"
	cacheAccountLiquidityPairAndAccountIndexPrefix = "cache::accountLiquidity:pairAndAccountIndex:"
)

type (
	AccountLiquidityModel interface {
		CreateAccountLiquidityTable() error
		DropAccountLiquidityTable() error
		CreateAccountLiquidity(liquidity *AccountLiquidity) error
		CreateAccountLiquidityInBatches(entities []*AccountLiquidity) error
		GetAccountLiquidityByAccountIndex(accountIndex uint32) (entities []*AccountLiquidity, err error)
		GetLiquidityByAccountIndexAndPairIndex(accountIndex uint32, pairIndex uint32) (accountLiquidity *AccountLiquidity, err error)
		UpdateAccountLiquidity(liquidity *AccountLiquidity) (bool, error)
		UpdateAccountLiquidityInBatches(entities []*AccountLiquidity) error
		GetAllLiquidityAssets() (accountLiquidity []*AccountLiquidity, err error)
	}

	defaultAccountLiquidityModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	AccountLiquidity struct {
		gorm.Model
		AccountIndex int64 `gorm:"index"`
		PairIndex    int64
		AssetA       int64
		AssetB       int64
		LpAmount     int64
	}
)

func NewAccountLiquidityModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AccountLiquidityModel {
	return &defaultAccountLiquidityModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      LiquidityAssetTable,
		DB:         db,
	}
}

func (*AccountLiquidity) TableName() string {
	return LiquidityAssetTable
}

/*
	Func: CreateAccountLiquidityTable
	Params:
	Return: err error
	Description: create account liquidity table
*/
func (m *defaultAccountLiquidityModel) CreateAccountLiquidityTable() error {
	return m.DB.AutoMigrate(AccountLiquidity{})
}

/*
	Func: DropAccountLiquidityTable
	Params:
	Return: err error
	Description: drop account liquidity table
*/
func (m *defaultAccountLiquidityModel) DropAccountLiquidityTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateAccountLiquidity
	Params: liquidity *AccountLiquidity
	Return: err error
	Description: create account liquidity entity
*/
func (m *defaultAccountLiquidityModel) CreateAccountLiquidity(liquidity *AccountLiquidity) error {
	dbTx := m.DB.Table(m.table).Create(liquidity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateAccountLiquidity] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.CreateAccountLiquidity] %s", ErrInvalidAccountLiquidityInput)
		logx.Error(err)
		return ErrInvalidAccountLiquidityInput
	}
	return nil
}

/*
	Func: CreateAccountLiquidityInBatches
	Params: entities []*AccountLiquidity
	Return: err error
	Description: create account liquidity entities
*/
func (m *defaultAccountLiquidityModel) CreateAccountLiquidityInBatches(entities []*AccountLiquidity) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(entities, len(entities))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.CreateAccountLiquidityInBatches] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.CreateAccountLiquidityInBatches] %s", ErrInvalidAccountLiquidityInput)
		logx.Error(err)
		return ErrInvalidAccountLiquidityInput
	}
	return nil
}

/*
	Func: GetAccountLiquidityByAccountIndex
	Params: accountIndex uint32
	Return: entities []*AccountLiquidity, err error
	Description: get account liquidity entities by account index
*/
func (m *defaultAccountLiquidityModel) GetAccountLiquidityByAccountIndex(accountIndex uint32) (entities []*AccountLiquidity, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&entities)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetAccountLiquidityByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetAccountLiquidityByAccountIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return entities, nil
}

/*
	Func: GetLiquidityByAccountIndexandPairIndex
	Params: accountIndex uint32, pairIndex uint32
	Return: accountLiquidity *AccountLiquidity, err error
	Description: get account liquidity entities by account index and pair index
*/
func (m *defaultAccountLiquidityModel) GetLiquidityByAccountIndexAndPairIndex(accountIndex uint32, pairIndex uint32) (accountLiquidity *AccountLiquidity, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? AND pair_index = ?", accountIndex, pairIndex).Find(&accountLiquidity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetLiquidityByAccountIndexandPairIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetLiquidityByAccountIndexandPairIndex] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return accountLiquidity, nil
}

/*
	Func: UpdateAccountLiquidity
	Params: liquidity *AccountLiquidity
	Return: err error
	Description: update account liquidity entity
*/
func (m *defaultAccountLiquidityModel) UpdateAccountLiquidity(liquidity *AccountLiquidity) (bool, error) {
	dbTx := m.DB.Table(m.table).Where("id = ?", liquidity.ID).
		Select("*").
		Updates(liquidity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.UpdateAccountLiquidity] %s", dbTx.Error)
		logx.Error(err)
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.UpdateAccountLiquidity] %s", ErrInvalidAccountLiquidityInput)
		logx.Error(err)
		return false, ErrInvalidAccountLiquidityInput
	}
	return true, dbTx.Error
}

/*
	Func: UpdateAccountLiquidityInBatches
	Params: entities []*AccountLiquidity
	Return: err error
	Description: update account liquidity entities
*/
func (m *defaultAccountLiquidityModel) UpdateAccountLiquidityInBatches(entities []*AccountLiquidity) error {
	err := m.DB.Table(m.table).Transaction(
		func(tx *gorm.DB) error { // transact
			for _, entity := range entities {
				dbTx := tx.Table(m.table).Where("id = ?", entity.ID).
					Select("*").
					Updates(entity)
				if dbTx.Error != nil {
					err := fmt.Sprintf("[liquidity.UpdateAccountLiquidityInBatches] %s", dbTx.Error)
					logx.Error(err)
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					accountAssetInfo, err := json.Marshal(entity)
					if err != nil {
						res := fmt.Sprintf("[liquidity.UpdateAccountLiquidityInBatches] %s", err)
						logx.Error(res)
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
func (m *defaultAccountLiquidityModel) GetAllLiquidityAssets() (accountLiquidity []*AccountLiquidity, err error) {
	dbTx := m.DB.Table(m.table).Order("account_index, pair_index").Find(&accountLiquidity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetAllLiquidityAssets] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetAllLiquidityAssets] %s", ErrNotFound)
		logx.Error(err)
		return accountLiquidity, ErrNotFound
	}
	return accountLiquidity, nil
}
