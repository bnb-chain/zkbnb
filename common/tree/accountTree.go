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
	"github.com/zeromicro/go-zero/core/logx"
	"log"
)

// TODO optimize, bad performance
func InitAccountTree(
	accountHistoryModel AccountHistoryModel,
	accountAssetHistoryModel AccountAssetHistoryModel,
	accountLiquidityHistoryModel AccountLiquidityHistoryModel,
	blockHeight int64,
) (
	accountTree *Tree, accountStateTrees []*AccountStateTree, err error,
) {
	// get all accounts
	_, accounts, err := accountHistoryModel.GetValidAccounts(blockHeight)
	if err != nil {
		logx.Errorf("[InitAccountTree] unable to get all accounts")
		return nil, nil, err
	}
	if len(accounts) == 0 {
		accountTree, err = merkleTree.NewEmptyTree(AccountTreeHeight, NilHash, zmimc.Hmimc)
		if err != nil {
			log.Println("[InitAccountTree] unable to create empty tree:", err)
			return nil, nil, err
		}
		return accountTree, nil, nil
	}
	if int64(len(accounts)) != accounts[len(accounts)-1].AccountIndex+1 {
		logx.Errorf("[InitAccountTree] index not match")
		return nil, nil, errors.New("[InitAccountTree] index not match")
	}
	// get all account assets
	_, accountAssets, err := accountAssetHistoryModel.GetLatestAccountAssetsByBlockHeight(blockHeight)
	if err != nil {
		logx.Errorf("[InitAccountTree] unable to get latest account assets")
		return nil, nil, err
	}
	// get all account liquidity assets
	_, accountLiquidityAssets, err := accountLiquidityHistoryModel.GetLatestAccountLiquidityAssetsByBlockHeight(blockHeight)
	if err != nil {
		logx.Errorf("[InitAccountTree] unable to get latest account liquidity assets")
		return nil, nil, err
	}
	// get related account info
	var (
		assetsMap          = make([]map[int64]*Node, len(accounts))
		liquidityAssetsMap = make([]map[int64]*Node, len(accounts))
		accountsNodes      = make([]*Node, len(accounts))
	)
	// create account assets node
	for _, accountAsset := range accountAssets {
		if assetsMap[accountAsset.AccountIndex] == nil {
			assetsMap[accountAsset.AccountIndex] = make(map[int64]*Node)
		}
		assetsMap[accountAsset.AccountIndex][accountAsset.AssetId], err = AssetToNode(accountAsset)
		if err != nil {
			logx.Errorf("[InitAccountTree] unable to convert asset to node: %s", err.Error())
			return nil, nil, err
		}
	}
	// create account liquidity assets node
	for _, accountLiquidityAsset := range accountLiquidityAssets {
		accountIndex := accountLiquidityAsset.AccountIndex
		pairIndex := accountLiquidityAsset.PairIndex
		if liquidityAssetsMap[accountIndex] == nil {
			liquidityAssetsMap[accountIndex] = make(map[int64]*Node)
		}
		liquidityAssetsMap[accountIndex][pairIndex], err = LiquidityAssetToNode(accountLiquidityAsset)
		if err != nil {
			logx.Errorf("[InitAccountTree] unable to convert liquidity asset to node: %s", err.Error())
			return nil, nil, err
		}
	}
	// init account state trees
	accountStateTrees = make([]*AccountStateTree, len(accounts))
	for index := 0; index < len(accounts); index++ {
		accountStateTrees[index] = new(AccountStateTree)
		// create account assets tree
		if assetsMap[index] == nil {
			accountStateTrees[index].AssetTree, err = merkleTree.NewEmptyTree(AssetTreeHeight, NilHash, zmimc.Hmimc)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to create new tree by assets: %s", err.Error())
				return nil, nil, err
			}
		} else {
			accountStateTrees[index].AssetTree, err = merkleTree.NewTreeByMap(assetsMap[index], AssetTreeHeight, NilHash, zmimc.Hmimc)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to create new tree by assets: %s", err.Error())
				return nil, nil, err
			}
		}
		// create account liquidity assets tree
		if liquidityAssetsMap[index] == nil {
			accountStateTrees[index].LiquidityTree, err = merkleTree.NewEmptyTree(LiquidityTreeHeight, NilHash, zmimc.Hmimc)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to create new tree by assets: %s", err.Error())
				return nil, nil, err
			}
		} else {
			accountStateTrees[index].LiquidityTree, err = merkleTree.NewTreeByMap(liquidityAssetsMap[index], LiquidityTreeHeight, NilHash, zmimc.Hmimc)
			if err != nil {
				logx.Errorf("[InitAccountTree] unable to create new tree by assets: %s", err.Error())
				return nil, nil, err
			}
		}
		accountsNodes[index], err = AccountHistoryToNode(
			accounts[index],
			accountStateTrees[index].AssetTree.RootNode.Value,
			accountStateTrees[index].LiquidityTree.RootNode.Value,
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
	return accountTree, accountStateTrees, nil
}
