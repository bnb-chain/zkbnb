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

package globalmapHandler

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
)

func GetLatestOfferIdForWrite(
	offerModel nft.OfferModel,
	redisConnection *Redis,
	accountIndex int64,
) (
	redisLock *RedisLock,
	nftIndex int64,
	err error,
) {
	key := util.GetOfferIdKeyForWrite(accountIndex)
	lockKey := util.GetLockKey(key)
	redisLock = GetRedisLockByKey(redisConnection, lockKey)
	err = TryAcquireLock(redisLock)
	if err != nil {
		logx.Errorf("[GetLatestOfferIdForWrite] unable to get lock: %s", err.Error())
		return nil, -1, err
	}
	lastOfferId, err := offerModel.GetLatestOfferId(accountIndex)
	if err != nil {
		redisLock.Release()
		logx.Errorf("[GetLatestOfferIdForWrite] unable to get latest nft index: %s", err.Error())
		return nil, -1, err
	}
	return redisLock, lastOfferId + 1, nil
}
