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
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"time"

	bsmt "github.com/bnb-chain/zkbnb-smt"

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

func CommitTrees(
	version uint64,
	accountTree bsmt.SparseMerkleTree,
	assetTrees *AssetTreeCache,
	nftTree bsmt.SparseMerkleTree) error {
	start := time.Now()
	assetTreeChanges := assetTrees.GetChanges()
	logx.Infof("GetChanges=%d", float64(time.Since(start).Milliseconds()))

	defer assetTrees.CleanChanges()
	totalTask := len(assetTreeChanges) + 2

	errChan := make(chan error, totalTask)
	defer close(errChan)

	err := gopool.Submit(func() {
		accPrunedVersion := bsmt.Version(version)
		if accountTree.LatestVersion() < accPrunedVersion {
			accPrunedVersion = accountTree.LatestVersion()
		}
		ver, err := accountTree.Commit(&accPrunedVersion)
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
				version := asset.LatestVersion()
				ver, err := asset.Commit(&version)
				if err != nil {
					errChan <- errors.Wrapf(err, "unable to commit asset tree [%d], tree ver: %d, prune ver: %d", i, ver, version)
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
		nftPrunedVersion := bsmt.Version(version)
		if nftTree.LatestVersion() < nftPrunedVersion {
			nftPrunedVersion = nftTree.LatestVersion()
		}
		ver, err := nftTree.Commit(&nftPrunedVersion)
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
	defer assetTrees.CleanChanges()
	totalTask := len(assetTreeChanges) + 3
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
	accountNameHash string,
	pk string,
	nonce int64,
	collectionNonce int64,
	assetRoot []byte,
) (hashVal []byte, err error) {
	e0, err := txtypes.FromHexStrToFr(accountNameHash)
	if err != nil {
		return nil, err
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
	return hash[:], nil
}

func ComputeAccountAssetLeafHash(
	balance string,
	offerCanceledOrFinalized string,
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
	return hash[:], nil
}

func ComputeNftAssetLeafHash(
	creatorAccountIndex int64,
	ownerAccountIndex int64,
	nftContentHash string,
	creatorTreasuryRate int64,
	collectionId int64,
) (hashVal []byte, err error) {
	e0 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(creatorAccountIndex))
	e1 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(ownerAccountIndex))
	e2, err := txtypes.FromHexStrToFr(nftContentHash)
	if err != nil {
		return nil, err
	}
	e3 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(creatorTreasuryRate))
	e4 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(collectionId))
	hash := poseidon.Poseidon(e0, e1, e2, e3, e4).Bytes()
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
