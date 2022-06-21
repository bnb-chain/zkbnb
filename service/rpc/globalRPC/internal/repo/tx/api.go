package tx

import (
	"sync"

	"github.com/bnb-chain/zkbas/pkg/multcache"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Tx interface {
	GetTxsTotalCount() (count int64, err error)
	GetTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
	GetTxsTotalCountByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8) (count int64, err error)
	GetTxsListByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8, limit int, offset int) (txs []*Tx, err error)
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
			cache:     multcache.NewGoCache(100, 10),
		}
	})
	return singletonValue
}
