/*
 * Copyright © 2021 ZkBAS Protocol
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
	"bytes"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	bsmt "github.com/bnb-chain/bas-smt"
	curve "github.com/bnb-chain/zkbas-crypto/ecc/ztwistededwards/tebn254"
	"github.com/bnb-chain/zkbas-crypto/ffmath"
	common2 "github.com/bnb-chain/zkbas/common"
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

	accPrunedVersion := bsmt.Version(version)
	if accountTree.LatestVersion() < accPrunedVersion {
		accPrunedVersion = accountTree.LatestVersion()
	}
	ver, err := accountTree.Commit(&accPrunedVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to commit account tree, tree ver: %d, prune ver: %d", ver, accPrunedVersion)
	}
	for idx := range *assetTrees {
		assetPrunedVersion := bsmt.Version(version)
		if (*assetTrees)[idx].LatestVersion() < assetPrunedVersion {
			assetPrunedVersion = (*assetTrees)[idx].LatestVersion()
		}
		ver, err := (*assetTrees)[idx].Commit(&assetPrunedVersion)
		if err != nil {
			return errors.Wrapf(err, "unable to commit asset tree [%d], tree ver: %d, prune ver: %d", idx, ver, assetPrunedVersion)
		}
	}
	liquidityPrunedVersion := bsmt.Version(version)
	if liquidityTree.LatestVersion() < liquidityPrunedVersion {
		liquidityPrunedVersion = liquidityTree.LatestVersion()
	}
	ver, err = liquidityTree.Commit(&liquidityPrunedVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to commit liquidity tree, tree ver: %d, prune ver: %d", ver, liquidityPrunedVersion)
	}
	nftPrunedVersion := bsmt.Version(version)
	if nftTree.LatestVersion() < nftPrunedVersion {
		nftPrunedVersion = nftTree.LatestVersion()
	}
	ver, err = nftTree.Commit(&nftPrunedVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to commit nft tree, tree ver: %d, prune ver: %d", ver, nftPrunedVersion)
	}
	return nil
}

func RollBackTrees(version uint64,
	accountTree bsmt.SparseMerkleTree,
	assetTrees *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree) error {

	ver := bsmt.Version(version)
	if accountTree.LatestVersion() > ver && !accountTree.IsEmpty() {
		err := accountTree.Rollback(ver)
		if err != nil {
			return errors.Wrapf(err, "unable to rollback account tree, ver: %d", ver)
		}
	}

	for idx, assetTree := range *assetTrees {
		if assetTree.LatestVersion() > ver && !assetTree.IsEmpty() {
			err := assetTree.Rollback(ver)
			if err != nil {
				return errors.Wrapf(err, "unable to rollback asset tree [%d], ver: %d", idx, ver)
			}
		}
	}

	if liquidityTree.LatestVersion() > ver && !liquidityTree.IsEmpty() {
		err := liquidityTree.Rollback(ver)
		if err != nil {
			return errors.Wrapf(err, "unable to rollback liquidity tree, ver: %d", ver)
		}
	}

	if nftTree.LatestVersion() > ver && !nftTree.IsEmpty() {
		err := nftTree.Rollback(ver)
		if err != nil {
			return errors.Wrapf(err, "unable to rollback nft tree, tree ver: %d", ver)
		}
	}
	return nil
}

func ComputeAccountLeafHash(
	accountNameHash string,
	pk string,
	nonce int64,
	collectionNonce int64,
	assetRoot []byte,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	buf.Write(common.FromHex(accountNameHash))
	err = common2.PaddingPkIntoBuf(&buf, pk)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] unable to write pk into buf: %s", err.Error())
		return nil, err
	}
	common2.PaddingInt64IntoBuf(&buf, nonce)
	common2.PaddingInt64IntoBuf(&buf, collectionNonce)
	buf.Write(assetRoot)
	hFunc.Reset()
	hFunc.Write(buf.Bytes())
	hashVal = hFunc.Sum(nil)
	return hashVal, nil
}

func ComputeAccountAssetLeafHash(
	balance string,
	lpAmount string,
	offerCanceledOrFinalized string,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	err = common2.PaddingStringBigIntIntoBuf(&buf, balance)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] invalid balance: %s", err.Error())
		return nil, err
	}
	err = common2.PaddingStringBigIntIntoBuf(&buf, lpAmount)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] invalid balance: %s", err.Error())
		return nil, err
	}
	err = common2.PaddingStringBigIntIntoBuf(&buf, offerCanceledOrFinalized)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] invalid balance: %s", err.Error())
		return nil, err
	}
	hFunc.Write(buf.Bytes())
	return hFunc.Sum(nil), nil
}

func ComputeLiquidityAssetLeafHash(
	assetAId int64,
	assetA string,
	assetBId int64,
	assetB string,
	lpAmount string,
	kLast string,
	feeRate int64,
	treasuryAccountIndex int64,
	treasuryRate int64,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	common2.PaddingInt64IntoBuf(&buf, assetAId)
	err = common2.PaddingStringBigIntIntoBuf(&buf, assetA)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	common2.PaddingInt64IntoBuf(&buf, assetBId)
	err = common2.PaddingStringBigIntIntoBuf(&buf, assetB)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	err = common2.PaddingStringBigIntIntoBuf(&buf, lpAmount)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	err = common2.PaddingStringBigIntIntoBuf(&buf, kLast)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	common2.PaddingInt64IntoBuf(&buf, feeRate)
	common2.PaddingInt64IntoBuf(&buf, treasuryAccountIndex)
	common2.PaddingInt64IntoBuf(&buf, treasuryRate)
	hFunc.Write(buf.Bytes())
	hashVal = hFunc.Sum(nil)
	return hashVal, nil
}

func ComputeNftAssetLeafHash(
	creatorAccountIndex int64,
	ownerAccountIndex int64,
	nftContentHash string,
	nftL1Address string,
	nftL1TokenId string,
	creatorTreasuryRate int64,
	collectionId int64,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	common2.PaddingInt64IntoBuf(&buf, creatorAccountIndex)
	common2.PaddingInt64IntoBuf(&buf, ownerAccountIndex)
	buf.Write(ffmath.Mod(new(big.Int).SetBytes(common.FromHex(nftContentHash)), curve.Modulus).FillBytes(make([]byte, 32)))
	err = common2.PaddingAddressIntoBuf(&buf, nftL1Address)
	if err != nil {
		logx.Errorf("[ComputeNftAssetLeafHash] unable to write address to buf: %s", err.Error())
		return nil, err
	}
	err = common2.PaddingStringBigIntIntoBuf(&buf, nftL1TokenId)
	if err != nil {
		logx.Errorf("[ComputeNftAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	common2.PaddingInt64IntoBuf(&buf, creatorTreasuryRate)
	common2.PaddingInt64IntoBuf(&buf, collectionId)
	hFunc.Write(buf.Bytes())
	hashVal = hFunc.Sum(nil)
	return hashVal, nil
}

func ComputeStateRootHash(
	accountRoot []byte,
	liquidityRoot []byte,
	nftRoot []byte,
) []byte {
	hFunc := mimc.NewMiMC()
	hFunc.Write(accountRoot)
	hFunc.Write(liquidityRoot)
	hFunc.Write(nftRoot)
	return hFunc.Sum(nil)
}