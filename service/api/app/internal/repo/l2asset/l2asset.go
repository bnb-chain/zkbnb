package l2asset

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	table "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type l2asset struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

/*
	Func: GetL2AssetsList
	Params:
	Return: err error
	Description: create account table
*/
func (m *l2asset) GetL2AssetsList(ctx context.Context) ([]*table.AssetInfo, error) {
	f := func() (interface{}, error) {
		var res []*table.AssetInfo
		dbTx := m.db.Table(m.table).Find(&res)
		if dbTx.Error != nil {
			logx.Errorf("fail to get assets, error: %s", dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		}
		if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		return &res, nil
	}
	var res []*table.AssetInfo
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetL2AssetsList, &res, multcache.AssetListTtl, f)
	if err != nil {
		return nil, err
	}
	res1, ok := value.(*[]*table.AssetInfo)
	if !ok {
		return nil, fmt.Errorf("[GetL2AssetsList] ErrConvertFail")
	}
	return *res1, nil
}

/*
	Func: GetL2AssetInfoBySymbol
	Params: symbol string
	Return: res *L2AssetInfo, err error
	Description: get l2 asset info by l2 symbol
*/
func (m *l2asset) GetL2AssetInfoBySymbol(ctx context.Context, symbol string) (*table.AssetInfo, error) {
	f := func() (interface{}, error) {
		res := table.AssetInfo{}
		dbTx := m.db.Table(m.table).Where("asset_symbol = ?", symbol).Find(&res)
		if dbTx.Error != nil {
			logx.Errorf("fail to get asset by symbol: %s, error: %s", symbol, dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		}
		if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		return &res, nil
	}
	res := table.AssetInfo{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetL2AssetInfoBySymbol+symbol, &res, multcache.AssetTtl, f)
	if err != nil {
		return nil, err
	}
	res1, ok := value.(*table.AssetInfo)
	if !ok {
		return nil, fmt.Errorf("[GetL2AssetInfoBySymbol] ErrConvertFail")
	}
	return res1, nil
}

/*
	Func: GetSimpleL2AssetInfoByAssetId
	Params: assetId uint32
	Return: L2AssetInfo, error
	Description: get layer-2 asset info by assetId
*/
func (m *l2asset) GetSimpleL2AssetInfoByAssetId(ctx context.Context, assetId uint32) (*table.AssetInfo, error) {
	f := func() (interface{}, error) {
		res := table.AssetInfo{}
		dbTx := m.db.Table(m.table).Where("asset_id = ?", assetId).Find(&res)
		if dbTx.Error != nil {
			logx.Errorf("fail to get asset by id: %d, error: %s", assetId, dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		}
		if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		return &res, nil
	}
	res := table.AssetInfo{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetSimpleL2AssetInfoByAssetId+strconv.Itoa(int(assetId)), &res, multcache.AssetTtl, f)
	if err != nil {
		return nil, err
	}
	res1, ok := value.(*table.AssetInfo)
	if !ok {
		return nil, fmt.Errorf("[GetSimpleL2AssetInfoByAssetId] ErrConvertFail")
	}
	return res1, nil
}
