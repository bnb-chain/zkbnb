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

package tree

import (
	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/zkbas-crypto/hash/bn254/zmimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/pkg/treedb"
)

// TODO replace history as nft + nftHistory
func InitNftTree(
	nftHistoryModel L2NftHistoryModel,
	blockHeight int64,
	ctx *treedb.Context,
) (
	nftTree bsmt.SparseMerkleTree, err error,
) {
	nftTree, err = bsmt.NewBASSparseMerkleTree(bsmt.NewHasher(zmimc.Hmimc),
		treedb.SetNamespace(ctx, NFTPrefix), NftTreeHeight, NilNftNodeHash,
		ctx.Options(blockHeight)...)
	if err != nil {
		logx.Errorf("[InitNftTree] unable to create tree from db: %s", err.Error())
		return nil, err
	}

	if ctx.IsLoad() {
		_, nftAssets, err := nftHistoryModel.GetLatestNftAssetsByBlockHeight(blockHeight)
		if err != nil {
			logx.Errorf("[InitNftTree] unable to get latest nft assets: %s", err.Error())
			return nil, err
		}
		for _, nftAsset := range nftAssets {
			nftIndex := nftAsset.NftIndex
			hashVal, err := NftAssetToNode(nftAsset)
			if err != nil {
				logx.Errorf("[InitNftTree] unable to convert nft asset to node: %s", err.Error())
				return nil, err
			}

			err = nftTree.Set(uint64(nftIndex), hashVal)
			if err != nil {
				logx.Errorf("[InitNftTree] unable to write nft asset to tree: %s", err.Error())
				return nil, err
			}
			_, err = nftTree.Commit(nil)
			if err != nil {
				logx.Errorf("[InitNftTree] unable to commit nft tree: %s", err.Error())
				return nil, err
			}
		}
	} else if nftTree.LatestVersion() > bsmt.Version(blockHeight) && !nftTree.IsEmpty() {
		err := nftTree.Rollback(bsmt.Version(blockHeight))
		logx.Infof("[InitNftTree] nft tree version [%d] if higher than block, rollback to %d", nftTree.LatestVersion(), blockHeight)
		if err != nil {
			logx.Errorf("[InitNftTree] unable to rollback nft tree: %s, version: %d", err.Error(), blockHeight)
			return nil, err
		}
	}
	return nftTree, nil
}

func NftAssetToNode(nftAsset *AccountL2NftHistory) (hashVal []byte, err error) {
	hashVal, err = ComputeNftAssetLeafHash(
		nftAsset.CreatorAccountIndex,
		nftAsset.OwnerAccountIndex,
		nftAsset.NftContentHash,
		nftAsset.NftL1Address,
		nftAsset.NftL1TokenId,
		nftAsset.CreatorTreasuryRate,
		nftAsset.CollectionId,
	)
	if err != nil {
		logx.Errorf("[NftAssetToNode] unable to compute nft asset leaf hash: %s", err.Error())
		return nil, err
	}
	return hashVal, nil
}
