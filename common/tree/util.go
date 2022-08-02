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
	"math/big"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/pkg/errors"
)

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

func CommitTrees(version uint64,
	accountTree bsmt.SparseMerkleTree,
	assetTrees *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree) error {

	prunedVersion := bsmt.Version(version)
	ver, err := accountTree.Commit(&prunedVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to commit account tree, tree ver: %d, prune ver: %d", ver, prunedVersion)
	}
	for idx, assetTree := range *assetTrees {
		ver, err := assetTree.Commit(&prunedVersion)
		if err != nil {
			return errors.Wrapf(err, "unable to commit asset tree [%d], tree ver: %d, prune ver: %d", idx, ver, prunedVersion)
		}
	}
	ver, err = liquidityTree.Commit(&prunedVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to commit liquidity tree, tree ver: %d, prune ver: %d", ver, prunedVersion)
	}
	ver, err = nftTree.Commit(&prunedVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to commit nft tree, tree ver: %d, prune ver: %d", ver, prunedVersion)
	}
	return nil
}
