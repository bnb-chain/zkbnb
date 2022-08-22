/*
 * Copyright © 2021 ZkBAS Protocol
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
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/util"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
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

func ComputeLpAmount(liquidityInfo *commonAsset.LiquidityInfo, assetAAmount *big.Int) (lpAmount *big.Int, err error) {
	// lp = assetAAmount / poolA * LpAmount
	sLp, err := ComputeSLp(liquidityInfo.AssetA, liquidityInfo.AssetB, liquidityInfo.KLast, liquidityInfo.FeeRate, liquidityInfo.TreasuryRate)
	if err != nil {
		return nil, err
	}
	poolLpAmount := ffmath.Sub(liquidityInfo.LpAmount, sLp)
	lpAmount, err = CleanPackedAmount(ffmath.Div(ffmath.Multiply(assetAAmount, poolLpAmount), liquidityInfo.AssetA))
	if err != nil {
		return nil, err
	}
	return lpAmount, nil
}

func ComputeRemoveLiquidityAmount(liquidityInfo *commonAsset.LiquidityInfo, lpAmount *big.Int) (assetAAmount, assetBAmount *big.Int, err error) {
	sLp, err := ComputeSLp(liquidityInfo.AssetA, liquidityInfo.AssetB, liquidityInfo.KLast, liquidityInfo.FeeRate, liquidityInfo.TreasuryRate)
	if err != nil {
		return nil, nil, err
	}
	lpAmount, err = CleanPackedAmount(lpAmount)
	if err != nil {
		return nil, nil, err
	}
	poolLp := ffmath.Sub(liquidityInfo.LpAmount, sLp)
	assetAAmount = ffmath.Multiply(lpAmount, liquidityInfo.AssetA)
	assetAAmount, _ = util.CleanPackedAmount(ffmath.Div(assetAAmount, poolLp))
	assetBAmount = ffmath.Multiply(lpAmount, liquidityInfo.AssetB)
	assetBAmount, _ = util.CleanPackedAmount(ffmath.Div(assetBAmount, poolLp))
	return assetAAmount, assetBAmount, nil
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
			delta, err := ComputeInputPrice(assetAAmount, assetBAmount, deltaAmount, feeRate)
			if err != nil {
				return nil, 0, err
			}
			return delta, assetBId, nil
		} else if assetBId == assetId {
			delta, err := ComputeInputPrice(assetBAmount, assetAAmount, deltaAmount, feeRate)
			if err != nil {
				return nil, 0, err
			}
			return delta, assetAId, nil
		} else {
			logx.Errorf("[ComputeDelta] invalid asset id")
			return ZeroBigInt, 0, errors.New("[ComputeDelta]: invalid asset id")
		}
	} else {
		if assetAId == assetId {
			delta, err := ComputeOutputPrice(assetAAmount, assetBAmount, deltaAmount, feeRate)
			if err != nil {
				return nil, 0, err
			}
			return delta, assetBId, nil
		} else if assetBId == assetId {
			delta, err := ComputeOutputPrice(assetBAmount, assetAAmount, deltaAmount, feeRate)
			if err != nil {
				return nil, 0, err
			}
			return delta, assetAId, nil
		} else {
			logx.Errorf("[ComputeDelta] invalid asset id")
			return ZeroBigInt, 0, errors.New("[ComputeDelta]: invalid asset id")
		}
	}
}

// InputPrice = （9970 * deltaX * y) / (10000 * x + 9970 * deltaX)
func ComputeInputPrice(x *big.Int, y *big.Int, inputX *big.Int, feeRate int64) (*big.Int, error) {
	rFeeR := big.NewInt(FeeRateBase - feeRate)
	res, err := util.CleanPackedAmount(ffmath.Div(ffmath.Multiply(rFeeR, ffmath.Multiply(inputX, y)), ffmath.Add(ffmath.Multiply(big.NewInt(FeeRateBase), x), ffmath.Multiply(rFeeR, inputX))))
	if err != nil {
		return nil, err
	}
	return res, nil
}

// OutputPrice = （10000 * x * deltaY) / (9970 * (y - deltaY)) + 1
func ComputeOutputPrice(x *big.Int, y *big.Int, inputY *big.Int, feeRate int64) (*big.Int, error) {
	rFeeR := big.NewInt(FeeRateBase - feeRate)
	res, err := CleanPackedAmount(ffmath.Add(ffmath.Div(ffmath.Multiply(big.NewInt(FeeRateBase), ffmath.Multiply(x, inputY)), ffmath.Multiply(rFeeR, ffmath.Sub(y, inputY))), big.NewInt(1)))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func ComputeSLp(poolA, poolB *big.Int, kLast *big.Int, feeRate, treasuryRate int64) (*big.Int, error) {
	kCurrent := ffmath.Multiply(poolA, poolB)
	if kCurrent.Cmp(ZeroBigInt) == 0 {
		return ZeroBigInt, nil
	}
	kCurrent.Sqrt(kCurrent)
	kLast.Sqrt(kLast)
	l := ffmath.Multiply(ffmath.Sub(kCurrent, kLast), big.NewInt(FeeRateBase))
	r := ffmath.Multiply(ffmath.Sub(ffmath.Multiply(big.NewInt(FeeRateBase), ffmath.Div(big.NewInt(feeRate), big.NewInt(treasuryRate))), big.NewInt(FeeRateBase)), kCurrent)
	r = ffmath.Add(r, ffmath.Multiply(big.NewInt(FeeRateBase), kLast))
	res, err := CleanPackedAmount(ffmath.Div(l, r))
	if err != nil {
		return nil, err
	}
	return res, nil
}
