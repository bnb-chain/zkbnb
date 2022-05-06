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
	"github.com/zecrey-labs/zecrey-crypto/accumulators/merkleTree"
	"github.com/zecrey-labs/zecrey-crypto/hash/bn254/zmimc"
	"log"
)

type AccountStateTree struct {
	AssetTree     *Tree
	LiquidityTree *Tree
}

func NewEmptyAccountStateTree() (accountStateTree *AccountStateTree, err error) {
	accountStateTree = new(AccountStateTree)
	accountStateTree.AssetTree, err = merkleTree.NewEmptyTree(AssetTreeHeight, NilHash, zmimc.Hmimc)
	if err != nil {
		log.Println("[NewEmptyAccountStateTree] unable to create empty tree")
		return nil, err
	}
	accountStateTree.LiquidityTree, err = merkleTree.NewEmptyTree(LiquidityTreeHeight, NilHash, zmimc.Hmimc)
	if err != nil {
		log.Println("[NewEmptyAccountStateTree] unable to create empty tree")
		return nil, err
	}
	return accountStateTree, nil
}
