package mempool

import (
	"context"
	"sync"

	table "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Mempool interface {
	GetMempoolTxs(offset int64, limit int64) (mempoolTx []*table.MempoolTx, err error)
	GetMempoolTxsTotalCount() (count int64, err error)
	GetMempoolTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
	GetMempoolTxsListByAccountIndex(accountIndex int64, limit int64, offset int64) (mempoolTxs []*table.MempoolTx, err error)

	//GetMempoolTxsTotalCountByPublicKey(pk string) (mempoolTx []*types.Tx, err error)
	GetMempoolTxByTxHash(hash string) (mempoolTxs *table.MempoolTx, err error)
	GetAccountAssetMempoolDetails(accountIndex int64, assetId int64, assetType int64) (mempoolTxDetails []*table.MempoolTxDetail, err error)
	GetMempoolTxsTotalCountByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8) (count int64, err error)
	GetMempoolTxsListByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8, limit int64, offset int64) (mempoolTxs []*table.MempoolTx, err error)
}

var singletonValue *mempool
var once sync.Once

func New(c config.Config) Mempool {
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
		singletonValue = &mempool{
			cachedConn: sqlc.NewConn(conn, c.CacheRedis),
			table:      MempoolTableName,
			db:         gormPointer,
			redisConn:  redisConn,
			cache:      multcache.NewGoCache(context.Background(), 100, 10),
		}
	})
	return singletonValue
}
