/*
 * Copyright © 2021 ZkBNB Protocol
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

package types

import (
	"encoding/json"
)

type NftInfo struct {
	NftIndex            int64
	CreatorAccountIndex int64
	OwnerAccountIndex   int64
	NftContentHash      string
	NftL1TokenId        string
	NftL1Address        string
	CreatorTreasuryRate int64
	CollectionId        int64
}

func (info *NftInfo) String() string {
	infoBytes, _ := json.Marshal(info)
	return string(infoBytes)
}

func (info *NftInfo) IsEmptyNft() bool {
	if info.CreatorAccountIndex == EmptyAccountIndex &&
		info.OwnerAccountIndex == EmptyAccountIndex &&
		info.NftContentHash == EmptyNftContentHash &&
		info.NftL1TokenId == EmptyL1TokenId &&
		info.NftL1Address == EmptyL1Address &&
		info.CreatorTreasuryRate == EmptyCreatorTreasuryRate &&
		info.CollectionId == EmptyCollectionNonce {
		return true
	}
	return false
}

func EmptyNftInfo(nftIndex int64) (info *NftInfo) {
	return &NftInfo{
		NftIndex:            nftIndex,
		CreatorAccountIndex: EmptyAccountIndex,
		OwnerAccountIndex:   EmptyAccountIndex,
		NftContentHash:      EmptyNftContentHash,
		NftL1TokenId:        EmptyL1TokenId,
		NftL1Address:        EmptyL1Address,
		CreatorTreasuryRate: EmptyCreatorTreasuryRate,
		CollectionId:        EmptyCollectionNonce,
	}
}

func ConstructNftInfo(
	NftIndex int64,
	CreatorAccountIndex int64,
	OwnerAccountIndex int64,
	NftContentHash string,
	NftL1TokenId string,
	NftL1Address string,
	creatorTreasuryRate int64,
	collectionId int64,
) (nftInfo *NftInfo) {
	return &NftInfo{
		NftIndex:            NftIndex,
		CreatorAccountIndex: CreatorAccountIndex,
		OwnerAccountIndex:   OwnerAccountIndex,
		NftContentHash:      NftContentHash,
		NftL1TokenId:        NftL1TokenId,
		NftL1Address:        NftL1Address,
		CreatorTreasuryRate: creatorTreasuryRate,
		CollectionId:        collectionId,
	}
}
