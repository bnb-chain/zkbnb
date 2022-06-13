package l2asset

import (
	"context"
	"sync"

	table "github.com/zecrey-labs/zecrey-legend/common/model/l2asset"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type L2asset interface {
	GetL2AssetsList() (res []*table.L2AssetInfo, err error)
	GetL2AssetInfoBySymbol(symbol string) (res *table.L2AssetInfo, err error)
	GetSimpleL2AssetInfoByAssetId(assetId uint32) (res *table.L2AssetInfo, err error)
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
