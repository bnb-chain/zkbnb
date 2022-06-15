package l2asset

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/assetInfo"

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
func (m *l2asset) GetL2AssetsList() (res []*table.AssetInfo, err error) {
	dbTx := m.db.Table(m.table).Find(&res)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return res, nil
}

/*
	Func: GetL2AssetInfoBySymbol
	Params: symbol string
	Return: res *L2AssetInfo, err error
	Description: get l2 asset info by l2 symbol
*/
func (m *l2asset) GetL2AssetInfoBySymbol(symbol string) (res *table.AssetInfo, err error) {
	dbTx := m.db.Table(m.table).Where("asset_symbol = ?", symbol).Find(&res)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, ErrNotExistInSql
	}
	return res, nil
}

/*
	Func: GetSimpleL2AssetInfoByAssetId
	Params: assetId uint32
	Return: L2AssetInfo, error
	Description: get layer-2 asset info by assetId
*/
func (m *l2asset) GetSimpleL2AssetInfoByAssetId(assetId uint32) (res *table.AssetInfo, err error) {
	dbTx := m.db.Table(m.table).Where("asset_id = ?", assetId).Find(&res)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return res, nil
}
