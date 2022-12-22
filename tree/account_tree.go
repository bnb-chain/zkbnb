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
	"hash"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb-smt/database/memory"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/types"
)

func accountAssetNamespace(index int64) string {
	return AccountAssetPrefix + strconv.Itoa(int(index)) + ":"
}

func InitAccountTree(
	accountModel account.AccountModel,
	accountHistoryModel account.AccountHistoryModel,
	blockHeight int64,
	ctx *Context,
	assetCacheSize int,
) (
	accountTree bsmt.SparseMerkleTree, accountAssetTrees *AssetTreeCache, err error,
) {

	//todo optimize if the history table has a lot of data, it will take a long time to load
	accountNums, err := accountHistoryModel.GetValidAccountCount(blockHeight)
	//accountNums, err := accountModel.GetAccountsTotalCount()
	if err != nil {
		logx.Errorf("unable to get all accountNums")
		return nil, nil, err
	}
	logx.Infof("getValidAccountCount end")

	opts := ctx.Options(blockHeight)
	nilAccountAssetNodeHashes := NilAccountAssetNodeHashes(AssetTreeHeight, NilAccountAssetNodeHash, ctx.Hasher())

	// init account state trees
	accountAssetTrees = NewLazyTreeCache(assetCacheSize, accountNums-1, blockHeight, func(index, block int64) bsmt.SparseMerkleTree {
		tree, err := bsmt.NewSparseMerkleTree(ctx.Hasher(),
			SetNamespace(ctx, accountAssetNamespace(index)), AssetTreeHeight, nilAccountAssetNodeHashes,
			ctx.Options(block)...)
		if err != nil {
			logx.Errorf("unable to create new tree by assets: %s", err.Error())
			panic(err.Error())
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

	if accountNums == 0 {
		return accountTree, accountAssetTrees, nil
	}

	if ctx.IsLoad() {
		for i := 0; i < int(accountNums); i += ctx.BatchReloadSize() {
			err := reloadAccountTreeFromRDB(
				accountModel, accountHistoryModel, blockHeight,
				i, i+ctx.BatchReloadSize(),
				accountTree, accountAssetTrees)
			if err != nil {
				return nil, nil, err
			}
		}

		for i := int64(0); i < accountNums; i++ {
			_, err := accountAssetTrees.Get(i).Commit(nil)
			if err != nil {
				logx.Errorf("unable to set asset to tree: %s", err.Error())
				return nil, nil, err
			}
		}

		_, err = accountTree.Commit(nil)
		if err != nil {
			logx.Errorf("unable to commit account tree: %s", err.Error())
			return nil, nil, err
		}
		return accountTree, accountAssetTrees, nil
	}

	// It's not loading from RDB, need to check tree version
	if accountTree.LatestVersion() > bsmt.Version(blockHeight) && !accountTree.IsEmpty() {
		logx.Infof("account tree version [%d] is higher than block, rollback to %d", accountTree.LatestVersion(), blockHeight)
		err := accountTree.Rollback(bsmt.Version(blockHeight))
		if err != nil {
			logx.Errorf("unable to rollback account tree: %s, version: %d", err.Error(), blockHeight)
			return nil, nil, err
		}
	}

	for i := int64(0); i < accountNums; i++ {
		asset := accountAssetTrees.Get(i)
		if asset.LatestVersion() > bsmt.Version(blockHeight) && !asset.IsEmpty() {
			logx.Infof("asset tree %d version [%d] is higher than block, rollback to %d", i, asset.LatestVersion(), blockHeight)
			err := asset.Rollback(bsmt.Version(blockHeight))
			if err != nil {
				logx.Errorf("unable to rollback asset [%d] tree: %s, version: %d", i, err.Error(), blockHeight)
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
	offset, limit int,
	accountTree bsmt.SparseMerkleTree,
	accountAssetTrees *AssetTreeCache,
) error {
	_, accountHistories, err := accountHistoryModel.GetValidAccounts(blockHeight,
		limit, offset)
	if err != nil {
		logx.Errorf("unable to get all accountHistories")
		return err
	}

	var (
		accountInfoMap = make(map[int64]*account.Account)
	)

	for _, accountHistory := range accountHistories {
		if accountInfoMap[accountHistory.AccountIndex] == nil {
			//todo optimize fetch all the data from account table,no need to fetch the data from the account table every time
			accountInfo, err := accountModel.GetAccountByIndex(accountHistory.AccountIndex)
			if err != nil {
				logx.Errorf("unable to get account by account index: %s", err.Error())
				return err
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
	for i := int64(0); i < int64(len(accountHistories)); i++ {
		accountIndex := accountHistories[i].AccountIndex
		if accountInfoMap[accountIndex] == nil {
			logx.Errorf("invalid account index")
			return errors.New("invalid account index")
		}
		oAccountInfo := accountInfoMap[accountIndex]
		accountInfo, err := chain.ToFormatAccountInfo(oAccountInfo)
		if err != nil {
			logx.Errorf("unable to convert to format account info: %s", err.Error())
			return err
		}
		// create account assets node
		for assetId, assetInfo := range accountInfo.AssetInfo {
			hashVal, err := AssetToNode(
				assetInfo.Balance.String(),
				assetInfo.OfferCanceledOrFinalized.String(),
			)
			if err != nil {
				logx.Errorf("unable to convert asset to node: %s", err.Error())
				return err
			}
			err = accountAssetTrees.Get(accountIndex).Set(uint64(assetId), hashVal)
			if err != nil {
				logx.Errorf("unable to set asset to tree: %s", err.Error())
				return err
			}
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
			return err
		}
		err = accountTree.Set(uint64(accountIndex), accountHashVal)
		if err != nil {
			logx.Errorf("unable to set account to tree: %s", err.Error())
			return err
		}
	}

	return nil
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
