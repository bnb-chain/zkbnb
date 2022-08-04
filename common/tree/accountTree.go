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
	"errors"
	"strconv"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/bas-smt/database/memory"
	"github.com/bnb-chain/zkbas-crypto/hash/bn254/zmimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/pkg/treedb"
)

func accountAssetNamespace(index int64) string {
	return AccountAssetPrefix + strconv.Itoa(int(index)) + ":"
}

// TODO optimize, bad performance
func InitAccountTree(
	accountModel AccountModel,
	accountHistoryModel AccountHistoryModel,
	blockHeight int64,
	ctx *treedb.Context,
) (
	accountTree bsmt.SparseMerkleTree, accountAssetTrees []bsmt.SparseMerkleTree, err error,
) {
	// TODO: If there are too many accounts, it may cause reading too long, which can be optimized again
	accountNums, err := accountHistoryModel.GetValidAccountNums(blockHeight)
	if err != nil {
		logx.Errorf("[InitAccountTree] unable to get all accountNums")
		return nil, nil, err
	}

	opts := ctx.Options(blockHeight)

	// init account state trees
	accountAssetTrees = make([]bsmt.SparseMerkleTree, accountNums)
	for index := int64(0); index < int64(accountNums); index++ {
		// create account assets tree
		accountAssetTrees[index], err = bsmt.NewBASSparseMerkleTree(bsmt.NewHasher(zmimc.Hmimc),
			treedb.SetNamespace(ctx, accountAssetNamespace(index)), AssetTreeHeight, NilAccountAssetNodeHash,
			opts...)
		if err != nil {
			logx.Errorf("[InitAccountTree] unable to create new tree by assets: %s", err.Error())
			return nil, nil, err
		}
	}
	accountTree, err = bsmt.NewBASSparseMerkleTree(bsmt.NewHasher(zmimc.Hmimc),
		treedb.SetNamespace(ctx, AccountPrefix), AccountTreeHeight, NilAccountNodeHash,
		opts...)
	if err != nil {
		logx.Errorf("[InitAccountTree] unable to create new account tree: %s", err.Error())
		return nil, nil, err
	}

	if accountNums == 0 {
		return accountTree, accountAssetTrees, nil
	}

	if ctx.IsLoad() {
		_, accountHistories, err := accountHistoryModel.GetValidAccounts(blockHeight)
		if err != nil {
			logx.Errorf("[InitAccountTree] unable to get all accountHistories")
			return nil, nil, err
		}

		var (
			accountInfoMap = make(map[int64]*account.Account)
		)

		for _, accountHistory := range accountHistories {
			if accountInfoMap[accountHistory.AccountIndex] == nil {
				accountInfo, err := accountModel.GetAccountByAccountIndex(accountHistory.AccountIndex)
				if err != nil {
					logx.Errorf("[InitAccountTree] unable to get account by account index: %s", err.Error())
					return nil, nil, err
				}
				accountInfoMap[accountHistory.AccountIndex] = &account.Account{
					AccountIndex:    accountInfo.AccountIndex,
					AccountName:     accountInfo.AccountName,
					PublicKey:       accountInfo.PublicKey,
					AccountNameHash: accountInfo.AccountNameHash,
					L1Address:       accountInfo.L1Address,
					Nonce:           0,
					CollectionNonce: 0,
					Status:          account.AccountStatusConfirmed,
				}
			}
			if accountHistory.Nonce != commonConstant.NilNonce {
				accountInfoMap[accountHistory.AccountIndex].Nonce = accountHistory.Nonce
			}
			if accountHistory.CollectionNonce != commonConstant.NilNonce {
				accountInfoMap[accountHistory.AccountIndex].CollectionNonce = accountHistory.CollectionNonce
			}
			accountInfoMap[accountHistory.AccountIndex].AssetInfo = accountHistory.AssetInfo
			accountInfoMap[accountHistory.AccountIndex].AssetRoot = accountHistory.AssetRoot
		}

		// get related account info
		for accountIndex := int64(0); accountIndex < int64(len(accountHistories)); accountIndex++ {
			if accountInfoMap[accountIndex] == nil {
				logx.Errorf("[InitAccountTree] invalid account index")
				return nil, nil, errors.New("[InitAccountTree] invalid account index")
			}
			oAccountInfo := accountInfoMap[accountIndex]
			accountInfo, err := commonAsset.ToFormatAccountInfo(oAccountInfo)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to convert to format account info: %s", err.Error())
				return nil, nil, err
			}
			// create account assets node
			for assetId, assetInfo := range accountInfo.AssetInfo {
				hashVal, err := AssetToNode(
					assetInfo.Balance.String(),
					assetInfo.LpAmount.String(),
					assetInfo.OfferCanceledOrFinalized.String(),
				)
				if err != nil {
					logx.Errorf("[InitAccountTree] unable to convert asset to node: %s", err.Error())
					return nil, nil, err
				}
				err = accountAssetTrees[accountIndex].Set(uint64(assetId), hashVal)
				if err != nil {
					logx.Errorf("[InitAccountTree] unable to set asset to tree: %s", err.Error())
					return nil, nil, err
				}
				_, err = accountAssetTrees[accountIndex].Commit(nil)
				if err != nil {
					logx.Errorf("[InitAccountTree] unable to commit asset tree: %s", err.Error())
					return nil, nil, err
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
				logx.Errorf("[InitAccountTree] unable to convert account to node: %s", err.Error())
				return nil, nil, err
			}
			err = accountTree.Set(uint64(accountIndex), accountHashVal)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to set account to tree: %s", err.Error())
				return nil, nil, err
			}
			_, err = accountTree.Commit(nil)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to commit account tree: %s", err.Error())
				return nil, nil, err
			}
		}
	} else {
		if accountTree.LatestVersion() > bsmt.Version(blockHeight) && !accountTree.IsEmpty() {
			err := accountTree.Rollback(bsmt.Version(blockHeight))
			logx.Infof("[InitAccountTree] account tree version [%d] if higher than block, rollback to %d", accountTree.LatestVersion(), blockHeight)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to rollback account tree: %s, version: %d", err.Error(), blockHeight)
				return nil, nil, err
			}
		}

		for i, assetTree := range accountAssetTrees {
			if assetTree.LatestVersion() > bsmt.Version(blockHeight) && !assetTree.IsEmpty() {
				err := assetTree.Rollback(bsmt.Version(blockHeight))
				logx.Infof("[InitAccountTree] asset tree %d version [%d] if higher than block, rollback to %d", i, assetTree.LatestVersion(), blockHeight)
				if err != nil {
					logx.Errorf("[InitAccountTree] unable to rollback asset [%d] tree: %s, version: %d", i, err.Error(), blockHeight)
					return nil, nil, err
				}
			}
		}
	}

	return accountTree, accountAssetTrees, nil
}

func AssetToNode(balance string, lpAmount string, offerCanceledOrFinalized string) (hashVal []byte, err error) {
	hashVal, err = ComputeAccountAssetLeafHash(balance, lpAmount, offerCanceledOrFinalized)
	if err != nil {
		logx.Errorf("[AccountToNode] unable to compute asset leaf hash: %s", err.Error())
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
		logx.Errorf("[AccountToNode] unable to compute account leaf hash: %s", err.Error())
		return nil, err
	}

	return hashVal, nil
}

func NewEmptyAccountAssetTree(
	ctx *treedb.Context,
	index int64,
	blockHeight uint64,
) (tree bsmt.SparseMerkleTree, err error) {
	return bsmt.NewBASSparseMerkleTree(
		bsmt.NewHasher(zmimc.Hmimc),
		treedb.SetNamespace(ctx, accountAssetNamespace(index)),
		AssetTreeHeight, NilAccountAssetNodeHash,
		ctx.Options(int64(blockHeight))...)
}

func NewMemAccountAssetTree() (tree bsmt.SparseMerkleTree, err error) {
	return bsmt.NewBASSparseMerkleTree(bsmt.NewHasher(zmimc.Hmimc),
		memory.NewMemoryDB(), AssetTreeHeight, NilAccountAssetNodeHash)
}
