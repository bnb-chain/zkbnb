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
	"errors"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func GetLatestNftInfoForRead(
	nftModel NftModel,
	mempoolTxModel MempoolModel,
	redisConnection *Redis,
	nftIndex int64,
) (
	nftInfo *NftInfo,
	err error,
) {
	key := util.GetNftKeyForRead(nftIndex)
	nftInfoStr, err := redisConnection.Get(key)
	if err != nil {
		logx.Errorf("[GetLatestNftInfoForRead] unable to get data from redis: %s", err.Error())
		return nil, err
	}
	var (
		dbNftInfo *nft.L2Nft
	)
	if nftInfoStr == "" {
		// get latest info from liquidity table
		dbNftInfo, err = nftModel.GetNftAsset(nftIndex)
		if err != nil {
			logx.Errorf("[GetLatestNftInfoForRead] unable to get latest nft by nft index: %s", err.Error())
			return nil, err
		}
		mempoolTxs, err := mempoolTxModel.GetPendingNftTxs()
		if err != nil {
			if err != mempool.ErrNotFound {
				logx.Errorf("[GetLatestAccountInfo] unable to get mempool txs by account index: %s", err.Error())
				return nil, err
			}
		}
		nftInfo = commonAsset.ConstructNftInfo(
			nftIndex,
			dbNftInfo.CreatorAccountIndex,
			dbNftInfo.OwnerAccountIndex,
			dbNftInfo.NftContentHash,
			dbNftInfo.NftL1TokenId,
			dbNftInfo.NftL1Address,
			dbNftInfo.CreatorTreasuryRate,
			dbNftInfo.CollectionId,
		)
		for _, mempoolTx := range mempoolTxs {
			for _, txDetail := range mempoolTx.MempoolDetails {
				if txDetail.AssetType != commonAsset.NftAssetType {
					continue
				}
				nBalance, err := commonAsset.ComputeNewBalance(commonAsset.NftAssetType, nftInfo.String(), txDetail.BalanceDelta)
				if err != nil {
					logx.Errorf("[GetLatestAccountInfo] unable to compute new balance: %s", err.Error())
					return nil, err
				}
				nftInfo, err = commonAsset.ParseNftInfo(nBalance)
				if err != nil {
					logx.Errorf("[GetLatestAccountInfo] unable to parse nft info: %s", err.Error())
					return nil, err
				}
			}
		}
		// write into cache
		lockKey := util.GetLockKey(key)
		redisLock := redis.NewRedisLock(redisConnection, lockKey)
		redisLock.SetExpire(5)
		isAcquired, err := redisLock.Acquire()
		if err != nil {
			logx.Errorf("[GetLatestNftInfoForRead] unable to acquire lock: %s", err.Error())
			return nil, err
		}
		if !isAcquired {
			logx.Errorf("[GetLatestNftInfoForRead] the lock has been used")
			return nil, errors.New("[GetLatestNftInfoForRead] the lock has been used")
		}
		infoBytes, err := json.Marshal(nftInfo)
		if err != nil {
			logx.Errorf("[GetLatestNftInfoForRead] unable to marshal: %s", err.Error())
			return nil, err
		}
		_ = redisConnection.Setex(key, string(infoBytes), NftExpiryTime)
	} else {
		err = json.Unmarshal([]byte(nftInfoStr), &nftInfo)
		if err != nil {
			logx.Errorf("[GetLatestNftInfoForRead] unable to unmarshal nft info: %s", err.Error())
			return nil, err
		}
	}
	return nftInfo, nil
}
