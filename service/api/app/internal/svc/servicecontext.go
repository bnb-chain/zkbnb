package svc

import (
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"github.com/bnb-chain/zkbas/service/api/app/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	Conn        sqlx.SqlConn
	GormPointer *gorm.DB
	RedisConn   *redis.Redis
	Cache       multcache.MultCache
}

func NewServiceContext(c config.Config) *ServiceContext {
	g, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Must(err)
	}
	return &ServiceContext{
		Config:      c,
		Conn:        sqlx.NewSqlConn("postgres", c.Postgres.DataSource),
		GormPointer: g,
		RedisConn: redis.New(c.CacheRedis[0].Host, func(p *redis.Redis) {
			p.Type = c.CacheRedis[0].Type
			p.Pass = c.CacheRedis[0].Pass
		}),
		Cache: multcache.NewGoCache(100, 10),
	}
}
