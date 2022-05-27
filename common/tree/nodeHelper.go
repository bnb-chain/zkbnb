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

package tree

import (
	"github.com/zecrey-labs/zecrey-crypto/accumulators/merkleTree"
	"github.com/zeromicro/go-zero/core/logx"
)

func AssetToNode(balance string, lpAmount string, offerCanceledOrFinalized string) (node *Node, err error) {
	hashVal, err := ComputeAccountAssetLeafHash(balance, lpAmount, offerCanceledOrFinalized)
	if err != nil {
		logx.Errorf("[AccountToNode] unable to compute asset leaf hash: %s", err.Error())
		return nil, err
	}
	node = merkleTree.CreateLeafNode(hashVal)
	return node, nil
}

func LiquidityAssetToNode(
	assetAId int64,
	assetA string,
	assetBId int64,
	assetB string,
	lpAmount string,
	kLast string,
	feeRate int64,
	treasuryAccountIndex int64,
	treasuryFeeRate int64,
) (node *Node, err error) {
	hashVal, err := ComputeLiquidityAssetLeafHash(
		assetAId, assetA,
		assetBId, assetB,
		lpAmount,
		kLast,
		feeRate,
		treasuryAccountIndex,
		treasuryFeeRate,
	)
	if err != nil {
		logx.Errorf("[AccountToNode] unable to compute liquidity asset leaf hash: %s", err.Error())
		return nil, err
	}
	node = merkleTree.CreateLeafNode(hashVal)
	return node, nil
}

func NftAssetToNode(nftAsset *AccountL2NftHistory) (node *Node, err error) {
	hashVal, err := ComputeNftAssetLeafHash(
		nftAsset.CreatorAccountIndex, nftAsset.OwnerAccountIndex,
		nftAsset.NftContentHash,
		nftAsset.NftL1Address, nftAsset.NftL1TokenId,
		nftAsset.CreatorTreasuryRate,
	)
	if err != nil {
		logx.Errorf("[NftAssetToNode] unable to compute nft asset leaf hash: %s", err.Error())
		return nil, err
	}
	node = merkleTree.CreateLeafNode(hashVal)
	return node, nil
}

func AccountToNode(
	accountNameHash string,
	publicKey string,
	nonce int64,
	collectionNonce int64,
	assetRoot []byte,
) (node *Node, err error) {
	hashVal, err := ComputeAccountLeafHash(
		accountNameHash,
		publicKey,
		nonce,
		collectionNonce,
		assetRoot)
	if err != nil {
		logx.Errorf("[AccountToNode] unable to compute account leaf hash: %s", err.Error())
		return nil, err
	}
	node = merkleTree.CreateLeafNode(hashVal)
	return node, nil
}
