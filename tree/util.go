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
	"github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon"
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
		AccountNameHash
		PubKey
		Nonce
		CollectionNonce
		AssetRoot
	*/
	zero := &fr.Element{0, 0, 0, 0}
	NilAccountAssetRootElement := txtypes.FromBigIntToFr(new(big.Int).SetBytes(NilAccountAssetRoot))
	hash := poseidon.Poseidon(zero, zero, zero, zero, zero, NilAccountAssetRootElement).Bytes()
	return hash[:]
}

func EmptyAccountAssetNodeHash() []byte {
	/*
		balance
		offerCanceledOrFinalized
	*/
	zero := &fr.Element{0, 0, 0, 0}
	hash := poseidon.Poseidon(zero, zero).Bytes()
	return hash[:]
}

func EmptyNftNodeHash() []byte {
	/*
		creatorAccountIndex
		ownerAccountIndex
		nftContentHash
		creatorTreasuryRate
		collectionId
	*/
	zero := &fr.Element{0, 0, 0, 0}
	hash := poseidon.Poseidon(zero, zero, zero, zero, zero).Bytes()
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

func RollBackTrees(
	version uint64,
	accountTree bsmt.SparseMerkleTree,
	assetTrees *AssetTreeCache,
	nftTree bsmt.SparseMerkleTree) error {

	assetTreeChanges := assetTrees.GetChanges()
	totalTask := len(assetTreeChanges) + 2
	errChan := make(chan error, totalTask)
	defer close(errChan)

	ver := bsmt.Version(version)
	err := gopool.Submit(func() {
		if accountTree.LatestVersion() > ver && !accountTree.IsEmpty() {
			err := accountTree.Rollback(ver)
			if err != nil {
				errChan <- errors.Wrapf(err, "unable to rollback account tree, ver: %d", ver)
				return
			}
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
				version := asset.RecentVersion()
				err := asset.Rollback(version)
				if err != nil {
					errChan <- errors.Wrapf(err, "unable to rollback asset tree [%d], ver: %d", i, version)
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
		if nftTree.LatestVersion() > ver && !nftTree.IsEmpty() {
			err := nftTree.Rollback(ver)
			if err != nil {
				errChan <- errors.Wrapf(err, "unable to rollback nft tree, tree ver: %d", ver)
				return
			}
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
	accountIndex int64,
	blockHeight int64,
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
	hash := poseidon.Poseidon(e0, e1, e2, e3, e4, e5).Bytes()
	logx.Infof("compute account leaf hash,blockHeight=%s,accountIndex=%s,nonce=%s,collectionNonce=%s,assetRoot=%s,hash=%s", blockHeight, accountIndex, nonce, collectionNonce, common.Bytes2Hex(assetRoot), common.Bytes2Hex(hash[:]))
	return hash[:], nil
}

func ComputeAccountAssetLeafHash(
	balance string,
	offerCanceledOrFinalized string,
	accountIndex int64,
	assetId int64,
	blockHeight int64,
) (hashVal []byte, err error) {
	balanceBigInt, isValid := new(big.Int).SetString(balance, 10)
	if !isValid {
		return nil, errors.New("invalid balance string")
	}
	e0 := txtypes.FromBigIntToFr(balanceBigInt)

	offerCanceledOrFinalizedBigInt, isValid := new(big.Int).SetString(offerCanceledOrFinalized, 10)
	if !isValid {
		return nil, errors.New("invalid balance string")
	}
	e1 := txtypes.FromBigIntToFr(offerCanceledOrFinalizedBigInt)
	hash := poseidon.Poseidon(e0, e1).Bytes()
	logx.Infof("compute account asset leaf hash,blockHeight=%s,accountIndex=%s,assetId=%s,balance=%s,offerCanceledOrFinalized=%s,hash=%s", blockHeight, accountIndex, assetId, balance, offerCanceledOrFinalized, common.Bytes2Hex(hash[:]))
	return hash[:], nil
}

func ComputeNftAssetLeafHash(
	creatorAccountIndex int64,
	ownerAccountIndex int64,
	nftContentHash string,
	creatorTreasuryRate int64,
	collectionId int64,
	nftContentType int8,
	nftIndex int64,
	blockHeight int64,
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

	e4 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(creatorTreasuryRate))
	e5 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(collectionId))
	e6 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(int64(nftContentType)))
	var hash [32]byte
	if e3 != nil {
		hash = poseidon.Poseidon(e0, e1, e2, e3, e4, e5, e6).Bytes()
	} else {
		hash = poseidon.Poseidon(e0, e1, e2, e4, e5, e6).Bytes()
	}
	logx.Infof("compute nft asset leaf hash,blockHeight=%s,nftIndex=%s,creatorAccountIndex=%s,ownerAccountIndex=%s,nftContentHash=%s,creatorTreasuryRate=%s,collectionId=%s,hash=%s", blockHeight, nftIndex, creatorAccountIndex, ownerAccountIndex, nftContentHash, creatorTreasuryRate, collectionId, common.Bytes2Hex(hash[:]))

	return hash[:], nil
}

func ComputeStateRootHash(
	accountRoot []byte,
	nftRoot []byte,
) []byte {
	e0 := txtypes.FromBigIntToFr(new(big.Int).SetBytes(accountRoot))
	e1 := txtypes.FromBigIntToFr(new(big.Int).SetBytes(nftRoot))
	hash := poseidon.Poseidon(e0, e1).Bytes()
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
