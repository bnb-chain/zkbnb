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
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/ethereum/go-ethereum/common"
	"github.com/panjf2000/ants/v2"
	"hash"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb-smt/database/memory"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/types"
)

type treeUpdateResp struct {
	pendingAccountItem []bsmt.Item
	err                error
}

func accountAssetNamespace(index int64) string {
	return AccountAssetPrefix + strconv.Itoa(int(index)) + ":"
}

func InitAccountTree(
	accountModel account.AccountModel,
	accountHistoryModel account.AccountHistoryModel,
	accountIndexList []int64,
	blockHeight int64,
	ctx *Context,
	assetCacheSize int,
	fromHistory bool,
) (
	accountTree bsmt.SparseMerkleTree, accountAssetTrees *AssetTreeCache, err error,
) {
	var maxAccountIndex int64
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, blockHeight)
	if fromHistory {
		maxAccountIndex, err = accountHistoryModel.GetMaxAccountIndex(blockHeight)
		if err != nil && err != types.DbErrNotFound {
			logx.WithContext(ctxLog).Errorf("unable to get maxAccountIndex")
			return nil, nil, err
		}
	} else {
		maxAccountIndex, err = accountModel.GetMaxAccountIndex()
		if err != nil && err != types.DbErrNotFound {
			logx.WithContext(ctxLog).Errorf("unable to get maxAccountIndex")
			return nil, nil, err
		}
	}

	logx.WithContext(ctxLog).Infof("get maxAccountIndex end")
	opts := ctx.Options(0)
	nilAccountAssetNodeHashes := NilAccountAssetNodeHashes(AssetTreeHeight, NilAccountAssetNodeHash, ctx.Hasher())

	// init account state trees
	accountAssetTrees = NewLazyTreeCache(assetCacheSize, maxAccountIndex, blockHeight, func(index, block int64) bsmt.SparseMerkleTree {
		tree, err := bsmt.NewSparseMerkleTree(ctx.Hasher(),
			SetNamespace(ctx, accountAssetNamespace(index)), AssetTreeHeight, nilAccountAssetNodeHashes,
			ctx.Options(0)...)
		if err != nil {
			logx.Severef("failed to create new tree by assets, %v", err)
			panic("failed to create new tree by assets, err:" + err.Error())
		}
		return tree
	})
	accountTree, err = bsmt.NewBNBSparseMerkleTree(ctx.Hasher(),
		SetNamespace(ctx, AccountPrefix), AccountTreeHeight, NilAccountNodeHash,
		opts...)
	if err != nil {
		logx.Errorf("unable to create new account tree: %s", err.Error())
		return nil, nil, err
	}
	logx.WithContext(ctxLog).Infof("newBASSparseMerkleTree end")

	if ctx.IsLoad() {
		if blockHeight == 0 || maxAccountIndex == -1 {
			return accountTree, accountAssetTrees, nil
		}

		start := time.Now()
		logx.WithContext(ctxLog).Infof("reloadAccountTree start")
		totalTask := 0
		resultChan := make(chan *treeUpdateResp, 1)
		defer close(resultChan)
		pool, err := ants.NewPool(100, ants.WithPanicHandler(func(p interface{}) {
			panic("worker exits from a panic")
		}))
		for i := 0; int64(i) <= maxAccountIndex; i += ctx.BatchReloadSize() {
			toAccountIndex := int64(i+ctx.BatchReloadSize()) - 1
			if toAccountIndex > maxAccountIndex {
				toAccountIndex = maxAccountIndex
			}
			totalTask++
			err := func(fromAccountIndex int64, toAccountIndex int64) error {
				return pool.Submit(func() {
					pendingAccountItem, err := reloadAccountTreeFromRDB(
						accountModel, accountHistoryModel, blockHeight,
						fromAccountIndex, toAccountIndex, accountAssetTrees, fromHistory, ctxLog)
					if err != nil {
						logx.Severef("reloadAccountTreeFromRDB failed:%s", err.Error())
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
			}(int64(i), toAccountIndex)
			if err != nil {
				return nil, nil, fmt.Errorf("reloadAccountTreeFromRDB failed: %s", err.Error())
			}
		}
		pendingAccountItem := make([]bsmt.Item, 0)
		for i := 0; i < totalTask; i++ {
			result := <-resultChan
			if result.err != nil {
				return nil, nil, fmt.Errorf("reloadAccountTree failed: %s", err.Error())
			}
			pendingAccountItem = append(pendingAccountItem, result.pendingAccountItem...)
		}
		newVersion := bsmt.Version(blockHeight)
		err = accountTree.MultiSetWithVersion(pendingAccountItem, newVersion)
		if err != nil {
			logx.WithContext(ctxLog).Errorf("unable to set account to tree: %s", err.Error())
			return nil, nil, err
		}
		_, err = accountTree.CommitWithNewVersion(nil, &newVersion)
		if err != nil {
			logx.WithContext(ctxLog).Errorf("unable to commit account tree: %s,newVersion:%d,tree.LatestVersion:%d", err.Error(), uint64(newVersion), uint64(accountTree.LatestVersion()))
			return nil, nil, err
		}
		logx.WithContext(ctxLog).Infof("reloadAccountTree end. cost time %v", time.Since(start))
		return accountTree, accountAssetTrees, nil
	}

	if ctx.IsOnlyQuery() {
		return accountTree, accountAssetTrees, nil
	}

	// It's not loading from RDB, need to check tree version block
	err = RollBackAssetTree(accountIndexList, blockHeight, accountAssetTrees)
	if err != nil {
		return nil, nil, err
	}
	logx.WithContext(ctxLog).Infof("end to roll back asset tree,when initializing the account tree")

	err = RollBackAccountTree(blockHeight, accountTree)
	if err != nil {
		return nil, nil, err
	}
	logx.WithContext(ctxLog).Infof("end to roll back account tree,when initializing the account tree")

	return accountTree, accountAssetTrees, nil
}

func reloadAccountTreeFromRDB(
	accountModel account.AccountModel,
	accountHistoryModel account.AccountHistoryModel,
	blockHeight int64,
	fromAccountIndex, toAccountIndex int64,
	accountAssetTrees *AssetTreeCache,
	fromHistory bool,
	ctx context.Context,
) ([]bsmt.Item, error) {
	pendingAccountItem := make([]bsmt.Item, 0)
	var accountInfoList []*account.Account
	var err error
	if fromHistory {
		_, accountHistories, err := accountHistoryModel.GetValidAccounts(blockHeight,
			fromAccountIndex, toAccountIndex)
		if err != nil {
			logx.WithContext(ctx).Errorf("unable to get all accountHistories")
			return nil, err
		}
		if len(accountHistories) == 0 {
			return pendingAccountItem, nil
		}
		accountIndexList := make([]int64, 0, len(accountHistories))
		for _, accountHistory := range accountHistories {
			accountIndexList = append(accountIndexList, accountHistory.AccountIndex)
		}
		accountInfoList, err = accountModel.GetAccountByIndexes(accountIndexList)
		if err != nil {
			logx.WithContext(ctx).Errorf("unable to get account by account index list: %s,accountIndexList:%d", err.Error(), accountIndexList)
			return nil, err
		}
		accountInfoDbMap := make(map[int64]*account.Account, 0)
		for _, accountInfo := range accountInfoList {
			accountInfoDbMap[accountInfo.AccountIndex] = accountInfo
		}
		for _, accountHistory := range accountHistories {
			accountInfo := accountInfoDbMap[accountHistory.AccountIndex]
			if accountInfo == nil {
				logx.WithContext(ctx).Errorf("unable to get account by account index: %s,AccountIndex:%d", err.Error(), accountHistory.AccountIndex)
				return nil, err
			}
			accountInfo.Nonce = accountHistory.Nonce
			accountInfo.CollectionNonce = accountHistory.CollectionNonce
			accountInfo.Status = accountHistory.Status
			accountInfo.AssetInfo = accountHistory.AssetInfo
			accountInfo.AssetRoot = accountHistory.AssetRoot
			accountInfo.L2BlockHeight = accountHistory.L2BlockHeight
			accountInfo.PublicKey = accountHistory.PublicKey
			accountInfo.L1Address = accountHistory.L1Address
		}
	} else {
		accountInfoList, err = accountModel.GetByAccountIndexRange(fromAccountIndex, toAccountIndex)
		if err != nil {
			logx.WithContext(ctx).Errorf("unable to get all accountHistories")
			return nil, err
		}
		if len(accountInfoList) == 0 {
			return pendingAccountItem, nil
		}
	}
	for _, oAccountInfo := range accountInfoList {
		ctx := log.UpdateCtxWithKV(ctx, log.AccountIndexCtx, oAccountInfo.AccountIndex)
		accountInfo, err := chain.ToFormatAccountInfo(oAccountInfo)
		if err != nil {
			logx.WithContext(ctx).Errorf("unable to convert to format account info: %s", err.Error())
			return nil, err
		}
		// create account assets node
		pendingUpdateAssetItem := make([]bsmt.Item, 0, len(accountInfo.AssetInfo))
		for assetId, assetInfo := range accountInfo.AssetInfo {
			ctx := log.UpdateCtxWithKV(ctx, log.AssetIdCtx, assetId)
			hashVal, err := AssetToNode(assetInfo.Balance.String(), assetInfo.OfferCanceledOrFinalized.String(), ctx)
			if err != nil {
				logx.WithContext(ctx).Errorf("unable to convert asset to node: %s", err.Error())
				return nil, err
			}
			pendingUpdateAssetItem = append(pendingUpdateAssetItem, bsmt.Item{Key: uint64(assetId), Val: hashVal})
		}
		newVersion := bsmt.Version(blockHeight)
		err = accountAssetTrees.Get(accountInfo.AccountIndex).MultiSetWithVersion(pendingUpdateAssetItem, newVersion)
		if err != nil {
			logx.WithContext(ctx).Errorf("unable to set asset to tree: %s", err.Error())
			return nil, err
		}
		_, err = accountAssetTrees.Get(accountInfo.AccountIndex).CommitWithNewVersion(nil, &newVersion)
		if err != nil {
			logx.WithContext(ctx).Errorf("unable to CommitWithNewVersion asset to tree: %s,newVersion:%d,tree.LatestVersion:%d", err.Error(), uint64(newVersion), uint64(accountAssetTrees.Get(accountInfo.AccountIndex).LatestVersion()))
			return nil, err
		}
		accountHashVal, err := AccountToNode(
			accountInfo.L1Address,
			accountInfo.PublicKey,
			accountInfo.Nonce,
			accountInfo.CollectionNonce,
			accountAssetTrees.Get(accountInfo.AccountIndex).Root(),
			ctx,
		)
		if err != nil {
			logx.WithContext(ctx).Errorf("unable to convert account to node: %s", err.Error())
			return nil, err
		}
		pendingAccountItem = append(pendingAccountItem, bsmt.Item{Key: uint64(accountInfo.AccountIndex), Val: accountHashVal})
	}
	return pendingAccountItem, nil
}

func AssetToNode(balance string, offerCanceledOrFinalized string,
	ctx context.Context) (hashVal []byte, err error) {
	hashVal, err = ComputeAccountAssetLeafHash(balance, offerCanceledOrFinalized, ctx)
	if err != nil {
		logx.WithContext(ctx).Errorf("unable to compute asset leaf hash: %s", err.Error())
		return nil, err
	}

	return hashVal, nil
}

func AccountToNode(
	l1Address string,
	publicKey string,
	nonce int64,
	collectionNonce int64,
	assetRoot []byte,
	ctx context.Context,
) (hashVal []byte, err error) {
	hashVal, err = ComputeAccountLeafHash(
		l1Address,
		publicKey,
		nonce,
		collectionNonce,
		assetRoot,
		ctx,
	)
	if err != nil {
		logx.WithContext(ctx).Errorf("unable to compute account leaf hash: %s", err.Error())
		return nil, err
	}

	return hashVal, nil
}

func NewMemAccountAssetTree() (tree bsmt.SparseMerkleTree, err error) {
	return bsmt.NewBNBSparseMerkleTree(bsmt.NewHasherPool(func() hash.Hash { return NewGMimc() }),
		memory.NewMemoryDB(), AssetTreeHeight, NilAccountAssetNodeHash)
}

func NilAccountAssetNodeHashes(maxDepth uint8, nilHash []byte, hasher *bsmt.Hasher) [][]byte {
	hashes := make([][]byte, maxDepth+1)
	hashes[maxDepth] = nilHash
	for i := 1; i <= int(maxDepth); i++ {
		nHash := hasher.Hash(nilHash, nilHash)
		hashes[maxDepth-uint8(i)] = nHash
		nilHash = nHash
	}
	return hashes
}

func CheckAssetRoot(accountIndexMap map[int64]bool, curHeight int64, assetTrees *AssetTreeCache, accountHistoryModel account.AccountHistoryModel) error {
	accountIndexSlice := make([]int64, 0)
	accountIndexLen := len(accountIndexMap)
	for accountIndex := range accountIndexMap {
		accountIndexLen--
		accountIndexSlice = append(accountIndexSlice, accountIndex)
		if len(accountIndexSlice) == 100 || accountIndexLen == 0 {
			_, accountHistoryList, err := accountHistoryModel.GetLatestAccountHistories(accountIndexSlice, curHeight)
			if err != nil && err != types.DbErrNotFound {
				return fmt.Errorf("get latest account histories failed: %s", err.Error())
			}
			for _, accountHistory := range accountHistoryList {
				asset := assetTrees.Get(accountHistory.AccountIndex)
				assetRoot := common.Bytes2Hex(asset.Root())
				if assetRoot != accountHistory.AssetRoot {
					return fmt.Errorf("check asset root error,accountIndex=%d,curHeight=%d,assetRoot=%s not equal accountHistory.AssetRoot=%s,asset.LatestVersion=%d,versions=%s", accountIndex, curHeight, assetRoot, accountHistory.AssetRoot, asset.LatestVersion(), common2.FormatVersion(asset.Versions()))
				}
				if asset.LatestVersion() > bsmt.Version(curHeight) {
					logx.Errorf("check asset root error,accountIndex=%d asset.LatestVersion=%d,versions=%s, curHeight=%d", accountIndex, asset.LatestVersion(), common2.FormatVersion(asset.Versions()), curHeight)
				}
			}
			accountIndexSlice = make([]int64, 0)
		}
	}
	return nil
}

func RollBackAssetTree(accountIndexList []int64, treeHeight int64, assetTrees *AssetTreeCache) error {
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, treeHeight)
	for _, accountIndex := range accountIndexList {
		asset := assetTrees.Get(accountIndex)
		ctxLog := log.UpdateCtxWithKV(ctxLog, log.AccountIndexCtx, accountIndex)
		assetRoot := common.Bytes2Hex(asset.Root())
		logx.WithContext(ctxLog).Infof("check to rollback asset tree, accountIndex:%d, latestVersion:%d,versions=%s,assetRoot:%s,rollback to height:%d", accountIndex, asset.LatestVersion(), common2.FormatVersion(asset.Versions()), assetRoot, treeHeight)
		if asset.LatestVersion() > bsmt.Version(treeHeight) {
			logx.WithContext(ctxLog).Infof("asset tree accountIndex:%d latestVersion:%d is higher than block, rollback to height:%d", accountIndex, asset.LatestVersion(), treeHeight)
			err := asset.Rollback(bsmt.Version(treeHeight))
			if err != nil {
				return fmt.Errorf("unable to rollback asset accountIndex:%d, latestVersion: %d,tree err: %s", accountIndex, asset.LatestVersion(), err.Error())
			}
			assetRoot := common.Bytes2Hex(asset.Root())
			logx.WithContext(ctxLog).Infof("end to rollback asset tree, accountIndex:%d, latestVersion:%d,versions=%s,assetRoot:%s,rollback to height:%d", accountIndex, asset.LatestVersion(), common2.FormatVersion(asset.Versions()), assetRoot, treeHeight)

			if asset.LatestVersion() > bsmt.Version(treeHeight) {
				return fmt.Errorf("call asset.Rollback successfully,but fail to rollback asset accountIndex:%d latestVersion: %d,versions=%s", accountIndex, asset.LatestVersion(), common2.FormatVersion(asset.Versions()))
			}
		}
	}
	return nil
}

func RollBackAccountTree(treeHeight int64, accountTree bsmt.SparseMerkleTree) error {
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, treeHeight)
	logx.WithContext(ctxLog).Infof("check to rollback account tree, latestVersion:%d,versions=%s,accountRoot:%s,rollback to height:%d", accountTree.LatestVersion(), common2.FormatVersion(accountTree.Versions()), common.Bytes2Hex(accountTree.Root()), treeHeight)
	if accountTree.LatestVersion() > bsmt.Version(treeHeight) {
		logx.WithContext(ctxLog).Infof("account tree latestVersion:%d is higher than block, rollback to %d", accountTree.LatestVersion(), treeHeight)
		err := accountTree.Rollback(bsmt.Version(treeHeight))
		if err != nil {
			return fmt.Errorf("unable to rollback account latestVersion:%d,err:%s", treeHeight, err.Error())
		}
		logx.WithContext(ctxLog).Infof("end to rollback account tree, latestVersion:%d,versions=%s,accountRoot:%s,rollback to height:%d", accountTree.LatestVersion(), common2.FormatVersion(accountTree.Versions()), common.Bytes2Hex(accountTree.Root()), treeHeight)

		if accountTree.LatestVersion() > bsmt.Version(treeHeight) {
			return fmt.Errorf("call accountTree.Rollback successfully,but fail to rollback accountTree,latestVersion: %d,versions=%s", accountTree.LatestVersion(), common2.FormatVersion(accountTree.Versions()))
		}
	}
	return nil
}

func CheckStateRoot(height int64, accountTree bsmt.SparseMerkleTree, nftTree bsmt.SparseMerkleTree, blockModel block.BlockModel) (err error) {
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, height)
	blockInfo, err := blockModel.GetBlockByHeightWithoutTx(height)
	if err != nil {
		return fmt.Errorf("failed to get block info by height=%d error=%v", height, err)
	}

	accountTreeRoot := accountTree.Root()
	nftTreeRoot := nftTree.Root()
	newStateRoot := ComputeStateRootHash(accountTreeRoot, nftTreeRoot)
	newStateRootStr := common.Bytes2Hex(newStateRoot)
	logx.WithContext(ctxLog).Infof("checkStateRoot account tree root=%s,nft tree root=%s,newStateRoot=%s,database StateRoot=%s", common.Bytes2Hex(accountTreeRoot), common.Bytes2Hex(nftTreeRoot), newStateRootStr, blockInfo.StateRoot)
	if newStateRootStr != blockInfo.StateRoot {
		return fmt.Errorf("checkStateRoot smt tree newStateRoot=%s not equal database StateRoot=%s,height:%d,AccountTree.Root=%s,NftTree.Root=%s", newStateRootStr, blockInfo.StateRoot, blockInfo.BlockHeight, common.Bytes2Hex(accountTreeRoot), common.Bytes2Hex(nftTreeRoot))
	}
	return nil
}
