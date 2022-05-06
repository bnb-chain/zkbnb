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
 */

package globalmapHandler

import (
	"errors"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"strconv"
)

func ResetGlobalMap(connection *redis.Redis, redisLockMap map[string]*redis.RedisLock) (err error) {
	for key, _ := range redisLockMap {
		_, err = connection.Del(key)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
	Func: UpdateGlobalMap
	Params: nTx *mempool.MempoolTx
	Return: error
	Description: Update Global Map by new transaction.
				LockedAssetType globalMap must be initialized before UpdateGlobalMap otherwise it will throw panic.
*/
func UpdateGlobalMap(redisConnection *redis.Redis, nTx *mempool.MempoolTx, redisLockMap map[string]*redis.RedisLock) (err error) {
	for _, detail := range nTx.MempoolDetails {
		var (
			key      string
			nBalance string
		)
		switch detail.AssetType {
		case commonAsset.GeneralAssetType:
			key = util.GetAccountAssetUniqueKey(detail.AccountIndex, detail.AssetId)
			nBalance, err = util.ComputeNewBalance(detail.AssetType, detail.Balance, detail.BalanceDelta)
			if err != nil {
				err = ResetGlobalMap(redisConnection, redisLockMap)
				if err == nil {
					ReleaseLock(redisLockMap)
				}
				errInfo := fmt.Sprintf("[globalmapHandler.UpdateGlobalMap.GeneralAssetType] %s", err.Error())
				logx.Error(errInfo)
				return errors.New(errInfo)
			}
			break
		case commonAsset.LiquidityAssetType:
			key = util.GetPoolLiquidityUniqueKey(detail.AccountIndex, detail.AssetId)
			nBalance, err = util.ComputeNewBalance(detail.AssetType, detail.Balance, detail.BalanceDelta)
			if err != nil {
				err = ResetGlobalMap(redisConnection, redisLockMap)
				if err == nil {
					ReleaseLock(redisLockMap)
				}
				errInfo := fmt.Sprintf("[globalmapHandler.UpdateGlobalMap.LiquidityAssetType] %s", err.Error())
				logx.Error(errInfo)
				return errors.New(errInfo)
			}
			break
		case commonAsset.LiquidityLpAssetType:
			key = util.GetAccountLPUniqueKey(detail.AccountIndex, detail.AssetId)
			nBalance, err = util.ComputeNewBalance(detail.AssetType, detail.Balance, detail.BalanceDelta)
			if err != nil {
				err = ResetGlobalMap(redisConnection, redisLockMap)
				if err == nil {
					ReleaseLock(redisLockMap)
				}
				errInfo := fmt.Sprintf("[globalmapHandler.UpdateGlobalMap.LiquidityLpAssetType] %s", err.Error())
				logx.Error(errInfo)
				return errors.New(errInfo)
			}
			break
		case commonAsset.NftAssetType:
			key = util.GetAccountNftUniqueKey(detail.AccountIndex, detail.AssetId)
			nBalance, err = util.ComputeNewBalance(detail.AssetType, detail.Balance, detail.BalanceDelta)
			if err != nil {
				err = ResetGlobalMap(redisConnection, redisLockMap)
				if err == nil {
					ReleaseLock(redisLockMap)
				}
				errInfo := fmt.Sprintf("[globalmapHandler.UpdateGlobalMap.NftAssetType] %s", err.Error())
				logx.Error(errInfo)
				return errors.New(errInfo)
			}
			break
		}
		HandleGlobalMapUpdate(redisConnection, nTx, key, nBalance)
	}

	ReleaseLock(redisLockMap)
	return nil
}

/*
	Func: HandleGlobalMapUpdate
	Params: key string, nBalance string
	Return:
	Description: Update Global Map by key / new value
*/
func HandleGlobalMapUpdate(connection *redis.Redis, nTx *mempool.MempoolTx, key string, nBalance string) {
	// update nonce
	accountKey := util.GetAccountKey(nTx.AccountIndex)
	err := connection.Setex(accountKey, strconv.FormatInt(nTx.Nonce+1, 10), BalanceExpiryTime)
	if err != nil {
		connection.Del(accountKey)
	}
	// update balance
	err = connection.Setex(key, nBalance, BalanceExpiryTime)
	if err != nil {
		connection.Del(key)
	}
}

type GlobalAssetInfo struct {
	AccountIndex int64
	AssetId      int64
	AssetType    int64
	BaseBalance  string
}
