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

package tree

import (
	"github.com/zeromicro/go-zero/core/logx"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/dao/nft"
)

func InitNftTree(
	nftHistoryModel nft.L2NftHistoryModel,
	blockHeight int64,
	ctx *Context,
) (
	nftTree bsmt.SparseMerkleTree, err error,
) {
	nftTree, err = bsmt.NewBNBSparseMerkleTree(ctx.Hasher(),
		SetNamespace(ctx, NFTPrefix), NftTreeHeight, NilNftNodeHash,
		ctx.Options(0)...)
	if err != nil {
		logx.Errorf("unable to create tree from db: %s", err.Error())
		return nil, err
	}

	if ctx.IsLoad() {
		newVersion := bsmt.Version(blockHeight)
		nums, err := nftHistoryModel.GetLatestNftsCountByBlockHeight(blockHeight)
		if err != nil {
			logx.Errorf("unable to get latest nft assets: %s", err.Error())
			return nil, err
		}
		for i := 0; i < int(nums); i += ctx.BatchReloadSize() {
			err := loadNftTreeFromRDB(
				nftHistoryModel, blockHeight,
				i, i+ctx.BatchReloadSize(), nftTree)
			if err != nil {
				return nil, err
			}
		}
		_, err = nftTree.CommitWithNewVersion(nil, &newVersion)
		if err != nil {
			logx.Errorf("unable to commit nft tree: %s", err.Error())
			return nil, err
		}
		return nftTree, nil
	}

	if ctx.IsOnlyQuery() {
		return nftTree, nil
	}

	// It's not loading from RDB, need to check tree version
	if nftTree.LatestVersion() > bsmt.Version(blockHeight) && !nftTree.IsEmpty() {
		logx.Infof("nft tree version [%d] is higher than block, rollback to %d", nftTree.LatestVersion(), blockHeight)
		err := nftTree.Rollback(bsmt.Version(blockHeight))
		if err != nil {
			logx.Errorf("unable to rollback nft tree: %s, version: %d", err.Error(), blockHeight)
			return nil, err
		}
	}
	return nftTree, nil
}

func loadNftTreeFromRDB(
	nftHistoryModel nft.L2NftHistoryModel,
	blockHeight int64,
	offset, limit int,
	nftTree bsmt.SparseMerkleTree,
) error {
	_, nftAssets, err := nftHistoryModel.GetLatestNftsByBlockHeight(blockHeight,
		limit, offset)
	if err != nil {
		logx.Errorf("unable to get latest nft assets: %s", err.Error())
		return err
	}
	for _, nftAsset := range nftAssets {
		nftIndex := nftAsset.NftIndex
		hashVal, err := NftAssetToNode(nftAsset)
		if err != nil {
			logx.Errorf("unable to convert nft asset to node: %s", err.Error())
			return err
		}

		err = nftTree.SetWithVersion(uint64(nftIndex), hashVal, bsmt.Version(blockHeight))
		if err != nil {
			logx.Errorf("unable to write nft asset to tree: %s", err.Error())
			return err
		}
	}
	return nil
}

func NftAssetToNode(nftAsset *nft.L2NftHistory) (hashVal []byte, err error) {
	hashVal, err = ComputeNftAssetLeafHash(
		nftAsset.CreatorAccountIndex,
		nftAsset.OwnerAccountIndex,
		nftAsset.NftContentHash,
		nftAsset.CreatorTreasuryRate,
		nftAsset.CollectionId,
	)
	if err != nil {
		logx.Errorf("unable to compute nft asset leaf hash: %s", err.Error())
		return nil, err
	}
	return hashVal, nil
}
