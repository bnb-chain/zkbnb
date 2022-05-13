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
	"encoding/json"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
)

func GetLatestLiquidityInfoForWrite(
	liquidityModel LiquidityModel,
	liquidityHistoryModel LiquidityHistoryModel,
	redisConnection *Redis,
	pairIndex int64,
) (
	redisLock *RedisLock,
	liquidityInfo *liquidity.Liquidity,
	err error,
) {
	key := util.GetLiquidityKeyForWrite(pairIndex)
	lockKey := util.GetLockKey(key)
	redisLock = GetRedisLockByKey(redisConnection, lockKey)
	err = TryAcquireLock(redisLock)
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to get lock: %s", err.Error())
		return nil, nil, err
	}
	liquidityInfoStr, err := redisConnection.Get(key)
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to get data from redis: %s", err.Error())
		return nil, nil, err
	}
	if liquidityInfoStr == "" {
		// get latest info from history
		liquidityHistory, err := liquidityHistoryModel.GetLatestLiquidityByPairIndex(pairIndex)
		if err != nil {
			if err != liquidity.ErrNotFound {
				logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to get latest liquidity by pair index: %s", err.Error())
				return nil, nil, err
			} else {
				// get liquidity info from liquidity
				liquidityInfo, err = liquidityModel.GetLiquidityByPairIndex(pairIndex)
				if err != nil {
					logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to get liquidity info: %s", err.Error())
					return nil, nil, err
				}
			}
		} else {
			liquidityInfo = &liquidity.Liquidity{
				PairIndex: liquidityHistory.PairIndex,
				AssetAId:  liquidityHistory.AssetAId,
				AssetA:    liquidityHistory.AssetA,
				AssetBId:  liquidityHistory.AssetBId,
				AssetB:    liquidityHistory.AssetB,
			}
		}
	} else {
		var liquidityInfo *liquidity.Liquidity
		err := json.Unmarshal([]byte(liquidityInfoStr), &liquidityInfo)
		if err != nil {
			logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to unmarshal account info: %s", err.Error())
			return nil, nil, err
		}
	}
	return redisLock, liquidityInfo, nil
}
