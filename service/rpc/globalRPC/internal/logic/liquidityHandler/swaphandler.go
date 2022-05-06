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

import "errors"

/*
	ComputeDeltaX:
	(x-deltaX)(y+deltaY) = k
	deltaX = x - k/(y+deltaY)
*/
func ComputeDeltaX(leftAssetBalance int64, rightAssetBalance int64, deltaY int64) int64 {
	k := leftAssetBalance * rightAssetBalance
	yPdeltaY := rightAssetBalance + deltaY
	rate := float64(k) / float64(yPdeltaY)
	delatX := float64(leftAssetBalance) - rate
	return int64(delatX)
}

/*
	ComputeDeltaXInverse:
	(x+deltaX)(y-deltaY) = k
	deltaX = k/(y-deltaY) - x
*/
func ComputeDeltaXInverse(leftAssetBalance int64, rightAssetBalance int64, deltaY int64) int64 {
	k := leftAssetBalance * rightAssetBalance
	yPdeltaY := rightAssetBalance - deltaY
	rate := float64(k) / float64(yPdeltaY)
	delatX := rate - float64(leftAssetBalance)
	return int64(delatX)
}

/*
	ComputeDeltaY:
	(x+deltaX)(y-deltaY) = k
	deltaY = y - k/(x+deltaX)
*/
func ComputeDeltaY(leftAssetBalance int64, rightAssetBalance int64, deltaX int64) int64 {
	k := leftAssetBalance * rightAssetBalance
	xPdeltaX := leftAssetBalance + deltaX
	rate := float64(k) / float64(xPdeltaX)
	deltaY := float64(rightAssetBalance) - rate
	return int64(deltaY)
}

/*
	ComputeDeltaY:
	(x-deltaX)(y+deltaY) = k
	deltaY = k/(x-deltaX) - y
*/
func ComputeDeltaYInverse(leftAssetBalance int64, rightAssetBalance int64, deltaX int64) int64 {
	k := leftAssetBalance * rightAssetBalance
	xSdeltaX := leftAssetBalance - deltaX
	rate := float64(k) / float64(xSdeltaX)
	deltaY := rate - float64(rightAssetBalance)
	return int64(deltaY)
}

func ComputeDelta(
	poolAccountLiquidity *LiquidityAccountInfo, assetId uint32, isFrom bool, deltaAmount int64,
) (int64, error) {
	if isFrom {
		if poolAccountLiquidity.AssetAId == assetId {
			delta := ComputeDeltaY(poolAccountLiquidity.AssetAAmount, poolAccountLiquidity.AssetBAmount, deltaAmount)
			return delta, nil
		} else if poolAccountLiquidity.AssetBId == assetId {
			delta := ComputeDeltaX(poolAccountLiquidity.AssetAAmount, poolAccountLiquidity.AssetBAmount, deltaAmount)
			return delta, nil
		} else {
			return 0, errors.New("err: invalid asset id")
		}
	} else {
		if poolAccountLiquidity.AssetAId == assetId {
			delta := ComputeDeltaYInverse(poolAccountLiquidity.AssetAAmount, poolAccountLiquidity.AssetBAmount, deltaAmount)
			return delta, nil
		} else if poolAccountLiquidity.AssetBId == assetId {
			delta := ComputeDeltaXInverse(poolAccountLiquidity.AssetAAmount, poolAccountLiquidity.AssetBAmount, deltaAmount)
			return delta, nil
		} else {
			return 0, errors.New("err: invalid asset id")
		}
	}
}
