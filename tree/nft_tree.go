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
	"context"
	"fmt"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/log"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"time"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/dao/nft"
)

func InitNftTree(
	l2NftModel nft.L2NftModel,
	nftHistoryModel nft.L2NftHistoryModel,
	blockHeight int64,
	ctx *Context,
	fromHistory bool,
) (
	nftTree bsmt.SparseMerkleTree, err error,
) {
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, blockHeight)
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
		var maxNftIndex int64
		if fromHistory {
			maxNftIndex, err = nftHistoryModel.GetMaxNftIndex(blockHeight)
			if err != nil && err != types.DbErrNotFound {
				logx.WithContext(ctxLog).Errorf("unable to get latest nft assets: %s", err.Error())
				return nil, err
			}
		} else {
			maxNftIndex, err = l2NftModel.GetMaxNftIndex()
			if err != nil && err != types.DbErrNotFound {
				logx.Errorf("unable to get latest nft assets: %s", err.Error())
				return nil, err
			}
		}
		newVersion := bsmt.Version(blockHeight)
		start := time.Now()
		logx.WithContext(ctxLog).Infof("reloadNftTree start")
		totalTask := 0
		resultChan := make(chan *treeUpdateResp, 1)
		defer close(resultChan)
		pool, err := ants.NewPool(100, ants.WithPanicHandler(func(p interface{}) {
			panic("worker exits from a panic")
		}))
		for i := 0; int64(i) <= maxNftIndex; i += ctx.BatchReloadSize() {
			toNftIndex := int64(i+ctx.BatchReloadSize()) - 1
			if toNftIndex > maxNftIndex {
				toNftIndex = maxNftIndex
			}
			totalTask++
			err := func(fromNftIndex int64, toNftIndex int64) error {
				return pool.Submit(func() {
					pendingAccountItem, err := loadNftTreeFromRDB(l2NftModel,
						nftHistoryModel, blockHeight, fromNftIndex, toNftIndex, fromHistory, ctxLog)
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
				return nil, fmt.Errorf("loadNftTreeFromRDB failed: %s", err.Error())
			}
		}
		pendingAccountItem := make([]bsmt.Item, 0)
		for i := 0; i < totalTask; i++ {
			result := <-resultChan
			if result.err != nil {
				return nil, fmt.Errorf("reloadNftTree failed: %s", err.Error())
			}
			pendingAccountItem = append(pendingAccountItem, result.pendingAccountItem...)
		}
		err = nftTree.MultiSetWithVersion(pendingAccountItem, bsmt.Version(blockHeight))
		if err != nil {
			logx.WithContext(ctxLog).Errorf("unable to write nft asset to tree: %s", err.Error())
			return nil, err
		}
		_, err = nftTree.CommitWithNewVersion(nil, &newVersion)
		if err != nil {
			logx.WithContext(ctxLog).Errorf("unable to commit nft tree: %s", err.Error())
			return nil, err
		}
		logx.Infof("reloadNftTree end. cost time %v", time.Since(start))
		return nftTree, nil
	}

	if ctx.IsOnlyQuery() {
		return nftTree, nil
	}

	// It's not loading from RDB, need to check tree version
	err = RollBackNftTree(blockHeight, nftTree)
	if err != nil {
		return nil, err
	}
	logx.WithContext(ctxLog).Infof("end to roll back nft tree,when initializing the nft tree")

	return nftTree, nil
}

func loadNftTreeFromRDB(
	l2NftModel nft.L2NftModel,
	nftHistoryModel nft.L2NftHistoryModel,
	blockHeight int64,
	fromNftIndex, toNftIndex int64,
	fromHistory bool,
	ctx context.Context,
) ([]bsmt.Item, error) {
	pendingAccountItem := make([]bsmt.Item, 0)
	var nftAssets []*nft.L2Nft
	var err error
	if fromHistory {
		nftAssets = make([]*nft.L2Nft, 0)
		_, nftHistories, err := nftHistoryModel.GetLatestNftsByBlockHeight(blockHeight,
			fromNftIndex, toNftIndex)
		if err != nil {
			logx.WithContext(ctx).Errorf("unable to get latest nft assets: %s", err.Error())
			return nil, err
		}
		if len(nftHistories) == 0 {
			return pendingAccountItem, nil
		}
		for _, nftHistory := range nftHistories {
			nftAsset := &nft.L2Nft{
				CreatorAccountIndex: nftHistory.CreatorAccountIndex,
				OwnerAccountIndex:   nftHistory.OwnerAccountIndex,
				NftContentHash:      nftHistory.NftContentHash,
				RoyaltyRate:         nftHistory.RoyaltyRate,
				CollectionId:        nftHistory.CollectionId,
				NftIndex:            nftHistory.NftIndex,
				L2BlockHeight:       nftHistory.L2BlockHeight,
				NftContentType:      nftHistory.NftContentType,
			}
			nftAssets = append(nftAssets, nftAsset)
		}
	} else {
		nftAssets, err = l2NftModel.GetByNftIndexRange(fromNftIndex, toNftIndex)
		if err != nil {
			logx.Errorf("unable to get latest nft assets: %s", err.Error())
			return nil, err
		}
	}
	for _, nftAsset := range nftAssets {
		ctx := log.UpdateCtxWithKV(ctx, log.NftIndexCtx, nftAsset.NftIndex)
		nftIndex := nftAsset.NftIndex
		hashVal, err := NftAssetToNode(nftAsset, ctx)
		if err != nil {
			logx.WithContext(ctx).Errorf("unable to convert nft asset to node: %s", err.Error())
			return nil, err
		}
		pendingAccountItem = append(pendingAccountItem, bsmt.Item{Key: uint64(nftIndex), Val: hashVal})
	}
	return pendingAccountItem, nil
}

func NftAssetToNode(nftAsset *nft.L2Nft, ctx context.Context) (hashVal []byte, err error) {
	hashVal, err = ComputeNftAssetLeafHash(
		nftAsset.CreatorAccountIndex,
		nftAsset.OwnerAccountIndex,
		nftAsset.NftContentHash,
		nftAsset.RoyaltyRate,
		nftAsset.CollectionId,
		ctx,
	)
	if err != nil {
		logx.WithContext(ctx).Errorf("unable to compute nft asset leaf hash: %s", err.Error())
		return nil, err
	}
	return hashVal, nil
}

func RollBackNftTree(treeHeight int64, nftTree bsmt.SparseMerkleTree) error {
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, treeHeight)
	logx.WithContext(ctxLog).Infof("check to rollback nft tree, latestVersion:%d,versions=%s,nftRoot:%s,rollback to height:%d", nftTree.LatestVersion(), common2.FormatVersion(nftTree.Versions()), common.Bytes2Hex(nftTree.Root()), treeHeight)

	if nftTree.LatestVersion() > bsmt.Version(treeHeight) {
		logx.WithContext(ctxLog).Infof("nft tree latestVersion:%d is higher than block, rollback to %d", nftTree.LatestVersion(), treeHeight)

		err := nftTree.Rollback(bsmt.Version(treeHeight))
		if err != nil {
			return fmt.Errorf("unable to rollback nft latestVersion:%d,err:%s", treeHeight, err.Error())
		}
		logx.WithContext(ctxLog).Infof("end to rollback nft tree, latestVersion:%d,versions=%s,nftRoot:%s,rollback to height:%d", nftTree.LatestVersion(), common2.FormatVersion(nftTree.Versions()), common.Bytes2Hex(nftTree.Root()), treeHeight)

		//check version,the account tree version cannot be greater than the block height
		if versionBeGreaterThanHeight(nftTree.LatestVersion(), bsmt.Version(treeHeight)) {
			return fmt.Errorf("call nftTree.Rollback successfully,but fail to rollback nftTree,latestVersion: %d,versions=%s", nftTree.LatestVersion(), common2.FormatVersion(nftTree.Versions()))
		}
	}
	return nil
}
