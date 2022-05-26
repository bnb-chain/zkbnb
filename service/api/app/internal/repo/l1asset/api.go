package l1asset

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

type L1asset interface {
	GetAssets() (assets []*L1AssetInfo, err error)
}

var singletonValue *l1asset
var once sync.Once

func New(c config.Config) L1asset {
	once.Do(func() {
		conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
		gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
		if err != nil {
			logx.Errorf("gorm connect db error, err = %s", err.Error())
		}
		singletonValue = &l1asset{
			cachedConn: sqlc.NewConn(conn, c.CacheRedis),
			table:      `l1_asset_info`,
			db:         gormPointer,
			cache:      multcache.NewGoCache(context.Background(), 100, 10),
		}
	})
	return singletonValue
}
