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

package commonAsset

import (
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
)

func ComputeSLp(
	poolA, poolB *big.Int, kLast *big.Int, feeRate, treasuryRate int64,
) *big.Int {
	kCurrent := ffmath.Multiply(poolA, poolB)
	if kCurrent.Cmp(ZeroBigInt) == 0 {
		return ZeroBigInt
	}
	kCurrent.Sqrt(kCurrent)
	kLast.Sqrt(kLast)
	l := ffmath.Multiply(ffmath.Sub(kCurrent, kLast), big.NewInt(RateBase))
	r := ffmath.Multiply(ffmath.Sub(ffmath.Multiply(big.NewInt(RateBase), ffmath.Div(big.NewInt(feeRate), big.NewInt(treasuryRate))), big.NewInt(RateBase)), kCurrent)
	r = ffmath.Add(r, ffmath.Multiply(big.NewInt(RateBase), kLast))
	return ffmath.Div(l, r)
}
