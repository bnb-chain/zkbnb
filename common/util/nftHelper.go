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
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

type NftInfo struct {
	NftIndex            int64
	CreatorAccountIndex int64
	OwnerAccountIndex   int64
	AssetId             int64
	AssetAmount         *big.Int
	NftContentHash      string
	NftL1TokenId        string
	NftL1Address        string
}

func (info *NftInfo) String() string {
	infoBytes, _ := json.Marshal(info)
	return string(infoBytes)
}

func ParseNftInfo(infoStr string) (info *NftInfo, err error) {
	err = json.Unmarshal([]byte(infoStr), &info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func EmptyNftInfo(nftIndex int64) (info *NftInfo) {
	return &NftInfo{
		NftIndex:            nftIndex,
		CreatorAccountIndex: -1,
		OwnerAccountIndex:   -1,
		AssetId:             -1,
		AssetAmount:         big.NewInt(0),
		// TODO zero hash
		NftContentHash: "0",
		NftL1TokenId:   "0",
		NftL1Address:   "0",
	}
}

func IsEmptyNftInfo(info *NftInfo) bool {
	if info.NftIndex != -1 || info.AssetId != -1 || info.AssetAmount.Cmp(big.NewInt(0)) != 0 ||
		info.NftContentHash != "" || info.NftL1TokenId != "0" || info.NftL1Address != "" {
		return false
	}
	return true
}

func ConstructNftInfo(
	NftIndex int64,
	CreatorAccountIndex int64,
	OwnerAccountIndex int64,
	AssetId int64,
	AssetAmount string,
	NftContentHash string,
	NftL1TokenId string,
	NftL1Address string,
) (nftInfo *NftInfo, err error) {
	assetAmount, isValid := new(big.Int).SetString(AssetAmount, Base)
	if !isValid {
		logx.Errorf("[ConstructNftInfo] invalid big int")
		return nil, errors.New("[ConstructNftInfo] invalid big int")
	}
	return &NftInfo{
		NftIndex:            NftIndex,
		CreatorAccountIndex: CreatorAccountIndex,
		OwnerAccountIndex:   OwnerAccountIndex,
		AssetId:             AssetId,
		AssetAmount:         assetAmount,
		NftContentHash:      NftContentHash,
		NftL1TokenId:        NftL1TokenId,
		NftL1Address:        NftL1Address,
	}, nil
}
