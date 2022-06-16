package accounthistory

import (
	"context"
	"sync"

	table "github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AccountHistory interface {
	GetAccountByAccountName(accountName string) (account *table.AccountHistory, err error)
	GetAccountByAccountIndex(accountIndex int64) (account *table.AccountHistory, err error)

	GetAccountsList(limit int, offset int64) (accounts []*table.AccountHistory, err error)
	GetAccountsTotalCount() (count int64, err error)
	GetAccountByPk(pk string) (account *table.AccountHistory, err error)
	GetAccountAssetsByIndex(accountIndex int64) (accountAssets []*table.AccountHistory, err error)
}

var singletonValue *accountHistory
var once sync.Once
var c config.Config

func New(c config.Config) AccountHistory {
	once.Do(func() {
		gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
		if err != nil {
			logx.Errorf("gorm connect db error, err = %s", err.Error())
		}
		redisConn := redis.New(c.CacheRedis[0].Host, func(p *redis.Redis) {
			p.Type = c.CacheRedis[0].Type
			p.Pass = c.CacheRedis[0].Pass
		})
		singletonValue = &accountHistory{
			table:     `tx`,
			db:        gormPointer,
			redisConn: redisConn,
			cache:     multcache.NewGoCache(context.Background(), 100, 10),
		}
	})
	return singletonValue
}
