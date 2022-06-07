package commglobalmap

import (
	"sync"

	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	commGlobalmapHandler "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Commglobalmap interface {
	GetLatestAccountInfo(accountIndex int64) (accountInfo *commGlobalmapHandler.AccountInfo, err error)
	GetLatestLiquidityInfoForRead(pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error)
}

var singletonValue *commglobalmap
var once sync.Once

func withRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func New(c config.Config) Commglobalmap {
	once.Do(func() {
		conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
		gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
		if err != nil {
			logx.Errorf("gorm connect db error, err = %s", err.Error())
		}
		redisConn := redis.New(c.CacheRedis[0].Host, withRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
		singletonValue = &commglobalmap{
			AccountModel:    account.NewAccountModel(conn, c.CacheRedis, gormPointer),
			MempoolModel:    mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
			RedisConnection: redisConn,
		}
	})
	return singletonValue
}
