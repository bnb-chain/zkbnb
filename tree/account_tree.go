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
	"github.com/zeromicro/go-zero/core/logx"
	"hash"
	"strconv"
	"sync"
	"time"

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
) (
	accountTree bsmt.SparseMerkleTree, accountAssetTrees *AssetTreeCache, err error,
) {
	var maxAccountIndex int64
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, blockHeight)
	if ctx.fromHistory {
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
	accountAssetTrees = NewLazyTreeCache(ctx.assetCacheSize, maxAccountIndex, blockHeight, func(index, block int64) bsmt.SparseMerkleTree {
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
		return createAccountTree(blockHeight, maxAccountIndex, ctx.fromHistory, accountModel, accountHistoryModel, ctx, accountTree, accountAssetTrees, ctxLog)
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

func createAccountTree(blockHeight int64, maxAccountIndex int64, fromHistory bool, accountModel account.AccountModel,
	accountHistoryModel account.AccountHistoryModel, ctx *Context, accountTree bsmt.SparseMerkleTree, accountAssetTrees *AssetTreeCache, ctxLog context.Context) (
	bsmt.SparseMerkleTree, *AssetTreeCache, error,
) {
	if blockHeight == 0 || maxAccountIndex == -1 {
		return accountTree, accountAssetTrees, nil
	}

	start := time.Now()
	logx.WithContext(ctxLog).Infof("reloadAccountTree start")

	accountQueue := make(chan *account.Account, 100000)
	defer close(accountQueue)

	loadAccountComplete := make(chan bool, 1)
	defer close(loadAccountComplete)

	commitAssetSmtComplete := make(chan bool, 1)
	defer close(commitAssetSmtComplete)

	assetStateRootQueue := make(chan *treeUpdateResp, 100000)
	defer close(assetStateRootQueue)

	go func() {
		start := time.Now()
		err := loadAccounts(maxAccountIndex, ctx, fromHistory, blockHeight, accountModel, accountHistoryModel, accountQueue)
		if err != nil {
			logx.Severe(err)
			panic("loadAccounts error:" + err.Error())
		}
		loadAccountComplete <- true
		logx.WithContext(ctxLog).Infof("loadAccounts end. cost time %v", time.Since(start))
	}()

	go func() {
		start := time.Now()
		err := commitAssetSmt(blockHeight, ctx, accountQueue, assetStateRootQueue, accountAssetTrees, loadAccountComplete, ctxLog)
		if err != nil {
			logx.Severe(err)
			panic("commitAssetSmt error:" + err.Error())
		}
		commitAssetSmtComplete <- true
		logx.WithContext(ctxLog).Infof("commitAssetSmt end. cost time %v", time.Since(start))
	}()

	pendingAccountItem, err := buildPendingAccountItem(assetStateRootQueue, commitAssetSmtComplete, ctxLog)
	if err != nil {
		return nil, nil, err
	}

	logx.WithContext(ctxLog).Infof("wait loadAccounts and commitAssetSmt...")

	accountTreeStart := time.Now()
	logx.WithContext(ctxLog).Infof("start update account smt, account count=%d", len(pendingAccountItem))

	newVersion := bsmt.Version(blockHeight)
	err = accountTree.MultiSetWithVersion(pendingAccountItem, newVersion)
	if err != nil {
		logx.WithContext(ctxLog).Errorf("unable to set account to tree: %s", err.Error())
		return nil, nil, err
	}
	logx.WithContext(ctxLog).Infof("start accountTree CommitWithNewVersion")

	_, err = accountTree.CommitWithNewVersion(nil, &newVersion)
	if err != nil {
		logx.WithContext(ctxLog).Errorf("unable to commit account tree: %s,newVersion:%d,tree.LatestVersion:%d", err.Error(), uint64(newVersion), uint64(accountTree.LatestVersion()))
		return nil, nil, err
	}
	logx.WithContext(ctxLog).Infof("end update account smt. cost time %v", time.Since(accountTreeStart))

	logx.WithContext(ctxLog).Infof("reloadAccountTree end. cost time %v", time.Since(start))
	return accountTree, accountAssetTrees, nil
}

func loadAccounts(maxAccountIndex int64, ctx *Context, fromHistory bool, blockHeight int64, accountModel account.AccountModel,
	accountHistoryModel account.AccountHistoryModel, accountQueue chan *account.Account) error {

	resultChan := make(chan error, common2.MaxInt64(maxAccountIndex/int64(ctx.BatchReloadSize()), 1))
	defer close(resultChan)

	pool, err := ants.NewPool(ctx.dbRoutineSize, ants.WithPanicHandler(func(p interface{}) {
		panic("worker exits from a panic")
	}))

	totalTask := 0
	for i := 0; int64(i) <= maxAccountIndex; i += ctx.BatchReloadSize() {
		toAccountIndex := int64(i+ctx.BatchReloadSize()) - 1
		if toAccountIndex > maxAccountIndex {
			toAccountIndex = maxAccountIndex
		}
		totalTask++
		err := func(fromAccountIndex int64, toAccountIndex int64) error {
			return pool.Submit(func() {
				doLoadAccounts(fromAccountIndex, toAccountIndex, fromHistory, blockHeight, accountModel, accountHistoryModel, accountQueue, resultChan)
			})
		}(int64(i), toAccountIndex)

		if err != nil {
			return fmt.Errorf("reloadAccountTreeFromRDB failed: %s", err.Error())
		}
	}
	for i := 0; i < totalTask; i++ {
		result := <-resultChan
		if result != nil {
			return fmt.Errorf("reloadAccountTree failed: %s", err.Error())
		}
	}
	return nil
}

func doLoadAccounts(fromAccountIndex int64, toAccountIndex int64, fromHistory bool, blockHeight int64, accountModel account.AccountModel,
	accountHistoryModel account.AccountHistoryModel, accountQueue chan *account.Account, resultChan chan error) {
	var accountInfoList []*account.Account
	var err error
	if fromHistory {
		_, accountHistories, err := accountHistoryModel.GetValidAccounts(blockHeight,
			fromAccountIndex, toAccountIndex)
		if err != nil {
			resultChan <- fmt.Errorf("unable to get all accountHistories,fromAccountIndex=%d,toAccountIndex=%d,err=%s", fromAccountIndex, toAccountIndex, err.Error())
			return
		}
		if len(accountHistories) == 0 {
			resultChan <- nil
			return
		}

		accountInfoDbMap := make(map[int64]*account.Account, 0)
		for _, accountInfo := range accountInfoList {
			accountInfoDbMap[accountInfo.AccountIndex] = accountInfo
		}
		for _, accountHistory := range accountHistories {
			accountInfo := &account.Account{}
			accountInfo.Nonce = accountHistory.Nonce
			accountInfo.CollectionNonce = accountHistory.CollectionNonce
			accountInfo.Status = accountHistory.Status
			accountInfo.AssetInfo = accountHistory.AssetInfo
			accountInfo.AssetRoot = accountHistory.AssetRoot
			accountInfo.L2BlockHeight = accountHistory.L2BlockHeight
			accountInfo.PublicKey = accountHistory.PublicKey
			accountInfo.L1Address = accountHistory.L1Address
			accountInfo.AccountIndex = accountHistory.AccountIndex
			accountInfoList = append(accountInfoList, accountInfo)
		}
	} else {
		accountInfoList, err = accountModel.GetByAccountIndexRange(fromAccountIndex, toAccountIndex)
		if err != nil {
			resultChan <- fmt.Errorf("unable to get all GetByAccountIndexRange,fromAccountIndex=%d,toAccountIndex=%d,err=%s", fromAccountIndex, toAccountIndex, err.Error())
			return
		}
	}
	logx.Infof("add oAccountInfo to accountQueue,count=%d", len(accountQueue))
	for _, oAccountInfo := range accountInfoList {
		accountQueue <- oAccountInfo
	}
	resultChan <- nil
	return
}

func commitAssetSmt(blockHeight int64, ctx *Context, accountQueue chan *account.Account, assetStateRootQueue chan *treeUpdateResp, accountAssetTrees *AssetTreeCache, loadAccountComplete chan bool, ctxLog context.Context) error {
	pool, err := ants.NewPool(ctx.dbRoutineSize*20, ants.WithPanicHandler(func(p interface{}) {
		panic("worker exits from a panic")
	}))
	if err != nil {
		return fmt.Errorf("init ants.NewPool failed: %s", err.Error())
	}
	wg := sync.WaitGroup{}
	run := true

	for run {
		select {
		case oAccountInfo := <-accountQueue:
			wg.Add(1)
			err := func(oAccountInfo *account.Account, accountAssetTrees *AssetTreeCache, ctxLog context.Context) error {
				return pool.Submit(func() {
					defer wg.Done()
					doCommitAssetSmt(blockHeight, oAccountInfo, accountAssetTrees, assetStateRootQueue, ctxLog)
				})
			}(oAccountInfo, accountAssetTrees, ctxLog)
			if err != nil {
				return fmt.Errorf("doCommitAssetSmt failed: %s", err.Error())
			}
		default:
			if len(loadAccountComplete) == 1 && len(accountQueue) == 0 {
				logx.WithContext(ctxLog).Infof("no data in accountQueue")
				run = false
			}
		}
	}

	logx.WithContext(ctxLog).Infof("wait doCommitAssetSmt...")
	wg.Wait()
	return nil
}

func doCommitAssetSmt(blockHeight int64, oAccountInfo *account.Account, accountAssetTrees *AssetTreeCache, assetStateRootQueue chan *treeUpdateResp, ctxLog context.Context) {
	ctx := log.UpdateCtxWithKV(ctxLog, log.AccountIndexCtx, oAccountInfo.AccountIndex)
	accountInfo, err := chain.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		assetStateRootQueue <- &treeUpdateResp{
			err: fmt.Errorf("unable to convert to format account info: %s", err.Error()),
		}
		return
	}
	// create account assets node
	pendingUpdateAssetItem := make([]bsmt.Item, 0, len(accountInfo.AssetInfo))
	for assetId, assetInfo := range accountInfo.AssetInfo {
		ctx := log.UpdateCtxWithKV(ctx, log.AssetIdCtx, assetId)
		hashVal, err := AssetToNode(assetInfo.Balance.String(), assetInfo.OfferCanceledOrFinalized.String(), ctx)
		if err != nil {
			assetStateRootQueue <- &treeUpdateResp{
				err: fmt.Errorf("unable to convert asset to node: %s", err.Error()),
			}
			return
		}
		pendingUpdateAssetItem = append(pendingUpdateAssetItem, bsmt.Item{Key: uint64(assetId), Val: hashVal})
	}
	newVersion := bsmt.Version(blockHeight)

	setWithVersionStart := time.Now()
	err = accountAssetTrees.Get(accountInfo.AccountIndex).MultiSetWithVersion(pendingUpdateAssetItem, newVersion)
	logx.WithContext(ctxLog).Debugf("doCommitAssetSmt1 MultiSetWithVersion end. cost time %v", time.Since(setWithVersionStart))

	if err != nil {
		assetStateRootQueue <- &treeUpdateResp{
			err: fmt.Errorf("unable to set asset to tree: %s", err.Error()),
		}
		return
	}
	logx.WithContext(ctxLog).Debugf("doCommitAssetSmt MultiSetWithVersion end. cost time %v", time.Since(setWithVersionStart))

	commitWithVersionStart := time.Now()
	_, err = accountAssetTrees.Get(accountInfo.AccountIndex).CommitWithNewVersion(nil, &newVersion)
	logx.WithContext(ctxLog).Debugf("doCommitAssetSmt1 CommitWithNewVersion end. cost time %v", time.Since(commitWithVersionStart))

	if err != nil {
		assetStateRootQueue <- &treeUpdateResp{
			err: fmt.Errorf("unable to CommitWithNewVersion asset to tree: %s,newVersion:%d,tree.LatestVersion:%d", err.Error(), uint64(newVersion), uint64(accountAssetTrees.Get(accountInfo.AccountIndex).LatestVersion())),
		}
		return
	}
	logx.WithContext(ctxLog).Debugf("doCommitAssetSmt CommitWithNewVersion end. cost time %v", time.Since(commitWithVersionStart))

	accountHashVal, err := AccountToNode(
		accountInfo.L1Address,
		accountInfo.PublicKey,
		accountInfo.Nonce,
		accountInfo.CollectionNonce,
		accountAssetTrees.Get(accountInfo.AccountIndex).Root(),
		ctx,
	)
	if err != nil {
		assetStateRootQueue <- &treeUpdateResp{
			err: fmt.Errorf("unable to convert account to node: %s", err.Error()),
		}
		return
	}
	assetStateRootQueue <- &treeUpdateResp{
		pendingAccountItem: []bsmt.Item{{Key: uint64(accountInfo.AccountIndex), Val: accountHashVal}},
		err:                nil,
	}
}

func buildPendingAccountItem(assetStateRootQueue chan *treeUpdateResp, commitAssetSmtComplete chan bool, ctxLog context.Context) ([]bsmt.Item, error) {
	pendingAccountItem := make([]bsmt.Item, 0)
	run := true
	for run {
		select {
		case assetStateRoot := <-assetStateRootQueue:
			if assetStateRoot.err != nil {
				return nil, assetStateRoot.err
			}
			pendingAccountItem = append(pendingAccountItem, assetStateRoot.pendingAccountItem...)
		default:
			time.Sleep(1 * time.Second)
			if len(commitAssetSmtComplete) == 1 && len(assetStateRootQueue) == 0 {
				logx.WithContext(ctxLog).Infof("no data in assetStateRootQueue")
				run = false
			}
		}
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

				//check assetRoot,the asset tree root must be equal to the asset tree root stored in the database
				if rootNotBeEqual(assetRoot, accountHistory.AssetRoot) {
					return fmt.Errorf("check asset root error,accountIndex=%d,curHeight=%d,assetRoot=%s not equal accountHistory.AssetRoot=%s,asset.LatestVersion=%d,versions=%s", accountIndex, curHeight, assetRoot, accountHistory.AssetRoot, asset.LatestVersion(), common2.FormatVersion(asset.Versions()))
				}

				//check version,the asset tree version cannot be greater than the block height
				if versionBeGreaterThanHeight(asset.LatestVersion(), bsmt.Version(curHeight)) {
					return fmt.Errorf("check asset root error,accountIndex=%d asset.LatestVersion=%d,versions=%s, curHeight=%d", accountIndex, asset.LatestVersion(), common2.FormatVersion(asset.Versions()), curHeight)
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

		if GetTreeLatestVersion(asset.Versions()) > bsmt.Version(treeHeight) {
			logx.WithContext(ctxLog).Infof("asset tree accountIndex:%d latestVersion:%d is higher than block, rollback to height:%d", accountIndex, asset.LatestVersion(), treeHeight)

			err := asset.Rollback(bsmt.Version(treeHeight))
			if err != nil {
				return fmt.Errorf("unable to rollback asset accountIndex:%d, latestVersion: %d,tree err: %s", accountIndex, asset.LatestVersion(), err.Error())
			}
			assetRoot := common.Bytes2Hex(asset.Root())
			logx.WithContext(ctxLog).Infof("end to rollback asset tree, accountIndex:%d, latestVersion:%d,versions=%s,assetRoot:%s,rollback to height:%d", accountIndex, asset.LatestVersion(), common2.FormatVersion(asset.Versions()), assetRoot, treeHeight)

			//check version,the asset tree version cannot be greater than the block height
			if versionBeGreaterThanHeight(asset.LatestVersion(), bsmt.Version(treeHeight)) {
				return fmt.Errorf("call asset.Rollback successfully,but fail to rollback asset accountIndex:%d latestVersion: %d,versions=%s", accountIndex, asset.LatestVersion(), common2.FormatVersion(asset.Versions()))
			}
		}
	}
	return nil
}

func RollBackAccountTree(treeHeight int64, accountTree bsmt.SparseMerkleTree) error {
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, treeHeight)
	logx.WithContext(ctxLog).Infof("check to rollback account tree, latestVersion:%d,versions=%s,accountRoot:%s,rollback to height:%d", accountTree.LatestVersion(), common2.FormatVersion(accountTree.Versions()), common.Bytes2Hex(accountTree.Root()), treeHeight)

	if GetTreeLatestVersion(accountTree.Versions()) > bsmt.Version(treeHeight) {
		logx.WithContext(ctxLog).Infof("account tree latestVersion:%d is higher than block, rollback to %d", accountTree.LatestVersion(), treeHeight)

		err := accountTree.Rollback(bsmt.Version(treeHeight))
		if err != nil {
			return fmt.Errorf("unable to rollback account latestVersion:%d,err:%s", treeHeight, err.Error())
		}
		logx.WithContext(ctxLog).Infof("end to rollback account tree, latestVersion:%d,versions=%s,accountRoot:%s,rollback to height:%d", accountTree.LatestVersion(), common2.FormatVersion(accountTree.Versions()), common.Bytes2Hex(accountTree.Root()), treeHeight)

		//check version,the account tree version cannot be greater than the block height
		if versionBeGreaterThanHeight(accountTree.LatestVersion(), bsmt.Version(treeHeight)) {
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

	//check stateRoot,the state root must be equal to the state root stored in the database
	if rootNotBeEqual(newStateRootStr, blockInfo.StateRoot) {
		return fmt.Errorf("checkStateRoot smt tree newStateRoot=%s not equal database StateRoot=%s,height:%d,AccountTree.Root=%s,NftTree.Root=%s", newStateRootStr, blockInfo.StateRoot, blockInfo.BlockHeight, common.Bytes2Hex(accountTreeRoot), common.Bytes2Hex(nftTreeRoot))
	}
	return nil
}

func rootNotBeEqual(rootStoredInCache string, rootStoredInDb string) bool {
	return rootStoredInCache != rootStoredInDb
}

func versionBeGreaterThanHeight(version bsmt.Version, height bsmt.Version) bool {
	return version > height
}
