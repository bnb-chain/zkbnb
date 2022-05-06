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

import (
	"errors"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"math/big"
)

func ComputeDelta(
	poolAccountLiquidity *LiquidityAccountInfo, assetId uint32, isFrom bool, deltaAmount *big.Int,
) (*big.Int, error) {
	if isFrom {
		if poolAccountLiquidity.AssetAId == assetId {
			delta := util.ComputeDeltaY(poolAccountLiquidity.AssetAAmount, poolAccountLiquidity.AssetBAmount, deltaAmount)
			return delta, nil
		} else if poolAccountLiquidity.AssetBId == assetId {
			delta := util.ComputeDeltaX(poolAccountLiquidity.AssetAAmount, poolAccountLiquidity.AssetBAmount, deltaAmount)
			return delta, nil
		} else {
			return nil, errors.New("err: invalid asset id")
		}
	} else {
		if poolAccountLiquidity.AssetAId == assetId {
			delta := util.ComputeDeltaYInverse(poolAccountLiquidity.AssetAAmount, poolAccountLiquidity.AssetBAmount, deltaAmount)
			return delta, nil
		} else if poolAccountLiquidity.AssetBId == assetId {
			delta := util.ComputeDeltaXInverse(poolAccountLiquidity.AssetAAmount, poolAccountLiquidity.AssetBAmount, deltaAmount)
			return delta, nil
		} else {
			return nil, errors.New("err: invalid asset id")
		}
	}
}
