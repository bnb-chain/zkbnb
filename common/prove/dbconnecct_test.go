package prove

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dsn        = "host=localhost user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable"
	db, _      = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	dbInfo, _  = db.DB()
	connection = sqlx.NewSqlConnFromDB(dbInfo)

	cacheConf = []cache.NodeConf{{
		RedisConf: redis.RedisConf{
			Host: "127.0.0.1:6379",
			Type: "node",
			Pass: "myredis",
		},
		Weight: 10,
	}}
)
