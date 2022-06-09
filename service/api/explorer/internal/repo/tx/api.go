package tx

import (
	"context"
	"sync"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Tx interface {
	GetTxsTotalCount() (count int64, err error)
	GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
}

var singletonValue *tx
var once sync.Once

func New(c config.Config) Tx {
	once.Do(func() {
		gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
		if err != nil {
			logx.Errorf("gorm connect db error, err = %s", err.Error())
		}
		redisConn := redis.New(c.CacheRedis[0].Host, func(p *redis.Redis) {
			p.Type = c.CacheRedis[0].Type
			p.Pass = c.CacheRedis[0].Pass
		})
		singletonValue = &tx{
			table:     `tx`,
			db:        gormPointer,
			redisConn: redisConn,
			cache:     multcache.NewGoCache(context.Background(), 100, 10),
		}
	})
	return singletonValue
}
