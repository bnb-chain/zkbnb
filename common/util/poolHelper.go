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
	"encoding/json"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

type PoolInfo struct {
	AssetAAmount *big.Int
	AssetBAmount *big.Int
}

func ConstructPoolInfo(a, b string) (info *PoolInfo, err error) {
	aInt, isValid := new(big.Int).SetString(a, Base)
	if !isValid {
		logx.Errorf("[ConstructPoolInfo] invalid big int")
		return nil, err
	}
	bInt, isValid := new(big.Int).SetString(b, Base)
	if !isValid {
		logx.Errorf("[ConstructPoolInfo] invalid big int")
		return nil, err
	}
	return &PoolInfo{
		AssetAAmount: aInt,
		AssetBAmount: bInt,
	}, nil
}

func (info *PoolInfo) String() string {
	infoBytes, _ := json.Marshal(info)
	return string(infoBytes)
}

func ParsePoolInfo(infoStr string) (info *PoolInfo, err error) {
	err = json.Unmarshal([]byte(infoStr), &info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func IsEqualPoolInfo(a, b *PoolInfo) bool {
	return a.String() == b.String()
}

func AddPoolInfoString(a, b string) (info string, err error) {
	aInfo, err := ParsePoolInfo(a)
	if err != nil {
		logx.Errorf("[AddPoolInfoString] unable to parse pool info: %s", err.Error())
		return "", err
	}
	bInfo, err := ParsePoolInfo(b)
	if err != nil {
		logx.Errorf("[AddPoolInfoString] unable to parse pool info: %s", err.Error())
		return "", err
	}
	updatedInfo := &PoolInfo{
		AssetAAmount: ffmath.Add(aInfo.AssetAAmount, bInfo.AssetAAmount),
		AssetBAmount: ffmath.Add(aInfo.AssetBAmount, bInfo.AssetBAmount),
	}
	return updatedInfo.String(), nil
}
