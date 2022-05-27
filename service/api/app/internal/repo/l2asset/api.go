package l2asset

import (
	"context"
	"sync"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type L2asset interface {
	GetL2AssetsList() (res []*L2AssetInfo, err error)
	GetL2AssetInfoBySymbol(symbol string) (res *L2AssetInfo, err error)
	GetL2AssetInfoByAssetId(assetId uint32) (res *L2AssetInfo, err error)

	// CreateL2AssetInfoTable() error
	// DropL2AssetInfoTable() error
	// CreateL2AssetInfo(l2AssetInfo *L2AssetInfo) (bool, error)
	// CreateL2AssetsInfoInBatches(l2AssetsInfo []*L2AssetInfo) (rowsAffected int64, err error)
	// GetL2AssetsCount() (latestHeight int64, err error)
	// GetL2AssetsListWithoutL1AssetsInfo() (res []*L2AssetInfo, err error)
	// GetL2AssetIdByChainIdAndAssetId(chainId uint8, assetId uint32) (l2AssetId int64, err error)
	// GetSimpleL2AssetInfoByAssetId(assetId uint32) (res *L2AssetInfo, err error)
	// GetAssetIdCount() (res int64, err error)
}

var singletonValue *l2asset
var once sync.Once

func New(c config.Config) L2asset {
	once.Do(func() {
		conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
		gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
		if err != nil {
			logx.Errorf("gorm connect db error, err = %s", err.Error())
		}
		singletonValue = &l2asset{
			cachedConn: sqlc.NewConn(conn, c.CacheRedis),
			table:      `l1_asset_info`,
			db:         gormPointer,
			cache:      multcache.NewGoCache(context.Background(), 100, 10),
		}
	})
	return singletonValue
}
