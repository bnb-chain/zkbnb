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
	"github.com/zecrey-labs/zecrey-crypto/accumulators/merkleTree"
	"github.com/zecrey-labs/zecrey-crypto/hash/bn254/zmimc"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
)

// TODO optimize, bad performance
func InitAccountTree(
	accountModel AccountModel,
	accountHistoryModel AccountHistoryModel,
	blockHeight int64,
) (
	accountTree *Tree, accountAssetTrees []*Tree, err error,
) {
	// get all confirmed accounts
	accounts, err := accountModel.GetConfirmedAccounts()
	if err != nil {
		if err != account.ErrNotFound {
			logx.Errorf("[InitAccountTree] unable to get accounts: %s", err.Error())
			return nil, nil, err
		} else {
			accountTree, err = merkleTree.NewEmptyTree(AccountTreeHeight, NilHash, zmimc.Hmimc)
			if err != nil {
				log.Println("[InitAccountTree] unable to create empty tree:", err)
				return nil, nil, err
			}
			return accountTree, nil, nil
		}
	}
	var (
		accountInfoMap = make(map[int64]*account.Account)
	)
	for _, accountInfo := range accounts {
		accountInfoMap[accountInfo.AccountIndex] = accountInfo
	}
	// get all accountHistories
	_, accountHistories, err := accountHistoryModel.GetValidAccounts(blockHeight)
	if err != nil {
		logx.Errorf("[InitAccountTree] unable to get all accountHistories")
		return nil, nil, err
	}
	for _, accountHistory := range accountHistories {
		accountInfoMap[accountHistory.AccountIndex].Nonce = accountHistory.Nonce
		accountInfoMap[accountHistory.AccountIndex].AssetInfo = accountHistory.AssetInfo
		accountInfoMap[accountHistory.AccountIndex].AssetRoot = accountHistory.AssetRoot
	}
	// get related account info
	var (
		assetsMap     = make([]map[int64]*Node, len(accounts))
		accountsNodes = make([]*Node, len(accounts))
	)
	for accountIndex := int64(0); accountIndex < int64(len(accounts)); accountIndex++ {
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
		assetsMap[accountIndex] = make(map[int64]*Node)
		// create account assets node
		for assetId, assetInfo := range accountInfo.AssetInfo {
			assetsMap[accountIndex][assetId], err = AssetToNode(
				assetInfo.Balance,
				assetInfo.LpAmount,
			)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to convert asset to node: %s", err.Error())
				return nil, nil, err
			}
		}
	}
	// init account state trees
	accountAssetTrees = make([]*Tree, len(accounts))
	for index := int64(0); index < int64(len(accounts)); index++ {
		// create account assets tree
		if assetsMap[index] == nil {
			accountAssetTrees[index], err = merkleTree.NewEmptyTree(AssetTreeHeight, NilHash, zmimc.Hmimc)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to create new tree by assets: %s", err.Error())
				return nil, nil, err
			}
		} else {
			accountAssetTrees[index], err = merkleTree.NewTreeByMap(assetsMap[index], AssetTreeHeight, NilHash, zmimc.Hmimc)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to create new tree by assets: %s", err.Error())
				return nil, nil, err
			}
		}
		accountsNodes[index], err = AccountToNode(
			accountInfoMap[index].AccountNameHash,
			accountInfoMap[index].PublicKey,
			accountInfoMap[index].Nonce,
			accountAssetTrees[index].RootNode.Value,
		)
		if err != nil {
			logx.Errorf("[InitAccountTree] unable to convert account to node: %s", err.Error())
			return nil, nil, err
		}
	}
	accountTree, err = merkleTree.NewTree(accountsNodes, AccountTreeHeight, NilHash, zmimc.Hmimc)
	if err != nil {
		logx.Errorf("[InitAccountTree] unable to create new account tree: %s", err.Error())
		return nil, nil, err
	}
	return accountTree, accountAssetTrees, nil
}
