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

package liquidityHandler

type LiquidityPairInfo struct {
}

type LiquidityAccountInfo struct {
	AccountIndex uint32
	AssetAId     uint32
	AssetAName   string
	AssetBId     uint32
	AssetBName   string
	PairIndex    uint32
	AssetAAmount int64
	AssetBAmount int64
	LpEnc        string
	CreatedAt    int64
}
