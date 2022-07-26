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

package commonAsset

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"
)

type LiquidityInfo struct {
	PairIndex            int64
	AssetAId             int64
	AssetA               *big.Int
	AssetBId             int64
	AssetB               *big.Int
	LpAmount             *big.Int
	KLast                *big.Int
	FeeRate              int64
	TreasuryAccountIndex int64
	TreasuryRate         int64
}

func (info *LiquidityInfo) String() string {
	infoBytes, _ := json.Marshal(info)
	return string(infoBytes)
}

func EmptyLiquidityInfo(pairIndex int64) (info *LiquidityInfo) {
	return &LiquidityInfo{
		PairIndex:            pairIndex,
		AssetAId:             0,
		AssetA:               ZeroBigInt,
		AssetBId:             0,
		AssetB:               ZeroBigInt,
		LpAmount:             ZeroBigInt,
		KLast:                ZeroBigInt,
		FeeRate:              0,
		TreasuryAccountIndex: 0,
		TreasuryRate:         0,
	}
}

func ConstructLiquidityInfo(pairIndex int64, assetAId int64, assetAAmount string, assetBId int64, assetBAmount string,
	lpAmount string, kLast string, feeRate int64, treasuryAccountIndex int64, treasuryRate int64) (info *LiquidityInfo, err error) {
	assetA, isValid := new(big.Int).SetString(assetAAmount, 10)
	if !isValid {
		logx.Errorf("[ConstructLiquidityInfo] invalid big int")
		return nil, errors.New("[ConstructLiquidityInfo] invalid bit int")
	}
	assetB, isValid := new(big.Int).SetString(assetBAmount, 10)
	if !isValid {
		logx.Errorf("[ConstructLiquidityInfo] invalid big int")
		return nil, errors.New("[ConstructLiquidityInfo] invalid bit int")
	}
	lp, isValid := new(big.Int).SetString(lpAmount, 10)
	if !isValid {
		logx.Errorf("[ConstructLiquidityInfo] invalid big int")
		return nil, errors.New("[ConstructLiquidityInfo] invalid bit int")
	}
	kLastInt, isValid := new(big.Int).SetString(kLast, 10)
	if !isValid {
		logx.Errorf("[ConstructLiquidityInfo] invalid big int")
		return nil, errors.New("[ConstructLiquidityInfo] invalid bit int")
	}
	info = &LiquidityInfo{
		PairIndex:            pairIndex,
		AssetAId:             assetAId,
		AssetA:               assetA,
		AssetBId:             assetBId,
		AssetB:               assetB,
		LpAmount:             lp,
		KLast:                kLastInt,
		FeeRate:              feeRate,
		TreasuryAccountIndex: treasuryAccountIndex,
		TreasuryRate:         treasuryRate,
	}
	return info, nil
}

func ParseLiquidityInfo(infoStr string) (info *LiquidityInfo, err error) {
	err = json.Unmarshal([]byte(infoStr), &info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
