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
	"errors"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

func ComputeEmptyLpAmount(
	assetAAmount *big.Int,
	assetBAmount *big.Int,
) (lpAmount *big.Int, err error) {
	lpSquare := ffmath.Multiply(assetAAmount, assetBAmount)
	lpFloat := ffmath.FloatSqrt(ffmath.IntToFloat(lpSquare))
	lpAmount, err = CleanPackedAmount(ffmath.FloatToInt(lpFloat))
	if err != nil {
		logx.Errorf("[ComputeEmptyLpAmount] unable to compute lp amount: %s", err.Error())
		return nil, err
	}
	return lpAmount, nil
}

func ComputeLpAmount(
	liquidityInfo *commonAsset.LiquidityInfo,
	assetAAmount *big.Int,
) (lpAmount *big.Int, err error) {
	// lp = assetAAmount / poolA * LpAmount
	sLp := commonAsset.ComputeSLp(liquidityInfo.AssetA, liquidityInfo.AssetB, liquidityInfo.KLast, liquidityInfo.FeeRate, liquidityInfo.TreasuryRate)
	poolLpAmount := ffmath.Sub(liquidityInfo.LpAmount, sLp)
	lpAmount, err = CleanPackedAmount(ffmath.Div(ffmath.Multiply(assetAAmount, poolLpAmount), liquidityInfo.AssetA))
	if err != nil {
		return nil, err
	}
	return lpAmount, nil
}

func ComputeRemoveLiquidityAmount(
	liquidityInfo *commonAsset.LiquidityInfo,
	lpAmount *big.Int,
) (assetAAmount, assetBAmount *big.Int) {
	sLp := commonAsset.ComputeSLp(
		liquidityInfo.AssetA,
		liquidityInfo.AssetB,
		liquidityInfo.KLast,
		liquidityInfo.FeeRate,
		liquidityInfo.TreasuryRate,
	)
	poolLp := ffmath.Sub(liquidityInfo.LpAmount, sLp)
	assetAAmount = ffmath.Multiply(lpAmount, liquidityInfo.AssetA)
	assetAAmount = ffmath.Div(assetAAmount, poolLp)
	assetBAmount = ffmath.Multiply(lpAmount, liquidityInfo.AssetB)
	assetBAmount = ffmath.Div(assetBAmount, poolLp)
	return assetAAmount, assetBAmount
}

func ComputeDelta(
	assetAAmount *big.Int,
	assetBAmount *big.Int,
	assetAId int64, assetBId int64, assetId int64, isFrom bool,
	deltaAmount *big.Int,
	feeRate int64,
) (assetAmount *big.Int, toAssetId int64, err error) {

	if isFrom {
		if assetAId == assetId {
			delta := ComputeInputPrice(assetAAmount, assetBAmount, deltaAmount, feeRate)
			return delta, assetBId, nil
		} else if assetBId == assetId {
			delta := ComputeInputPrice(assetBAmount, assetAAmount, deltaAmount, feeRate)
			return delta, assetAId, nil
		} else {
			logx.Errorf("[ComputeDelta] invalid asset id")
			return ZeroBigInt, 0, errors.New("[ComputeDelta]: invalid asset id")
		}
	} else {
		if assetAId == assetId {
			delta := ComputeOutputPrice(assetAAmount, assetBAmount, deltaAmount, feeRate)
			return delta, assetBId, nil
		} else if assetBId == assetId {
			delta := ComputeOutputPrice(assetBAmount, assetAAmount, deltaAmount, feeRate)
			return delta, assetAId, nil
		} else {
			logx.Errorf("[ComputeDelta] invalid asset id")
			return ZeroBigInt, 0, errors.New("[ComputeDelta]: invalid asset id")
		}
	}
}

/*
	Implementation Reference:
	https://github.com/runtimeverification/verified-smart-contracts/blob/master/uniswap/x-y-k.pdf
*/

/*
	InputPrice = （9970 * deltaX * y) / (10000 * x + 9970 * deltaX)
*/
func ComputeInputPrice(x *big.Int, y *big.Int, inputX *big.Int, feeRate int64) *big.Int {
	rFeeR := big.NewInt(FeeRateBase - feeRate)
	return ffmath.Div(ffmath.Multiply(rFeeR, ffmath.Multiply(inputX, y)), ffmath.Add(ffmath.Multiply(big.NewInt(FeeRateBase), x), ffmath.Multiply(rFeeR, inputX)))
}

/*
	OutputPrice = （10000 * x * deltaY) / (9970 * (y - deltaY)) + 1
*/
func ComputeOutputPrice(x *big.Int, y *big.Int, inputY *big.Int, feeRate int64) *big.Int {
	rFeeR := big.NewInt(FeeRateBase - feeRate)
	return ffmath.Add(ffmath.Div(ffmath.Multiply(big.NewInt(FeeRateBase), ffmath.Multiply(x, inputY)), ffmath.Multiply(rFeeR, ffmath.Sub(y, inputY))), big.NewInt(1))
}
