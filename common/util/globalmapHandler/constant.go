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
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/model/nft"
)

type (
	NftModel  = nft.L2NftModel
	Redis     = redis.Redis
	RedisLock = redis.RedisLock

	LiquidityInfo = commonAsset.LiquidityInfo
)

const (
	LockExpiryTime = 10 // seconds
	RetryInterval  = 500
	MaxRetryTimes  = 3

	LiquidityExpiryTime = 30 // seconds
	NftExpiryTime       = 30 // seconds
)
