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
 */

package globalmapHandler

import (
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type (
	AccountModel          = account.AccountModel
	AccountHistoryModel   = account.AccountHistoryModel
	MempoolModel          = mempool.MempoolModel
	MempoolTxDetailModel  = mempool.MempoolTxDetailModel
	LiquidityModel        = liquidity.LiquidityModel
	LiquidityHistoryModel = liquidity.LiquidityHistoryModel
	NftModel              = nft.L2NftModel
	Redis                 = redis.Redis
	RedisLock             = redis.RedisLock

	AccountInfo   = commonAsset.AccountInfo
	LiquidityInfo = commonAsset.LiquidityInfo
	NftInfo       = commonAsset.NftInfo
)

const (
	LpPrefix            = "LP::"
	LockPrefix          = "Lock::"
	PoolLiquidityPrefix = "PoolLiquidity::"
	AccountAssetPrefix  = "AccountAsset::"

	LockNumber = 50

	LockExpiryTime = 10 // seconds
	RetryInterval  = 500
	MaxRetryTimes  = 3

	AccountExpiryTime      = 30 // seconds
	LiquidityExpiryTime    = 30 // seconds
	NftExpiryTime          = 30 // seconds
	BasicAccountExpiryTime = 30 // seconds
)
