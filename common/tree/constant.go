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
	"github.com/zecrey-labs/zecrey-crypto/accumulators/merkleTree"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
)

type (
	SysconfigModel        = sysconfig.SysconfigModel
	AccountModel          = account.AccountModel
	AccountHistoryModel   = account.AccountHistoryModel
	L2NftHistoryModel     = nft.L2NftHistoryModel
	LiquidityModel        = liquidity.LiquidityModel
	LiquidityHistoryModel = liquidity.LiquidityHistoryModel
	AccountHistory        = account.AccountHistory
	AccountL2NftHistory   = nft.L2NftHistory

	Tree = merkleTree.Tree
	Node = merkleTree.Node
)

const (
	AccountTreeHeight   = 32
	AssetTreeHeight     = 16
	LiquidityTreeHeight = 16
	NftTreeHeight       = 40
)

var (
	NilHash                                                                           = merkleTree.NilHash
	NilAccountAssetRoot, NilStateRoot, NilAccountRoot, NilLiquidityRoot, NilNftRoot   []byte
	NilAccountAssetNodeHash, NilAccountNodeHash, NilLiquidityNodeHash, NilNftNodeHash []byte
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
