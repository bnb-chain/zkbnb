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
	"github.com/zecrey-labs/zecrey-core/common/general/model/liquidityPair"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func GetLatestAccountInfo(
	accountModel AccountModel,
	accountHistoryModel AccountHistoryModel,
	mempoolTxModel MempoolModel,
	mempoolTxDetailModel MempoolTxDetailModel,
	liquidityPairModel LiquidityPairModel,
	redisConnection *Redis,
	accountIndex int64,
) (
	accountInfo *FormatAccountInfo,
	err error,
) {
	key := util.GetAccountKey(accountIndex)
	accountInfoStr, err := redisConnection.Get(key)
	if err != nil {
		logx.Errorf("[GetLatestAccountInfo] unable to get data from redis: %s", err.Error())
		return nil, err
	}
	if accountInfoStr == "" {
		// get data from db
		oAccountInfo, err := accountModel.GetAccountByAccountIndex(accountIndex)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] unable to get account by account index: %s", err.Error())
			return nil, err
		}
		// get latest info from account history
		accountHistoryInfo, err := accountHistoryModel.GetLatestAccountInfoByAccountIndex(accountIndex)
		if err != nil {
			if err != account.ErrNotFound {
				logx.Errorf("[GetLatestAccountInfo] unable to get account info by account index from history table: %s", err.Error())
				return nil, err
			}
		} else {
			oAccountInfo.AssetInfo = accountHistoryInfo.AssetInfo
			oAccountInfo.AssetRoot = accountHistoryInfo.AssetRoot
			oAccountInfo.LiquidityInfo = accountHistoryInfo.LiquidityInfo
			oAccountInfo.LiquidityRoot = accountHistoryInfo.LiquidityRoot
			// get latest nonce
			latestNonce, err := accountHistoryModel.GetLatestAccountNonceByAccountIndex(accountIndex)
			if err != nil {
				if err != account.ErrNotFound {
					logx.Errorf("[GetLatestAccountInfo] unable to get latest nonce: %s", err.Error())
					return nil, err
				}
			} else {
				oAccountInfo.Nonce = latestNonce
			}
		}
		// convert to format account info
		accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] unable to convert to format account info: %s", err.Error())
			return nil, err
		}
		// compute latest nonce
		mempoolTxs, err := mempoolTxModel.GetMempoolTxsByAccountIndex(accountIndex)
		if err != nil {
			if err != mempool.ErrNotFound {
				logx.Errorf("[GetLatestAccountInfo] unable to get mempool txs by account index: %s", err.Error())
				return nil, err
			}
		}
		for _, mempoolTx := range mempoolTxs {
			if mempoolTx.Nonce != -1 {
				accountInfo.Nonce = mempoolTx.Nonce
			}
		}
		var (
			liquidityPairMap = make(map[int64]*liquidityPair.LiquidityPair)
		)
		mempoolTxDetails, err := mempoolTxDetailModel.GetAccountMempoolDetails(accountIndex)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] unable to get account mempool details: %s", err.Error())
			return nil, err
		}
		var (
			isHandledAsset = make(map[int64]map[int64]map[int64]bool)
		)
		for _, mempoolTxDetail := range mempoolTxDetails {
			if isHandledAsset[mempoolTxDetail.TxId] == nil {
				isHandledAsset[mempoolTxDetail.TxId] = make(map[int64]map[int64]bool)
			}
			if isHandledAsset[mempoolTxDetail.TxId][mempoolTxDetail.AccountIndex] == nil {
				isHandledAsset[mempoolTxDetail.TxId][mempoolTxDetail.AccountIndex] = make(map[int64]bool)
			}
			switch mempoolTxDetail.AssetType {
			case commonAsset.GeneralAssetType:
				// TODO maybe less than 0
				if accountInfo.AssetInfo[mempoolTxDetail.AssetId] == "" {
					accountInfo.AssetInfo[mempoolTxDetail.AssetId] = util.ZeroBigInt.String()
				}
				if !isHandledAsset[mempoolTxDetail.TxId][mempoolTxDetail.AccountIndex][mempoolTxDetail.AssetId] {
					accountInfo.AssetInfo[mempoolTxDetail.AssetId], err = util.ComputeNewBalance(
						commonAsset.GeneralAssetType,
						accountInfo.AssetInfo[mempoolTxDetail.AssetId],
						mempoolTxDetail.BalanceDelta,
					)
					if err != nil {
						logx.Errorf("[GetLatestAccountInfo] unable to compute new balance: %s", err.Error())
						return nil, err
					}
					isHandledAsset[mempoolTxDetail.TxId][mempoolTxDetail.AccountIndex][mempoolTxDetail.AssetId] = true
				}
				break
			case commonAsset.LiquidityAssetType:
				if accountInfo.LiquidityInfo[mempoolTxDetail.AssetId] == nil {
					// get pair info from liquidityPair
					if liquidityPairMap[mempoolTxDetail.AssetId] == nil {
						liquidityPairMap[mempoolTxDetail.AssetId], err = liquidityPairModel.GetLiquidityPairByIndex(mempoolTxDetail.AssetId)
						if err != nil {
							logx.Errorf("[GetLatestAccountInfo] cannot get liquidity pair by index: %s", err.Error())
							return nil, err
						}
					}
					accountInfo.LiquidityInfo[mempoolTxDetail.AssetId] = &commonAsset.Liquidity{
						PairIndex: mempoolTxDetail.AssetId,
						AssetAId:  liquidityPairMap[mempoolTxDetail.AssetId].AssetAId,
						AssetA:    util.ZeroBigInt.String(),
						AssetBId:  liquidityPairMap[mempoolTxDetail.AssetId].AssetBId,
						AssetB:    util.ZeroBigInt.String(),
						LpAmount:  util.ZeroBigInt.String(),
					}
				}
				poolInfo, err := util.ConstructPoolInfo(
					accountInfo.LiquidityInfo[mempoolTxDetail.AssetId].AssetA,
					accountInfo.LiquidityInfo[mempoolTxDetail.AssetId].AssetB,
				)
				if err != nil {
					logx.Errorf("[GetLatestAccountInfo] unable to construct pool info: %s", err.Error())
					return nil, err
				}
				// compute new balance
				nBalance, err := util.ComputeNewBalance(
					commonAsset.LiquidityAssetType, poolInfo.String(), mempoolTxDetail.BalanceDelta)
				if err != nil {
					logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
					return nil, err
				}
				newPoolInfo, err := util.ParsePoolInfo(nBalance)
				if err != nil {
					logx.Errorf("[CommitterTask] unable to parse pair info: %s", err.Error())
					return nil, err
				}
				accountInfo.LiquidityInfo[mempoolTxDetail.AssetId].AssetA =
					newPoolInfo.AssetAAmount.String()
				accountInfo.LiquidityInfo[mempoolTxDetail.AssetId].AssetB =
					newPoolInfo.AssetBAmount.String()
				break
			case commonAsset.LiquidityLpAssetType:
				if accountInfo.LiquidityInfo[mempoolTxDetail.AssetId] == nil {
					// get pair info from liquidityPair
					if liquidityPairMap[mempoolTxDetail.AssetId] == nil {
						liquidityPairMap[mempoolTxDetail.AssetId], err = liquidityPairModel.GetLiquidityPairByIndex(mempoolTxDetail.AssetId)
						if err != nil {
							logx.Errorf("[GetLatestAccountInfo] cannot get liquidity pair by index: %s", err.Error())
							return nil, err
						}
					}
					accountInfo.LiquidityInfo[mempoolTxDetail.AssetId] = &commonAsset.Liquidity{
						PairIndex: mempoolTxDetail.AssetId,
						AssetAId:  liquidityPairMap[mempoolTxDetail.AssetId].AssetAId,
						AssetA:    util.ZeroBigInt.String(),
						AssetBId:  liquidityPairMap[mempoolTxDetail.AssetId].AssetBId,
						AssetB:    util.ZeroBigInt.String(),
						LpAmount:  util.ZeroBigInt.String(),
					}
				}
				// compute new balance
				nBalance, err := util.ComputeNewBalance(
					commonAsset.LiquidityLpAssetType, accountInfo.LiquidityInfo[mempoolTxDetail.AssetId].LpAmount, mempoolTxDetail.BalanceDelta)
				if err != nil {
					logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
					return nil, err
				}
				accountInfo.LiquidityInfo[mempoolTxDetail.AssetId].LpAmount = nBalance
				break
			case commonAsset.NftAssetType:
				break
			default:
				logx.Errorf("[GetLatestAccountInfo] invalid asset type")
				return nil, errors.New("[GetLatestAccountInfo] invalid asset type")
			}
		}
		// write into cache
		lockKey := util.GetLockKey(key)
		redisLock := redis.NewRedisLock(redisConnection, lockKey)
		redisLock.SetExpire(5)
		isAcquired, err := redisLock.Acquire()
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] unable to acquire lock: %s", err.Error())
			return nil, err
		}
		if !isAcquired {
			logx.Errorf("[GetLatestAccountInfo] the lock has been used")
			return nil, errors.New("[GetLatestAccountInfo] the lock has been used")
		}
		// latest nonce
		accountInfo.Nonce = accountInfo.Nonce + 1
		info, err := commonAsset.FromFormatAccountInfo(accountInfo)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] unable to convert format account info to account info: %s", err.Error())
			return nil, err
		}
		infoBytes, err := json.Marshal(info)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] unable to marshal: %s", err.Error())
			return nil, err
		}
		_ = redisConnection.Setex(key, string(infoBytes), AccountExpiryTime)
	} else {
		var oAccountInfo *account.Account
		err := json.Unmarshal([]byte(accountInfoStr), &oAccountInfo)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] unable to unmarshal account info: %s", err.Error())
			return nil, err
		}
		accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] unable convert to format account info: %s", err.Error())
			return nil, err
		}
		// update cache
		_ = redisConnection.Expire(key, AccountExpiryTime)
	}
	return accountInfo, nil
}

func GetBasicAccountInfo(
	accountModel AccountModel,
	redisConnection *Redis,
	accountIndex int64,
) (
	accountInfo *FormatAccountInfo,
	err error,
) {
	key := util.GetBasicAccountKey(accountIndex)
	basicAccountInfoStr, err := redisConnection.Get(key)
	if err != nil {
		logx.Errorf("[GetBasicAccountInfo] unable to get account info: %s", err.Error())
		return nil, err
	}
	if basicAccountInfoStr == "" {
		oAccountInfo, err := accountModel.GetAccountByAccountIndex(accountIndex)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to get account by account index: %s", err.Error())
			return nil, err
		}
		accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to get basic account info: %s", err.Error())
			return nil, err
		}
		// update cache
		oAccountInfoBytes, err := json.Marshal(oAccountInfo)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to marshal account info: %s", err.Error())
			return nil, err
		}
		_ = redisConnection.Setex(key, string(oAccountInfoBytes), BasicAccountExpiryTime)
	} else {
		var oAccountInfo *account.Account
		err = json.Unmarshal([]byte(basicAccountInfoStr), &oAccountInfo)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to parse account info: %s", err.Error())
			return nil, err
		}
		accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to get basic account info: %s", err.Error())
			return nil, err
		}
		// update cache
		_ = redisConnection.Expire(key, BasicAccountExpiryTime)
	}
	return accountInfo, nil
}
