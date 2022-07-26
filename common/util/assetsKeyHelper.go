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
 *
 */

package util

import (
	"strconv"
)

/*
	Func: GetAccountAssetUniqueKey
	Params: accountIndex int64, assetId int64
	Return: string
	Description: Generating the leaf index for accountAssetTree, and fetch account asset info from BalanceDeltaMap
				 Used for BalanceDelta Map.
*/
func GetAccountAssetUniqueKey(accountIndex int64, assetId int64) string {
	return AccountAssetPrefix + strconv.FormatInt(accountIndex, 10) + strconv.FormatInt(assetId, 10)
}

/*
	Func: GetAccountNftUniqueKey
	Params: accountIndex int64, nftAssetId int64
	Return: string
*/
func GetAccountNftUniqueKey(accountIndex int64, nftIndex int64) string {
	return AccountNftPrefix + strconv.FormatInt(accountIndex, 10) + strconv.FormatInt(nftIndex, 10)
}

/*
	Func: GetPoolLiquidityUniqueKey
	Params: accountIndex int64, pairIndex int64
	Return: string
	Description: Generating the leaf index for accountLiquidityTree
				 Used for LiquidityPoolDelta Map.
				 Account Index always equals to GasAccountIndex.
*/
func GetPoolLiquidityUniqueKey(accountIndex int64, pairIndex int64) string {
	return PoolLiquidityPrefix + strconv.FormatInt(accountIndex, 10) + strconv.FormatInt(pairIndex, 10)
}

/*
	Func: GetAccountLPUniqueKey
	Params: accountIndex int64, pairIndex int64
	Return: string
	Description: Generating the leaf index for accountLiquidityTree
				 Used for LPDelta Map.
*/
func GetAccountLPUniqueKey(accountIndex int64, pairIndex int64) string {
	return LpPrefix + strconv.FormatInt(accountIndex, 10) + strconv.FormatInt(pairIndex, 10)
}
