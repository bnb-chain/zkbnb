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
	"github.com/orcaman/concurrent-map"
	"sync"
)

const (
	LpPrefix            = "LP::"
	LockPrefix          = "Lock::"
	PoolLiquidityPrefix = "PoolLiquidity::"
	AccountAssetPrefix  = "AccountAsset::"

	L1AmountPrefix = "L1Amount::"

	LockNumber = 50

	AccountPrefix = "AccountIndex::"

	LockKeySuffix = "ByLock"

	LockExpiryTime = 10 // seconds
	RetryInterval  = 500
	MaxRetryTimes  = 3

	BalanceExpiryTime = 30 // seconds
)

var (
	GlobalMap = cmap.New()
)

var LogicMutexList [LockNumber]*sync.Mutex

func init() {
	for i := 0; i < LockNumber; i++ {
		LogicMutexList[i] = new(sync.Mutex)
	}
}
