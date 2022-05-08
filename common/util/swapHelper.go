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
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

func ComputeLpPortion(
	PairATotal *big.Int,
	PairBTotal *big.Int,
	LpAmount *big.Int,
) (
	PairAAmount *big.Int, PairBAmount *big.Int, err error,
) {
	var (
		lpTotal   *big.Int
		lpPortion *big.Float
	)

	lpTotal = ffmath.Multiply(PairATotal, PairBTotal)
	//lpTotal = math.Sqrt(float64(PairATotal) * float64(PairBTotal))
	if lpTotal.Cmp(ZeroBigInt) == 0 {
		return big.NewInt(0), big.NewInt(0), nil
	}
	lpPortion = ffmath.FloatDivByInt(LpAmount, lpTotal)
	PairAAmount, err = CleanPackedAmount(ffmath.FloatToInt(ffmath.FloatMul(lpPortion, ffmath.IntToFloat(PairATotal))))
	if err != nil {
		logx.Errorf("[ComputeLpPortion] unable to compute clean packed amount: %s", err.Error())
		return nil, nil, err
	}
	PairBAmount, err = CleanPackedAmount(ffmath.FloatToInt(ffmath.FloatMul(lpPortion, ffmath.IntToFloat(PairBTotal))))
	if err != nil {
		logx.Errorf("[ComputeLpPortion] unable to compute clean packed amount: %s", err.Error())
		return nil, nil, err
	}
	return PairAAmount, PairBAmount, nil
}

/*
	ComputeDeltaX:
	(x-deltaX)(y+deltaY) = k
	deltaX = x - k/(y+deltaY)
*/
func ComputeDeltaX(x *big.Int, y *big.Int, deltaY *big.Int) (*big.Int, error) {
	k := ffmath.Multiply(x, y)
	yAddDeltaY := ffmath.Add(y, deltaY)
	kDivYAddDeltaY := ffmath.FloatDivByInt(k, yAddDeltaY)
	delatX, err := CleanPackedAmount(ffmath.Sub(x, ffmath.FloatToInt(kDivYAddDeltaY)))
	if err != nil {
		logx.Errorf("[ComputeDeltaX] unable to compute delta x: %s", err.Error())
		return nil, err
	}
	return delatX, nil
}

/*
	ComputeDeltaXInverse:
	(x+deltaX)(y-deltaY) = k
	deltaX = k/(y-deltaY) - x
*/
func ComputeDeltaXInverse(assetAAmount *big.Int, assetBAmount *big.Int, deltaY *big.Int) (*big.Int, error) {
	k := ffmath.Multiply(assetAAmount, assetBAmount)
	ySubDeltaY := ffmath.Sub(assetBAmount, deltaY)
	rate := ffmath.FloatDivByInt(k, ySubDeltaY)
	delatX, err := CleanPackedAmount(ffmath.Sub(ffmath.FloatToInt(rate), assetAAmount))
	if err != nil {
		logx.Errorf("[ComputeDeltaXInverse] unable to compute delta x: %s", err.Error())
		return nil, err
	}
	return delatX, nil
}

/*
	ComputeDeltaY:
	(x+deltaX)(y-deltaY) = k
	deltaY = y - k/(x+deltaX)
*/
func ComputeDeltaY(assetAAmount *big.Int, assetBAmount *big.Int, deltaX *big.Int) (*big.Int, error) {
	k := ffmath.Multiply(assetAAmount, assetBAmount)
	xAddDeltaX := ffmath.Add(assetAAmount, deltaX)
	if xAddDeltaX.Cmp(ZeroBigInt) == 0 {
		return big.NewInt(0), nil
	} else {
		rate := ffmath.FloatDivByInt(k, xAddDeltaX)
		deltaY, err := CleanPackedAmount(ffmath.Sub(assetBAmount, ffmath.FloatToInt(rate)))
		if err != nil {
			logx.Errorf("[ComputeDeltaY] unable to compute delta x: %s", err.Error())
			return nil, err
		}
		return deltaY, nil
	}
}

/*
	ComputeDeltaY:
	(x-deltaX)(y+deltaY) = k
	deltaY = k/(x-deltaX) - y
*/
func ComputeDeltaYInverse(assetAAmount *big.Int, assetBAmount *big.Int, deltaX *big.Int) (*big.Int, error) {
	k := ffmath.Multiply(assetAAmount, assetBAmount)
	//xSubDeltaX := assetAAmount - deltaX
	xSubDeltaX := ffmath.Sub(assetAAmount, deltaX)
	if xSubDeltaX.Cmp(ZeroBigInt) == 0 {
		return ZeroBigInt, nil
	} else {
		rate := ffmath.FloatDivByInt(k, xSubDeltaX)
		deltaY, err := CleanPackedAmount(ffmath.Sub(ffmath.FloatToInt(rate), assetBAmount))
		if err != nil {
			logx.Errorf("[ComputeDeltaYInverse] unable to compute delta x: %s", err.Error())
			return nil, err
		}
		return deltaY, nil
	}
}

func ComputeDelta(
	assetAAmount *big.Int,
	assetBAmount *big.Int,
	assetAId int64, assetBId int64, assetId int64, isFrom bool,
	deltaAmount *big.Int,
	feeRate int64,
) (assetAmount *big.Int, toAssetId int64, err error) {

	if isFrom {
		nDeltaAmount := ComputeDeltaWithFeeRate(deltaAmount, feeRate)
		if assetAId == assetId {
			delta, err := ComputeDeltaY(assetAAmount, assetBAmount, nDeltaAmount)
			if err != nil {
				logx.Errorf("[ComputeDelta] unable to compute delta Y: %s", err.Error())
				return nil, assetBId, err
			}
			return delta, assetBId, nil
		} else if assetBId == assetId {
			delta, err := ComputeDeltaX(assetAAmount, assetBAmount, nDeltaAmount)
			if err != nil {
				logx.Errorf("[ComputeDelta] unable to compute delta X: %s", err.Error())
				return nil, assetBId, err
			}
			return delta, assetAId, nil
		} else {
			logx.Errorf("[ComputeDelta] invalid asset id")
			return ZeroBigInt, 0, errors.New("[ComputeDelta]: invalid asset id")
		}
	} else {
		if assetAId == assetId {
			delta, err := ComputeDeltaYInverse(assetAAmount, assetBAmount, deltaAmount)
			if err != nil {
				logx.Errorf("[ComputeDelta] unable to ComputeDeltaYInverse: %s", err.Error())
				return nil, assetBId, err
			}
			amount, err := ComputeRealDeltaXWithFeeRate(delta, feeRate)
			if err != nil {
				logx.Errorf("[ComputeDelta] unable to ComputeRealDeltaXWithFeeRate: %s", err.Error())
				return nil, assetBId, err
			}
			return amount, assetBId, nil
		} else if assetBId == assetId {
			delta, err := ComputeDeltaXInverse(assetAAmount, assetBAmount, deltaAmount)
			if err != nil {
				logx.Errorf("[ComputeDelta] unable to ComputeDeltaXInverse: %s", err.Error())
				return nil, assetBId, err
			}
			amount, err := ComputeRealDeltaXWithFeeRate(delta, feeRate)
			if err != nil {
				logx.Errorf("[ComputeDelta] unable to ComputeRealDeltaXWithFeeRate: %s", err.Error())
				return nil, assetBId, err
			}
			return amount, assetAId, nil
		} else {
			logx.Errorf("[ComputeDelta] invalid asset id")
			return ZeroBigInt, 0, errors.New("[utils.ComputeDelta]: invalid asset id")
		}
	}
}

// deltaX - gas = deltaX * (10000 - feeRate) / 10000
func ComputeDeltaWithFeeRate(iDelta *big.Int, feeRate int64) *big.Int {
	realADeltaBigInt := ffmath.Div(ffmath.Multiply(iDelta, big.NewInt(FeeRateBase-feeRate)), big.NewInt(int64(FeeRateBase)))
	return realADeltaBigInt
}

// realDeltaX = deltaX / (1 - feeRate / 10000)
func ComputeRealDeltaXWithFeeRate(deltaX *big.Int, feeRate int64) (realDeltaX *big.Int, err error) {

	realDeltaX, err = CleanPackedAmount(ffmath.FloatToInt(
		ffmath.FloatDiv(
			ffmath.IntToFloat(deltaX),
			new(big.Float).SetFloat64(float64(FeeRateBase-feeRate)/float64(FeeRateBase))),
	))
	if err != nil {
		return nil, err
	}
	return realDeltaX, nil
}
