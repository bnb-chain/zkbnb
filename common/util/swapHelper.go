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

package util

import (
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"math/big"
)

/*
	ComputeDeltaX:
	(x-deltaX)(y+deltaY) = k
	deltaX = x - k/(y+deltaY)
*/
func ComputeDeltaX(leftAssetBalance *big.Int, rightAssetBalance *big.Int, deltaY *big.Int) *big.Int {
	k := ffmath.Multiply(leftAssetBalance, rightAssetBalance)
	yPdeltaY := ffmath.Add(rightAssetBalance, deltaY)
	rate := ffmath.Div(k, yPdeltaY)
	delatX := ffmath.Sub(leftAssetBalance, rate)
	return delatX
}

/*
	ComputeDeltaXInverse:
	(x+deltaX)(y-deltaY) = k
	deltaX = k/(y-deltaY) - x
*/
func ComputeDeltaXInverse(leftAssetBalance *big.Int, rightAssetBalance *big.Int, deltaY *big.Int) *big.Int {
	k := ffmath.Multiply(leftAssetBalance, rightAssetBalance)
	yPdeltaY := ffmath.Sub(rightAssetBalance, deltaY)
	rate := ffmath.Div(k, yPdeltaY)
	delatX := ffmath.Sub(rate, leftAssetBalance)
	return delatX
}

/*
	ComputeDeltaY:
	(x+deltaX)(y-deltaY) = k
	deltaY = y - k/(x+deltaX)
*/
func ComputeDeltaY(leftAssetBalance *big.Int, rightAssetBalance *big.Int, deltaX *big.Int) *big.Int {
	k := ffmath.Multiply(leftAssetBalance, rightAssetBalance)
	xPdeltaX := ffmath.Add(leftAssetBalance, deltaX)
	rate := ffmath.Div(k, xPdeltaX)
	deltaY := ffmath.Sub(rightAssetBalance, rate)
	return deltaY
}

/*
	ComputeDeltaY:
	(x-deltaX)(y+deltaY) = k
	deltaY = k/(x-deltaX) - y
*/
func ComputeDeltaYInverse(leftAssetBalance *big.Int, rightAssetBalance *big.Int, deltaX *big.Int) *big.Int {
	k := ffmath.Multiply(leftAssetBalance, rightAssetBalance)
	xSdeltaX := ffmath.Sub(leftAssetBalance, deltaX)
	rate := ffmath.Div(k, xSdeltaX)
	deltaY := ffmath.Sub(rate, rightAssetBalance)
	return deltaY
}
