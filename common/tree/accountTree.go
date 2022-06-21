/*
 * Copyright © 2021 Zecrey Protocol
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

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-crypto/accumulators/merkleTree"
	"github.com/bnb-chain/zkbas-crypto/hash/bn254/zmimc"
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/account"
)

// TODO optimize, bad performance
func InitAccountTree(
	accountModel AccountModel,
	accountHistoryModel AccountHistoryModel,
	blockHeight int64,
) (
	accountTree *Tree, accountAssetTrees []*Tree, err error,
) {
	var (
		accountInfoMap = make(map[int64]*account.Account)
	)
	// get all accountHistories
	_, accountHistories, err := accountHistoryModel.GetValidAccounts(blockHeight)
	if err != nil {
		logx.Errorf("[InitAccountTree] unable to get all accountHistories")
		return nil, nil, err
	}
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
				Status:          account.StatusConfirmed,
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
	if len(accountHistories) == 0 {
		accountTree, err = NewEmptyAccountTree()
		if err != nil {
			logx.Errorf("[InitAccountTree] unable to create empty account tree: %s", err.Error())
			return nil, nil, err
		}
		return accountTree, accountAssetTrees, nil
	}
	// get related account info
	var (
		assetsMap     = make([]map[int64]*Node, len(accountHistories))
		accountsNodes = make([]*Node, len(accountHistories))
	)
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
		if oAccountInfo.AssetInfo != commonConstant.NilAssetInfo {
			assetsMap[accountIndex] = make(map[int64]*Node)
		}
		// create account assets node
		for assetId, assetInfo := range accountInfo.AssetInfo {
			assetsMap[accountIndex][assetId], err = AssetToNode(
				assetInfo.Balance.String(),
				assetInfo.LpAmount.String(),
				assetInfo.OfferCanceledOrFinalized.String(),
			)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to convert asset to node: %s", err.Error())
				return nil, nil, err
			}
		}
	}
	// init account state trees
	accountAssetTrees = make([]*Tree, len(accountHistories))
	for index := int64(0); index < int64(len(accountHistories)); index++ {
		// create account assets tree
		if assetsMap[index] == nil {
			accountAssetTrees[index], err = NewEmptyAccountAssetTree()
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to create new tree by assets: %s", err.Error())
				return nil, nil, err
			}
		} else {
			accountAssetTrees[index], err = merkleTree.NewTreeByMap(assetsMap[index], AssetTreeHeight, NilAccountAssetNodeHash, zmimc.Hmimc)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to create new tree by assets: %s", err.Error())
				return nil, nil, err
			}
		}
		accountsNodes[index], err = AccountToNode(
			accountInfoMap[index].AccountNameHash,
			accountInfoMap[index].PublicKey,
			accountInfoMap[index].Nonce,
			accountInfoMap[index].CollectionNonce,
			accountAssetTrees[index].RootNode.Value,
		)
		if err != nil {
			logx.Errorf("[InitAccountTree] unable to convert account to node: %s", err.Error())
			return nil, nil, err
		}
	}
	accountTree, err = merkleTree.NewTree(accountsNodes, AccountTreeHeight, NilAccountNodeHash, zmimc.Hmimc)
	if err != nil {
		logx.Errorf("[InitAccountTree] unable to create new account tree: %s", err.Error())
		return nil, nil, err
	}
	return accountTree, accountAssetTrees, nil
}
