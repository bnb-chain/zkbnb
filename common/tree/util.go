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
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/bnb-chain/zkbas-crypto/accumulators/merkleTree"
	"github.com/bnb-chain/zkbas-crypto/hash/bn254/zmimc"
	"math/big"
)

func NewEmptyAccountAssetTree() (tree *Tree, err error) {
	return merkleTree.NewEmptyTree(AssetTreeHeight, NilAccountAssetNodeHash, zmimc.Hmimc)
}

func NewEmptyAccountTree() (tree *Tree, err error) {
	return merkleTree.NewEmptyTree(AccountTreeHeight, NilAccountNodeHash, zmimc.Hmimc)
}

func NewEmptyLiquidityTree() (tree *Tree, err error) {
	return merkleTree.NewEmptyTree(LiquidityTreeHeight, NilLiquidityNodeHash, zmimc.Hmimc)
}

func NewEmptyNftTree() (tree *Tree, err error) {
	return merkleTree.NewEmptyTree(NftTreeHeight, NilNftNodeHash, zmimc.Hmimc)
}

func EmptyAccountNodeHash() []byte {
	hFunc := mimc.NewMiMC()
	zero := big.NewInt(0).FillBytes(make([]byte, 32))
	/*
		AccountNameHash
		PubKey
		Nonce
		CollectionNonce
		AssetRoot
	*/
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	// asset root
	hFunc.Write(NilAccountAssetRoot)
	return hFunc.Sum(nil)
}

func EmptyAccountAssetNodeHash() []byte {
	hFunc := mimc.NewMiMC()
	zero := big.NewInt(0).FillBytes(make([]byte, 32))
	/*
		balance
		lpAmount
		offerCanceledOrFinalized
	*/
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	return hFunc.Sum(nil)
}

func EmptyLiquidityNodeHash() []byte {
	hFunc := mimc.NewMiMC()
	zero := big.NewInt(0).FillBytes(make([]byte, 32))
	/*
		assetAId
		assetA
		assetBId
		assetB
		lpAmount
		kLast
		feeRate
		treasuryAccountIndex
		treasuryRate
	*/
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	return hFunc.Sum(nil)
}

func EmptyNftNodeHash() []byte {
	hFunc := mimc.NewMiMC()
	zero := big.NewInt(0).FillBytes(make([]byte, 32))
	/*
		creatorAccountIndex
		ownerAccountIndex
		nftContentHash
		nftL1Address
		nftL1TokenId
		creatorTreasuryRate
		collectionId
	*/
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	return hFunc.Sum(nil)
}
