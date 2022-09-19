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
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"

	"github.com/bnb-chain/zkbnb-crypto/accumulators/merkleTree"
)

const (
	AccountTreeHeight   = 32
	AssetTreeHeight     = 16
	LiquidityTreeHeight = 16
	NftTreeHeight       = 40
)

const (
	NFTPrefix          = "nft:"
	LiquidityPrefix    = "liquidity:"
	AccountPrefix      = "account:"
	AccountAssetPrefix = "account_asset:"
)

var (
	NilHash                 = merkleTree.NilHash
	NilStateRoot            []byte
	NilAccountRoot          []byte
	NilLiquidityRoot        []byte
	NilNftRoot              []byte
	NilAccountAssetRoot     []byte
	NilAccountNodeHash      []byte
	NilLiquidityNodeHash    []byte
	NilNftNodeHash          []byte
	NilAccountAssetNodeHash []byte
)

func init() {
	NilAccountAssetNodeHash = EmptyAccountAssetNodeHash()
	NilAccountAssetRoot = NilAccountAssetNodeHash
	hFunc := mimc.NewMiMC()
	for i := 0; i < AssetTreeHeight; i++ {
		hFunc.Reset()
		hFunc.Write(NilAccountAssetRoot)
		hFunc.Write(NilAccountAssetRoot)
		NilAccountAssetRoot = hFunc.Sum(nil)
	}
	NilAccountNodeHash = EmptyAccountNodeHash()
	NilAccountRoot = NilAccountNodeHash
	NilLiquidityNodeHash = EmptyLiquidityNodeHash()
	NilNftNodeHash = EmptyNftNodeHash()
	for i := 0; i < AccountTreeHeight; i++ {
		hFunc.Reset()
		hFunc.Write(NilAccountRoot)
		hFunc.Write(NilAccountRoot)
		NilAccountRoot = hFunc.Sum(nil)
	}
	NilLiquidityRoot = NilLiquidityNodeHash
	for i := 0; i < LiquidityTreeHeight; i++ {
		hFunc.Reset()
		hFunc.Write(NilLiquidityRoot)
		hFunc.Write(NilLiquidityRoot)
		NilLiquidityRoot = hFunc.Sum(nil)
	}
	NilNftRoot = NilNftNodeHash
	for i := 0; i < NftTreeHeight; i++ {
		hFunc.Reset()
		hFunc.Write(NilNftRoot)
		hFunc.Write(NilNftRoot)
		NilNftRoot = hFunc.Sum(nil)
	}
	// nil state root
	hFunc.Reset()
	hFunc.Write(NilAccountRoot)
	hFunc.Write(NilLiquidityRoot)
	hFunc.Write(NilNftRoot)
	NilStateRoot = hFunc.Sum(nil)
}
