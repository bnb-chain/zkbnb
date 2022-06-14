package block

import (
	"context"
	"sync"

	table "github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Block interface {
	GetCommitedBlocksCount() (count int64, err error)
	GetExecutedBlocksCount() (count int64, err error)
	GetBlockWithTxsByCommitment(BlockCommitment string) (block *table.Block, err error)
	GetBlockByBlockHeight(blockHeight int64) (block *table.Block, err error)
	GetBlockWithTxsByBlockHeight(blockHeight int64) (block *table.Block, err error)
	GetBlocksList(limit int64, offset int64) (blocks []*table.Block, err error)
	GetBlocksTotalCount() (count int64, err error)
}

var singletonValue *block
var once sync.Once

func New(c config.Config) Block {
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
		singletonValue = &block{
			cachedConn: sqlc.NewConn(conn, c.CacheRedis),
			table:      BlockTableName,
			db:         gormPointer,
			redisConn:  redisConn,
			cache:      multcache.NewGoCache(context.Background(), 100, 10),
			blockModel: table.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		}
	})
	return singletonValue
}
