package l1asset

import (
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

var (
	cacheL1AssetInfoIdPrefix        = "cache:l1AssetInfo:id:"
	cacheL1AssetInfoL2AssetPkPrefix = "cache:l1AssetInfo:l2_asset_pk:"
)

type l1asset struct {
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	cache      multcache.MultCache
}

type L1AssetInfo struct {
	gorm.Model
	ChainId           int64 `gorm:"index"`
	AssetId           int64 `gorm:"index"`
	L2AssetPk         int64 `gorm:"index"`
	AssetName         string
	AssetSymbol       string
	AssetAddress      string
	Decimals          int64
	LockedAssetAmount int64
	WithdrawFeeRate   int64
}

/*
	Func: GetAssets
	Params:
	Return: assets []*L1AssetInfo, err error
	Description:
*/
func (m *l1asset) GetAssets() (assets []*L1AssetInfo, err error) {
	// TODO: select all data in table?
	dbTx := m.db.Table(m.table).Find(&assets)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1asset.GetAssets] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1asset.GetAssets] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return assets, dbTx.Error
}
