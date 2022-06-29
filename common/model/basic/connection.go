package basic

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dsn = "host=34.122.163.215 user=postgres password=zecreyLegendTest@ dbname=zecreylegend port=5432 sslmode=disable"
	//dsn        = "host=localhost user=postgres password=ZecreyProtocolDB@123 dbname=zecreylegend port=5432 sslmode=disable"
	DB, _      = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	DbInfo, _  = DB.DB()
	Connection = sqlx.NewSqlConnFromDB(DbInfo)
	CacheConf  = []cache.NodeConf{{
		RedisConf: redis.RedisConf{
			Host: "127.0.0.1:6379",
			Type: "node",
			Pass: "myredis",
		},
		Weight: 10,
	}}
)
