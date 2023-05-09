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
	"context"
	"github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	zkbnbtypes "github.com/bnb-chain/zkbnb/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"sort"
	"time"

	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/gopool"
)

func EmptyAccountNodeHash() []byte {
	/*
		L1Address
		PubKey
		Nonce
		CollectionNonce
		AssetRoot
	*/
	zero := &fr.Element{0, 0, 0, 0}
	NilAccountAssetRootElement := txtypes.FromBigIntToFr(new(big.Int).SetBytes(NilAccountAssetRoot))
	ele := GMimcElements([]*fr.Element{zero, zero, zero, zero, zero, NilAccountAssetRootElement})
	hash := ele.Bytes()
	return hash[:]
}

func EmptyAccountAssetNodeHash() []byte {
	/*
		balance
		offerCanceledOrFinalized
	*/
	zero := &fr.Element{0, 0, 0, 0}
	ele := GMimcElements([]*fr.Element{zero, zero})
	hash := ele.Bytes()
	return hash[:]
}

func EmptyNftNodeHash() []byte {
	/*
		creatorAccountIndex
		ownerAccountIndex
		nftContentHash
		royaltyRate
		collectionId
	*/
	zero := &fr.Element{0, 0, 0, 0}
	ele := GMimcElements([]*fr.Element{zero, zero, zero, zero, zero})
	hash := ele.Bytes()
	return hash[:]
}

func CommitAccountAndNftTree(
	prunedVersion uint64,
	blockHeight int64,
	accountTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree) error {
	totalTask := 2
	errChan := make(chan error, totalTask)
	defer close(errChan)

	err := gopool.Submit(func() {
		accPrunedVersion := bsmt.Version(GetAssetLatestVerifiedHeight(int64(prunedVersion), accountTree.Versions()))
		newVersion := bsmt.Version(blockHeight)
		if accountTree.LatestVersion() < accPrunedVersion {
			accPrunedVersion = accountTree.LatestVersion()
		}
		ver, err := accountTree.CommitWithNewVersion(&accPrunedVersion, &newVersion)
		if err != nil {
			errChan <- errors.Wrapf(err, "unable to commit account tree, tree ver: %d, prune ver: %d", ver, accPrunedVersion)
			return
		}
		errChan <- nil
	})
	if err != nil {
		return err
	}

	err = gopool.Submit(func() {
		nftPrunedVersion := bsmt.Version(GetAssetLatestVerifiedHeight(int64(prunedVersion), nftTree.Versions()))
		newVersion := bsmt.Version(blockHeight)
		if nftTree.LatestVersion() < nftPrunedVersion {
			nftPrunedVersion = nftTree.LatestVersion()
		}
		ver, err := nftTree.CommitWithNewVersion(&nftPrunedVersion, &newVersion)
		if err != nil {
			errChan <- errors.Wrapf(err, "unable to commit nft tree, tree ver: %d, prune ver: %d", ver, nftPrunedVersion)
			return
		}
		errChan <- nil
	})
	if err != nil {
		return err
	}

	for i := 0; i < totalTask; i++ {
		err := <-errChan
		if err != nil {
			return err
		}
	}

	return nil
}

func CommitTrees(
	prunedVersion uint64,
	blockHeight int64,
	accountTree bsmt.SparseMerkleTree,
	assetTrees *AssetTreeCache,
	nftTree bsmt.SparseMerkleTree) error {
	start := time.Now()
	assetTreeChanges := assetTrees.GetChanges()
	logx.Infof("GetChanges=%v", time.Since(start))
	totalTask := len(assetTreeChanges) + 2
	errChan := make(chan error, totalTask)
	defer close(errChan)

	err := gopool.Submit(func() {
		accPrunedVersion := bsmt.Version(GetAssetLatestVerifiedHeight(int64(prunedVersion), accountTree.Versions()))
		newVersion := bsmt.Version(blockHeight)
		if accountTree.LatestVersion() < accPrunedVersion {
			accPrunedVersion = accountTree.LatestVersion()
		}
		ver, err := accountTree.CommitWithNewVersion(&accPrunedVersion, &newVersion)
		if err != nil {
			errChan <- errors.Wrapf(err, "unable to commit account tree, tree ver: %d, prune ver: %d", ver, accPrunedVersion)
			return
		}
		errChan <- nil
	})
	if err != nil {
		return err
	}

	for _, idx := range assetTreeChanges {
		err := func(i int64) error {
			return gopool.Submit(func() {
				asset := assetTrees.Get(i)
				prunedVersion := bsmt.Version(GetAssetLatestVerifiedHeight(int64(prunedVersion), asset.Versions()))
				latestVersion := asset.LatestVersion()
				if prunedVersion > latestVersion {
					prunedVersion = latestVersion
				}
				newVersion := bsmt.Version(blockHeight)
				ver, err := asset.CommitWithNewVersion(&prunedVersion, &newVersion)
				if err != nil {
					errChan <- errors.Wrapf(err, "unable to commit asset tree [%d], tree ver: %d, prune ver: %d", i, ver, prunedVersion)
					return
				}
				errChan <- nil
			})
		}(idx)
		if err != nil {
			return err
		}
	}

	err = gopool.Submit(func() {
		nftPrunedVersion := bsmt.Version(GetAssetLatestVerifiedHeight(int64(prunedVersion), nftTree.Versions()))
		newVersion := bsmt.Version(blockHeight)
		if nftTree.LatestVersion() < nftPrunedVersion {
			nftPrunedVersion = nftTree.LatestVersion()
		}
		ver, err := nftTree.CommitWithNewVersion(&nftPrunedVersion, &newVersion)
		if err != nil {
			errChan <- errors.Wrapf(err, "unable to commit nft tree, tree ver: %d, prune ver: %d", ver, nftPrunedVersion)
			return
		}
		errChan <- nil
	})
	if err != nil {
		return err
	}

	for i := 0; i < totalTask; i++ {
		err := <-errChan
		if err != nil {
			return err
		}
	}

	return nil
}

func ComputeAccountLeafHash(
	l1Address string,
	pk string,
	nonce int64,
	collectionNonce int64,
	assetRoot []byte,
	ctx context.Context,
) (hashVal []byte, err error) {
	var e0 *fr.Element
	if l1Address == "" {
		e0 = &fr.Element{0, 0, 0, 0}
		e0.SetBytes([]byte{})
	} else {
		e0, err = txtypes.FromBytesToFr(common.FromHex(l1Address))
		if err != nil {
			return nil, err
		}
	}
	pubKey, err := common2.ParsePubKey(pk)
	if err != nil {
		return nil, err
	}
	e1 := &pubKey.A.X
	e2 := &pubKey.A.Y
	e3 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(nonce))
	e4 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(collectionNonce))
	e5 := txtypes.FromBigIntToFr(new(big.Int).SetBytes(assetRoot))
	ele := GMimcElements([]*fr.Element{e0, e1, e2, e3, e4, e5})
	hash := ele.Bytes()
	logx.WithContext(ctx).Infof("compute account leaf hash,l1Address=%s,pk=%s,nonce=%d,collectionNonce=%d,assetRoot=%s,hash=%s",
		l1Address, pk, nonce, collectionNonce, common.Bytes2Hex(assetRoot), common.Bytes2Hex(hash[:]))
	return hash[:], nil
}

func ComputeAccountAssetLeafHash(
	balance string,
	offerCanceledOrFinalized string,
	ctx context.Context,
) (hashVal []byte, err error) {
	balanceBigInt, isValid := new(big.Int).SetString(balance, 10)
	if !isValid {
		return nil, zkbnbtypes.AppErrInvalidBalanceString
	}
	e0 := txtypes.FromBigIntToFr(balanceBigInt)

	offerCanceledOrFinalizedBigInt, isValid := new(big.Int).SetString(offerCanceledOrFinalized, 10)
	if !isValid {
		return nil, zkbnbtypes.AppErrInvalidBalanceString
	}
	e1 := txtypes.FromBigIntToFr(offerCanceledOrFinalizedBigInt)
	ele := GMimcElements([]*fr.Element{e0, e1})
	hash := ele.Bytes()
	logx.WithContext(ctx).Infof("compute account asset leaf hash,balance=%s,offerCanceledOrFinalized=%s,hash=%s",
		balance, offerCanceledOrFinalized, common.Bytes2Hex(hash[:]))
	return hash[:], nil
}

func ComputeNftAssetLeafHash(
	creatorAccountIndex int64,
	ownerAccountIndex int64,
	nftContentHash string,
	royaltyRate int64,
	collectionId int64,
	ctx context.Context,
) (hashVal []byte, err error) {
	e0 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(creatorAccountIndex))
	e1 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(ownerAccountIndex))

	var e2 *fr.Element
	var e3 *fr.Element
	contentHash := common.Hex2Bytes(nftContentHash)
	if len(contentHash) >= types.NftContentHashBytesSize {
		e2, err = txtypes.FromBytesToFr(contentHash[:types.NftContentHashBytesSize])
		e3, err = txtypes.FromBytesToFr(contentHash[types.NftContentHashBytesSize:])
	} else {
		e2, err = txtypes.FromBytesToFr(contentHash[:])
	}
	if err != nil {
		return nil, err
	}

	e4 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(royaltyRate))
	e5 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(collectionId))
	var hash [32]byte
	if e3 != nil {
		ele := GMimcElements([]*fr.Element{e0, e1, e2, e3, e4, e5})
		hash = ele.Bytes()
	} else {
		ele := GMimcElements([]*fr.Element{e0, e1, e2, e4, e5})
		hash = ele.Bytes()
	}
	logx.WithContext(ctx).Infof("compute nft asset leaf hash,creatorAccountIndex=%d,ownerAccountIndex=%d,nftContentHash=%s,royaltyRate=%d,collectionId=%d,hash=%s",
		creatorAccountIndex, ownerAccountIndex, nftContentHash, royaltyRate, collectionId, common.Bytes2Hex(hash[:]))

	return hash[:], nil
}

func ComputeStateRootHash(
	accountRoot []byte,
	nftRoot []byte,
) []byte {
	e0 := txtypes.FromBigIntToFr(new(big.Int).SetBytes(accountRoot))
	e1 := txtypes.FromBigIntToFr(new(big.Int).SetBytes(nftRoot))

	ele := GMimcElements([]*fr.Element{e0, e1})
	hash := ele.Bytes()
	return hash[:]
}

func GetAssetLatestVerifiedHeight(height int64, versions []bsmt.Version) int64 {
	if versions == nil {
		return height
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] < versions[j]
	})

	latestVerifiedHeight := height
	for _, version := range versions {
		if int64(version) > height {
			break
		}
		latestVerifiedHeight = int64(version)
	}
	return latestVerifiedHeight
}

func GetTreeLatestVersion(versions []bsmt.Version) bsmt.Version {
	if versions == nil || len(versions) == 0 {
		return bsmt.Version(0)
	}
	return versions[len(versions)-1]
}
