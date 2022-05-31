package liquidity

import (
	"context"
	"sync"

	table "github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Liquidity interface {
	GetLiquidityByPairIndex(pairIndex int64) (entity *table.Liquidity, err error)
}

var singletonValue *liquidity
var once sync.Once

func New(c config.Config) Liquidity {
	once.Do(func() {
		conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
		gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
		if err != nil {
			logx.Errorf("gorm connect db error, err = %s", err.Error())
		}
		redisConn := redis.New(c.CacheRedis[0].Host, func(p *redis.Redis) {
			p.Type = c.CacheRedis[0].Type
			p.Pass = c.CacheRedis[0].Pass
		})
		singletonValue = &liquidity{
			cachedConn: sqlc.NewConn(conn, c.CacheRedis),
			table:      `account_liquidity`,
			db:         gormPointer,
			redisConn:  redisConn,
			cache:      multcache.NewGoCache(context.Background(), 100, 10),
		}
	})
	return singletonValue
}
