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
	"github.com/bnb-chain/zkbnb/types"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"time"

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
		if blockHeight == 0 {
			return nftTree, nil
		}
		newVersion := bsmt.Version(blockHeight)
		maxNftIndex, err := nftHistoryModel.GetMaxNftIndex(blockHeight)
		if err != nil && err != types.DbErrNotFound {
			logx.Errorf("unable to get latest nft assets: %s", err.Error())
			return nil, err
		}

		start := time.Now()
		logx.Infof("reloadNftTree start")
		totalTask := 0
		resultChan := make(chan *treeUpdateResp, 1)
		defer close(resultChan)
		pool, err := ants.NewPool(100)
		for i := 0; int64(i) <= maxNftIndex; i += ctx.BatchReloadSize() {
			toNftIndex := int64(i+ctx.BatchReloadSize()) - 1
			if toNftIndex > maxNftIndex {
				toNftIndex = maxNftIndex
			}
			totalTask++
			err := func(fromNftIndex int64, toNftIndex int64) error {
				return pool.Submit(func() {
					pendingAccountItem, err := loadNftTreeFromRDB(
						nftHistoryModel, blockHeight,
						fromNftIndex, toNftIndex)
					if err != nil {
						logx.Severef("loadNftTreeFromRDB failed:%s", err.Error())
						resultChan <- &treeUpdateResp{
							err: err,
						}
						return
					}
					resultChan <- &treeUpdateResp{
						pendingAccountItem: pendingAccountItem,
						err:                err,
					}
				})
			}(int64(i), toNftIndex)
			if err != nil {
				logx.Severef("loadNftTreeFromRDB failed:%s", err.Error())
				panic("loadNftTreeFromRDB failed: " + err.Error())
			}
		}
		pendingAccountItem := make([]bsmt.Item, 0)
		for i := 0; i < totalTask; i++ {
			result := <-resultChan
			if result.err != nil {
				logx.Severef("reloadNftTree failed:%s", result.err.Error())
				panic("reloadNftTree failed: " + result.err.Error())
			}
			pendingAccountItem = append(pendingAccountItem, result.pendingAccountItem...)
		}
		err = nftTree.MultiSetWithVersion(pendingAccountItem, bsmt.Version(blockHeight))
		if err != nil {
			logx.Errorf("unable to write nft asset to tree: %s", err.Error())
			return nil, err
		}
		_, err = nftTree.CommitWithNewVersion(nil, &newVersion)
		if err != nil {
			logx.Errorf("unable to commit nft tree: %s", err.Error())
			return nil, err
		}
		logx.Infof("reloadNftTree end. cost time %s", float64(time.Since(start).Milliseconds()))
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
	fromNftIndex, toNftIndex int64,
) ([]bsmt.Item, error) {
	_, nftAssets, err := nftHistoryModel.GetLatestNftsByBlockHeight(blockHeight,
		fromNftIndex, toNftIndex)
	if err != nil {
		logx.Errorf("unable to get latest nft assets: %s", err.Error())
		return nil, err
	}
	pendingAccountItem := make([]bsmt.Item, 0, len(nftAssets))
	if len(nftAssets) == 0 {
		return pendingAccountItem, nil
	}
	for _, nftAsset := range nftAssets {
		nftIndex := nftAsset.NftIndex
		hashVal, err := NftAssetToNode(nftAsset)
		if err != nil {
			logx.Errorf("unable to convert nft asset to node: %s", err.Error())
			return nil, err
		}
		pendingAccountItem = append(pendingAccountItem, bsmt.Item{Key: uint64(nftIndex), Val: hashVal})
	}
	return pendingAccountItem, nil
}

func NftAssetToNode(nftAsset *nft.L2NftHistory) (hashVal []byte, err error) {
	hashVal, err = ComputeNftAssetLeafHash(
		nftAsset.CreatorAccountIndex,
		nftAsset.OwnerAccountIndex,
		nftAsset.NftContentHash,
		nftAsset.CreatorTreasuryRate,
		nftAsset.CollectionId,
		nftAsset.NftIndex,
		nftAsset.L2BlockHeight,
	)
	if err != nil {
		logx.Errorf("unable to compute nft asset leaf hash: %s", err.Error())
		return nil, err
	}
	return hashVal, nil
}
