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
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
)

func GetLatestLiquidityInfoForWrite(
	liquidityModel LiquidityModel,
	mempoolTxDetailModel MempoolTxDetailModel,
	redisConnection *Redis,
	pairIndex int64,
) (
	redisLock *RedisLock,
	liquidityInfo *LiquidityInfo,
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
	var (
		dbLiquidityInfo *liquidity.Liquidity
	)
	if liquidityInfoStr == "" {
		// get latest info from liquidity table
		dbLiquidityInfo, err = liquidityModel.GetLiquidityByPairIndex(pairIndex)
		if err != nil {
			logx.Errorf("[GetLatestLiquidityInfoForRead] unable to get latest liquidity by pair index: %s", err.Error())
			return nil, nil, err
		}
		txDetails, err := mempoolTxDetailModel.GetMempoolTxDetailsByAssetIdAndAssetType(pairIndex, commonAsset.LiquidityAssetType)
		if err != nil {
			if err != mempool.ErrNotFound {
				logx.Errorf("[GetLatestAccountInfo] unable to get mempool txs by account index: %s", err.Error())
				return nil, nil, err
			}
		}
		liquidityInfo, err = commonAsset.ConstructLiquidityInfo(
			pairIndex,
			dbLiquidityInfo.AssetAId,
			dbLiquidityInfo.AssetA,
			dbLiquidityInfo.AssetBId,
			dbLiquidityInfo.AssetB,
			dbLiquidityInfo.LpAmount,
			dbLiquidityInfo.KLast,
			dbLiquidityInfo.FeeRate,
			dbLiquidityInfo.TreasuryAccountIndex,
			dbLiquidityInfo.TreasuryRate,
		)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] unable to construct pool info: %s", err.Error())
			return nil, nil, err
		}
		for _, txDetail := range txDetails {
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[GetLatestAccountInfo] unable to compute new balance: %s", err.Error())
				return nil, nil, err
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				logx.Errorf("[GetLatestAccountInfo] unable to parse pool info: %s", err.Error())
				return nil, nil, err
			}
		}
	} else {
		err = json.Unmarshal([]byte(liquidityInfoStr), &liquidityInfo)
		if err != nil {
			logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to unmarshal liquidity info: %s", err.Error())
			return nil, nil, err
		}
	}
	return redisLock, liquidityInfo, nil
}
