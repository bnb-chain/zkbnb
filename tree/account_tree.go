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
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon"
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
) (
	accountTree bsmt.SparseMerkleTree, accountAssetTrees *AssetTreeCache, err error,
) {

	maxAccountIndex, err := accountHistoryModel.GetMaxAccountIndex(blockHeight)
	if err != nil && err != types.DbErrNotFound {
		logx.Errorf("unable to get maxAccountIndex")
		return nil, nil, err
	}
	logx.Infof("get maxAccountIndex end")
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
	logx.Infof("newBASSparseMerkleTree end")

	if ctx.IsLoad() {
		if blockHeight == 0 || maxAccountIndex == -1 {
			return accountTree, accountAssetTrees, nil
		}

		start := time.Now()
		logx.Infof("reloadAccountTree start")
		totalTask := 0
		resultChan := make(chan *treeUpdateResp, 1)
		defer close(resultChan)
		pool, err := ants.NewPool(100)
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
						fromAccountIndex, toAccountIndex, accountAssetTrees)
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
				logx.Severef("reloadAccountTreeFromRDB failed:%s", err.Error())
				panic("reloadAccountTreeFromRDB failed: " + err.Error())
			}
		}
		pendingAccountItem := make([]bsmt.Item, 0)
		for i := 0; i < totalTask; i++ {
			result := <-resultChan
			if result.err != nil {
				logx.Severef("reloadAccountTree failed:%s", result.err.Error())
				panic("reloadAccountTree failed: " + result.err.Error())
			}
			pendingAccountItem = append(pendingAccountItem, result.pendingAccountItem...)
		}
		newVersion := bsmt.Version(blockHeight)
		err = accountTree.MultiSetWithVersion(pendingAccountItem, newVersion)
		if err != nil {
			logx.Errorf("unable to set account to tree: %s", err.Error())
			return nil, nil, err
		}
		_, err = accountTree.CommitWithNewVersion(nil, &newVersion)
		if err != nil {
			logx.Errorf("unable to commit account tree: %s,newVersion:%s,tree.LatestVersion:%s", err.Error(), uint64(newVersion), uint64(accountTree.LatestVersion()))
			return nil, nil, err
		}
		logx.Infof("reloadAccountTree end. cost time %s", float64(time.Since(start).Milliseconds()))
		return accountTree, accountAssetTrees, nil
	}

	if ctx.IsOnlyQuery() {
		return accountTree, accountAssetTrees, nil
	}

	// It's not loading from RDB, need to check tree versionblock
	if accountTree.LatestVersion() > bsmt.Version(blockHeight) && !accountTree.IsEmpty() {
		logx.Infof("account tree version [%d] is higher than block, rollback to %d", accountTree.LatestVersion(), blockHeight)
		err := accountTree.Rollback(bsmt.Version(blockHeight))
		if err != nil {
			logx.Errorf("unable to rollback account tree: %s, version: %d", err.Error(), blockHeight)
			return nil, nil, err
		}
	}

	for _, accountIndex := range accountIndexList {
		asset := accountAssetTrees.Get(accountIndex)
		if asset.LatestVersion() > bsmt.Version(blockHeight) && !asset.IsEmpty() {
			logx.Infof("asset tree %d version [%d] is higher than block, rollback to %d", accountIndex, asset.LatestVersion(), blockHeight)
			err := asset.Rollback(bsmt.Version(blockHeight))
			if err != nil {
				logx.Errorf("unable to rollback asset [%d] tree: %s, version: %d", accountIndex, err.Error(), blockHeight)
				return nil, nil, err
			}
		}
	}
	return accountTree, accountAssetTrees, nil
}

func reloadAccountTreeFromRDB(
	accountModel account.AccountModel,
	accountHistoryModel account.AccountHistoryModel,
	blockHeight int64,
	fromAccountIndex, toAccountIndex int64,
	accountAssetTrees *AssetTreeCache,
) ([]bsmt.Item, error) {
	_, accountHistories, err := accountHistoryModel.GetValidAccounts(blockHeight,
		fromAccountIndex, toAccountIndex)
	if err != nil {
		logx.Errorf("unable to get all accountHistories")
		return nil, err
	}
	pendingAccountItem := make([]bsmt.Item, 0, len(accountHistories))
	if len(accountHistories) == 0 {
		return pendingAccountItem, nil
	}
	accountIndexList := make([]int64, 0, len(accountHistories))
	for _, accountHistory := range accountHistories {
		accountIndexList = append(accountIndexList, accountHistory.AccountIndex)
	}
	accountInfoList, err := accountModel.GetAccountByIndexes(accountIndexList)
	if err != nil {
		logx.Errorf("unable to get account by account index list: %s,accountIndexList:%s", err.Error(), accountIndexList)
		return nil, err
	}
	accountInfoDbMap := make(map[int64]*account.Account, 0)
	for _, accountInfo := range accountInfoList {
		accountInfoDbMap[accountInfo.AccountIndex] = accountInfo
	}
	var (
		accountInfoMap = make(map[int64]*account.Account)
	)

	for _, accountHistory := range accountHistories {
		if accountInfoMap[accountHistory.AccountIndex] == nil {
			accountInfo := accountInfoDbMap[accountHistory.AccountIndex]
			if accountInfo == nil {
				logx.Errorf("unable to get account by account index: %s,AccountIndex:%s", err.Error(), accountHistory.AccountIndex)
				return nil, err
			}
			accountInfoMap[accountHistory.AccountIndex] = &account.Account{
				AccountIndex:    accountInfo.AccountIndex,
				AccountName:     accountInfo.AccountName,
				PublicKey:       accountInfo.PublicKey,
				AccountNameHash: accountInfo.AccountNameHash,
				L1Address:       accountInfo.L1Address,
				Nonce:           types.EmptyNonce,
				CollectionNonce: types.EmptyCollectionNonce,
				Status:          account.AccountStatusConfirmed,
			}
		}
		if accountHistory.Nonce != types.EmptyNonce {
			accountInfoMap[accountHistory.AccountIndex].Nonce = accountHistory.Nonce
		}
		if accountHistory.CollectionNonce != types.EmptyCollectionNonce {
			accountInfoMap[accountHistory.AccountIndex].CollectionNonce = accountHistory.CollectionNonce
		}
		accountInfoMap[accountHistory.AccountIndex].AssetInfo = accountHistory.AssetInfo
		accountInfoMap[accountHistory.AccountIndex].AssetRoot = accountHistory.AssetRoot
	}
	// get related account info
	for _, accountHistory := range accountHistories {
		accountIndex := accountHistory.AccountIndex
		if accountInfoMap[accountIndex] == nil {
			logx.Errorf("invalid account index")
			return nil, errors.New("invalid account index")
		}
		oAccountInfo := accountInfoMap[accountIndex]
		accountInfo, err := chain.ToFormatAccountInfo(oAccountInfo)
		if err != nil {
			logx.Errorf("unable to convert to format account info: %s", err.Error())
			return nil, err
		}
		// create account assets node
		pendingUpdateAssetItem := make([]bsmt.Item, 0, len(accountInfo.AssetInfo))
		for assetId, assetInfo := range accountInfo.AssetInfo {
			hashVal, err := AssetToNode(
				assetInfo.Balance.String(),
				assetInfo.OfferCanceledOrFinalized.String(),
			)
			if err != nil {
				logx.Errorf("unable to convert asset to node: %s", err.Error())
				return nil, err
			}
			pendingUpdateAssetItem = append(pendingUpdateAssetItem, bsmt.Item{Key: uint64(assetId), Val: hashVal})
		}
		newVersion := bsmt.Version(blockHeight)
		err = accountAssetTrees.Get(accountIndex).MultiSetWithVersion(pendingUpdateAssetItem, newVersion)
		if err != nil {
			logx.Errorf("unable to set asset to tree: %s", err.Error())
			return nil, err
		}
		_, err = accountAssetTrees.Get(accountIndex).CommitWithNewVersion(nil, &newVersion)
		if err != nil {
			logx.Errorf("unable to CommitWithNewVersion asset to tree: %s,newVersion:%s,tree.LatestVersion:%s", err.Error(), uint64(newVersion), uint64(accountAssetTrees.Get(accountIndex).LatestVersion()))
			return nil, err
		}
		accountHashVal, err := AccountToNode(
			accountInfoMap[accountIndex].AccountNameHash,
			accountInfoMap[accountIndex].PublicKey,
			accountInfoMap[accountIndex].Nonce,
			accountInfoMap[accountIndex].CollectionNonce,
			accountAssetTrees.Get(accountIndex).Root(),
		)
		if err != nil {
			logx.Errorf("unable to convert account to node: %s", err.Error())
			return nil, err
		}
		pendingAccountItem = append(pendingAccountItem, bsmt.Item{Key: uint64(accountIndex), Val: accountHashVal})
	}
	return pendingAccountItem, nil
}

func AssetToNode(balance string, offerCanceledOrFinalized string) (hashVal []byte, err error) {
	hashVal, err = ComputeAccountAssetLeafHash(balance, offerCanceledOrFinalized)
	if err != nil {
		logx.Errorf("unable to compute asset leaf hash: %s", err.Error())
		return nil, err
	}

	return hashVal, nil
}

func AccountToNode(
	accountNameHash string,
	publicKey string,
	nonce int64,
	collectionNonce int64,
	assetRoot []byte,
) (hashVal []byte, err error) {
	hashVal, err = ComputeAccountLeafHash(
		accountNameHash,
		publicKey,
		nonce,
		collectionNonce,
		assetRoot)
	if err != nil {
		logx.Errorf("unable to compute account leaf hash: %s", err.Error())
		return nil, err
	}

	return hashVal, nil
}

func NewMemAccountAssetTree() (tree bsmt.SparseMerkleTree, err error) {
	return bsmt.NewBNBSparseMerkleTree(bsmt.NewHasherPool(func() hash.Hash { return poseidon.NewPoseidon() }),
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
