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
	"strconv"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
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
	accountModel AccountModel,
	accountHistoryModel AccountHistoryModel,
	blockHeight int64,
	ctx *Context,
) (
	accountTree bsmt.SparseMerkleTree, accountAssetTrees []bsmt.SparseMerkleTree, err error,
) {
	accountNums, err := accountHistoryModel.GetValidAccountCount(blockHeight)
	if err != nil {
		logx.Errorf("unable to get all accountNums")
		return nil, nil, err
	}

	opts := ctx.Options(blockHeight)

	// init account state trees
	accountAssetTrees = make([]bsmt.SparseMerkleTree, accountNums)
	for index := int64(0); index < accountNums; index++ {
		// create account assets tree
		accountAssetTrees[index], err = bsmt.NewBASSparseMerkleTree(bsmt.NewHasher(mimc.NewMiMC()),
			SetNamespace(ctx, accountAssetNamespace(index)), AssetTreeHeight, NilAccountAssetNodeHash,
			opts...)
		if err != nil {
			logx.Errorf("unable to create new tree by assets: %s", err.Error())
			return nil, nil, err
		}
	}
	accountTree, err = bsmt.NewBASSparseMerkleTree(bsmt.NewHasher(mimc.NewMiMC()),
		SetNamespace(ctx, AccountPrefix), AccountTreeHeight, NilAccountNodeHash,
		opts...)
	if err != nil {
		logx.Errorf("unable to create new account tree: %s", err.Error())
		return nil, nil, err
	}

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

		for i := range accountAssetTrees {
			_, err := accountAssetTrees[i].Commit(nil)
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

	for i := range accountAssetTrees {
		if accountAssetTrees[i].LatestVersion() > bsmt.Version(blockHeight) && !accountAssetTrees[i].IsEmpty() {
			logx.Infof("asset tree %d version [%d] is higher than block, rollback to %d", i, accountAssetTrees[i].LatestVersion(), blockHeight)
			err := accountAssetTrees[i].Rollback(bsmt.Version(blockHeight))
			if err != nil {
				logx.Errorf("unable to rollback asset [%d] tree: %s, version: %d", i, err.Error(), blockHeight)
				return nil, nil, err
			}
		}
	}

	return accountTree, accountAssetTrees, nil
}

func reloadAccountTreeFromRDB(
	accountModel AccountModel,
	accountHistoryModel AccountHistoryModel,
	blockHeight int64,
	offset, limit int,
	accountTree bsmt.SparseMerkleTree,
	accountAssetTrees []bsmt.SparseMerkleTree,
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
				assetInfo.LpAmount.String(),
				assetInfo.OfferCanceledOrFinalized.String(),
			)
			if err != nil {
				logx.Errorf("unable to convert asset to node: %s", err.Error())
				return err
			}
			err = accountAssetTrees[accountIndex].Set(uint64(assetId), hashVal)
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
			accountAssetTrees[accountIndex].Root(),
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

func AssetToNode(balance string, lpAmount string, offerCanceledOrFinalized string) (hashVal []byte, err error) {
	hashVal, err = ComputeAccountAssetLeafHash(balance, lpAmount, offerCanceledOrFinalized)
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

func NewEmptyAccountAssetTree(
	ctx *Context,
	index int64,
	blockHeight uint64,
) (tree bsmt.SparseMerkleTree, err error) {
	return bsmt.NewBASSparseMerkleTree(
		bsmt.NewHasher(mimc.NewMiMC()),
		SetNamespace(ctx, accountAssetNamespace(index)),
		AssetTreeHeight, NilAccountAssetNodeHash,
		ctx.Options(int64(blockHeight))...)
}

func NewMemAccountAssetTree() (tree bsmt.SparseMerkleTree, err error) {
	return bsmt.NewBASSparseMerkleTree(bsmt.NewHasher(mimc.NewMiMC()),
		memory.NewMemoryDB(), AssetTreeHeight, NilAccountAssetNodeHash)
}
