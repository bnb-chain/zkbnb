/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package account

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

type AccountModel interface {
	IfAccountNameExist(name string) (bool, error)
	IfAccountExistsByAccountIndex(accountIndex int64) (bool, error)
	GetAccountByAccountIndex(accountIndex int64) (account *table.Account, err error)
	GetVerifiedAccountByAccountIndex(accountIndex int64) (account *table.Account, err error)
	GetConfirmedAccountByAccountIndex(accountIndex int64) (account *table.Account, err error)
	GetAccountByPk(pk string) (account *table.Account, err error)
	GetAccountByAccountName(accountName string) (account *table.Account, err error)
	GetAccountByAccountNameHash(accountNameHash string) (account *table.Account, err error)
	GetAccountsList(limit int, offset int64) (accounts []*table.Account, err error)
	GetAccountsTotalCount() (count int64, err error)
	GetAllAccounts() (accounts []*table.Account, err error)
	GetLatestAccountIndex() (accountIndex int64, err error)
	GetConfirmedAccounts() (accounts []*table.Account, err error)
}

var singletonValue *account
var once sync.Once
var c config.Config

func New(c config.Config) AccountModel {
	once.Do(func() {
		gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
		if err != nil {
			logx.Errorf("gorm connect db error, err = %s", err.Error())
		}
		redisConn := redis.New(c.CacheRedis[0].Host, func(p *redis.Redis) {
			p.Type = c.CacheRedis[0].Type
			p.Pass = c.CacheRedis[0].Pass
		})
		singletonValue = &account{
			table:     `account`,
			db:        gormPointer,
			redisConn: redisConn,
			cache:     multcache.NewGoCache(context.Background(), 100, 10),
		}
	})
	return singletonValue
}
