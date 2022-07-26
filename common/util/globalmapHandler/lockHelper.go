/*
 * Copyright © 2021 Zkbas Protocol
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

package globalmapHandler

import (
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"time"
)

func GetRedisLockByKey(conn *redis.Redis, keyLock string) (redisLock *redis.RedisLock) {
	// get lock
	redisLock = redis.NewRedisLock(conn, keyLock)
	// set expiry time
	redisLock.SetExpire(LockExpiryTime)
	return redisLock
}

func TryAcquireLock(redisLock *redis.RedisLock) (err error) {
	// lock
	notUsed, err := redisLock.Acquire()
	if err != nil {
		logx.Errorf("[GetLatestAssetByLock] unable to acquire the lock:", err.Error())
		return err
	}
	// re-try for three times
	if !notUsed {
		ticker := time.NewTicker(RetryInterval)
		defer ticker.Stop()
		count := 0
		for {
			if count > MaxRetryTimes {
				logx.Errorf("[GetLatestAssetByLock] the lock has been used, re-try later")
				return errors.New("[GetLatestAssetByLock] the lock has been used, re-try later")
			}
			notUsed, err = redisLock.Acquire()
			if err != nil {
				logx.Errorf("[GetLatestAssetByLock] unable to acquire the lock:", err.Error())
				return err
			}
			if notUsed {
				break
			}
			count++
			<-ticker.C
		}
	}
	return nil
}

func ReleaseLock(redisLockMap map[string]*redis.RedisLock) {
	for _, redisLock := range redisLockMap {
		redisLock.Release()
	}
}
