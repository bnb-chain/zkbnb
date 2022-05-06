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
	"errors"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/asset"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"strconv"
)

func GetLatestAccountInfoByLock(
	svcCtx *svc.ServiceContext,
	accountIndex int64,
	redisLockMap map[string]*redis.RedisLock,
) (
	accountInfo *account.Account,
	err error,
) {
	// get account info by account index
	accountHistory, err := svcCtx.AccountHistoryModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		errInfo := fmt.Sprintf("[GetLatestAccountInfoByLock] %s. invalid accountIndex %v",
			err.Error(), accountIndex)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	}
	// convert into Account
	accountInfo = &account.Account{
		AccountIndex:    accountHistory.AccountIndex,
		AccountName:     accountHistory.AccountName,
		PublicKey:       accountHistory.PublicKey,
		AccountNameHash: accountHistory.AccountNameHash,
		L1Address:       accountHistory.L1Address,
		Nonce:           accountHistory.Nonce,
	}
	// get latest nonce
	key := AccountPrefix + strconv.FormatInt(accountIndex, 10)
	lockKey := key + LockKeySuffix
	// get lock
	redisLock := GetRedisLockByKey(svcCtx.RedisConnection, lockKey)
	// try acquire lock
	err = TryAcquireLock(redisLock)
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByLock] unable to acquire lock: %s", err.Error())
		return nil, err
	}
	// get nonce from redis first
	nonceStr, err := svcCtx.RedisConnection.Get(key)
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByLock] unable to get from redis: %s", err.Error())
		// release lock
		redisLock.Release()
		return nil, err
	}
	if nonceStr != "" {
		accountInfo.Nonce, err = strconv.ParseInt(nonceStr, 10, 64)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfoByLock] unable to parse int: %s", err.Error())
			return nil, err
		}
	} else {
		// get latest nonce from mempool
		l2MempoolTx, err := svcCtx.MempoolModel.GetLatestL2MempoolTxByAccountIndex(accountIndex)
		if err != nil {
			if err != mempool.ErrNotFound {
				logx.Errorf("[GetLatestAccountInfoByLock] unable to get latest mempool tx: %s", err.Error())
				return nil, err
			} else {
				redisLockMap[lockKey] = redisLock
				return accountInfo, nil
			}
		}
		accountInfo.Nonce = l2MempoolTx.Nonce
	}

	// append it into redisLockMap for later release
	redisLockMap[lockKey] = redisLock
	return accountInfo, nil
}

func GetLatestAssetByLock(
	svcCtx *svc.ServiceContext,
	accountIndex int64,
	assetId int64,
	redisLockMap map[string]*redis.RedisLock,
) (assetInfo *asset.AccountAsset, err error) {
	// get asset info
	assetInfo = &asset.AccountAsset{
		AccountIndex: accountIndex,
		AssetId:      assetId,
		Balance:      "0",
	}
	// get latest account info by accountIndex and assetId
	key := util.GetAccountAssetGlobalKey(uint32(accountIndex), uint32(assetId))
	lockKey := key + LockKeySuffix
	// get lock
	redisLock := GetRedisLockByKey(svcCtx.RedisConnection, lockKey)
	// lock
	err = TryAcquireLock(redisLock)
	if err != nil {
		logx.Errorf("[GetLatestAssetByLock] unable to acquire lock: %s", err.Error())
		return nil, err
	}
	// get data from redis
	latestBalance, err := svcCtx.RedisConnection.Get(key)
	if err != nil {
		logx.Errorf("[GetLatestAssetByLock] unable to get balance from redis: %s", err.Error())
		// release lock
		redisLock.Release()
		return nil, err
	}
	if latestBalance != "" {
		assetInfo.Balance = latestBalance
	} else {
		// get accountAssetInfo by accountIndex and assetId
		resAccountSingleAsset, err := svcCtx.AssetHistoryModel.GetSingleAccountAssetHistory(accountIndex, assetId)
		if err != nil {
			if err != asset.ErrNotFound {
				errInfo := fmt.Sprintf("[GetLatestAssetByLock] %s. Invalid accountIndex/assetId %v/%v",
					err.Error(), accountIndex, assetId)
				logx.Error(errInfo)
				// release lock
				redisLock.Release()
				return nil, errors.New(errInfo)
			} else {
				// get data from asset table
				accountAssetInfo, err := svcCtx.AssetModel.GetSingleAccountAsset(accountIndex, assetId)
				if err != nil {
					if err != asset.ErrNotFound {
						errInfo := fmt.Sprintf("[GetLatestAssetByLock] %s. Invalid accountIndex/assetId %v/%v",
							err.Error(), accountIndex, assetId)
						logx.Error(errInfo)
						// release lock
						redisLock.Release()
						return nil, err
					}
				}
				assetInfo.Balance = accountAssetInfo.Balance
			}
		} else {
			assetInfo.Balance = resAccountSingleAsset.Balance
		}
		// fetch latest generalAssetType transaction
		mempoolDetail, err := svcCtx.MempoolDetailModel.GetLatestAccountAssetMempoolDetail(
			accountIndex,
			assetId,
			commonAsset.GeneralAssetType,
		)
		if err != nil {
			if err != mempool.ErrNotFound {
				errInfo := fmt.Sprintf("[GetLatestAssetByLock] %s",
					err.Error())
				logx.Error(errInfo)
				// release lock
				redisLock.Release()
				return nil, errors.New(errInfo)
			}
		} else {
			latestBalance, err = util.ComputeNewBalance(commonAsset.GeneralAssetType, mempoolDetail.Balance, mempoolDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[GetLatestAssetByLock] cannot compute new balance: %s", err.Error())
				// release lock
				redisLock.Release()
				return nil, err
			}
			assetInfo.Balance = latestBalance
		}
		redisLockMap[lockKey] = redisLock
	}

	return assetInfo, err
}
