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
	"github.com/zecrey-labs/zecrey-crypto/hash/bn254/zmimc"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
)

// TODO replace history as nft + nftHistory
func InitNftTree(
	nftHistoryModel L2NftHistoryModel,
	blockHeight int64,
) (
	nftTree *Tree, err error,
) {
	_, nftAssets, err := nftHistoryModel.GetLatestNftAssetsByBlockHeight(blockHeight)
	if err != nil {
		logx.Errorf("[InitNftTree] unable to get latest nft assets: %s", err.Error())
		return nil, err
	}
	// empty tree
	if len(nftAssets) == 0 {
		nftTree, err = merkleTree.NewEmptyTree(NftTreeHeight, NilHash, zmimc.Hmimc)
		if err != nil {
			log.Println("[InitNftTree] unable to create empty tree:", err)
			return nil, err
		}
		return nftTree, nil
	}

	nftAssetsMap := make(map[int64]*Node)
	for _, nftAsset := range nftAssets {
		nftIndex := nftAsset.NftIndex
		node, err := NftAssetToNode(nftAsset)
		if err != nil {
			logx.Errorf("[InitNftTree] unable to convert nft asset to node: %s", err.Error())
			return nil, err
		}
		nftAssetsMap[nftIndex] = node
	}
	nftTree, err = merkleTree.NewTreeByMap(nftAssetsMap, NftTreeHeight, NilHash, zmimc.Hmimc)
	if err != nil {
		logx.Errorf("[InitNftTree] unable to create new tree by map")
		return nil, err
	}
	return nftTree, nil
}
