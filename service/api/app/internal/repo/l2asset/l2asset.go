package l2asset

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/l2asset"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

var (
	cacheZecreyL2AssetInfoIdPrefix          = "cache:zecrey:l2AssetInfo:id:"
	cacheZecreyL2AssetInfoL2AssetIdPrefix   = "cache:zecrey:l2AssetInfo:l2AssetId:"
	cacheZecreyL2AssetInfoL2AssetNamePrefix = "cache:zecrey:l2AssetInfo:l2AssetName:"
)

type l2asset struct {
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	cache      multcache.MultCache
}

/*
	Func: GetL2AssetsList
	Params:
	Return: err error
	Description: create account table
*/
func (m *l2asset) GetL2AssetsList() (res []*table.L2AssetInfo, err error) {
	dbTx := m.db.Table(m.table).Find(&res)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, ErrNotExistInSql
	}
	for _, asset := range res {
		err := m.db.Model(&asset).Association("L1AssetsInfo").Find(&asset.L1AssetsInfo)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

/*
	Func: GetL2AssetInfoBySymbol
	Params: symbol string
	Return: res *L2AssetInfo, err error
	Description: get l2 asset info by l2 symbol
*/
func (m *l2asset) GetL2AssetInfoBySymbol(symbol string) (res *table.L2AssetInfo, err error) {
	dbTx := m.db.Table(m.table).Where("l2_symbol = ?", symbol).Find(&res)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, ErrNotExistInSql
	}
	return res, nil
}

/*
	Func: GetL2AssetInfoByAssetId
	Params: assetId uint32
	Return: L2AssetInfo, error
	Description: get layer-2 asset info by assetId
*/
func (m *l2asset) GetL2AssetInfoByAssetId(assetId uint32) (res *table.L2AssetInfo, err error) {
	var L2AssetInfoForeignKeyColumn = "L1AssetsInfo"
	dbTx := m.db.Table(m.table).Where("l2_asset_id = ?", assetId).Find(&res)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, ErrNotExistInSql
	}
	err = m.db.Model(&res).Association(L2AssetInfoForeignKeyColumn).Find(&res.L1AssetsInfo)
	if err != nil {
		return nil, err
	}
	return res, nil
}
