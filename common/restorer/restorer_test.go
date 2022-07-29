package restorer

import (
	"math/big"
	"os"
	"testing"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dsn = "host=localhost user=postgres password=*** dbname=zkbas port=5432 sslmode=disable"
)

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func TestRestoreManager(t *testing.T) {
	c := &Config{
		Postgres: struct{ DataSource string }{DataSource: dsn},
		CacheRedis: []cache.NodeConf{{
			RedisConf: redis.RedisConf{
				Host: "127.0.0.1:6379",
				Type: "node",
				Pass: "myredis",
			},
			Weight: 10,
		}},
	}
	db, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	restoreMnager, err := NewRestoreManager(db, conn, redisConn, c.CacheRedis)
	if err != nil {
		logx.Errorf("new restore manager failed: %s", err.Error())
		os.Exit(1)
	}

	err = restoreMnager.RestoreHistoryData(restoreMnager.l1genesisNumber, new(big.Int).Add(restoreMnager.l1genesisNumber, big.NewInt(1000)))
	if err != nil {
		t.Errorf("restore hitory data failed: %v", err)
	}
}
