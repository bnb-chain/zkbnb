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

package prove

import (
	"github.com/bnb-chain/zkbnb-crypto/circuit/bn254/block"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type (
	TxWitness = block.Tx
)

const (
	NbAccountAssetsPerAccount = block.NbAccountAssetsPerAccount
	NbAccountsPerTx           = block.NbAccountsPerTx
	AssetMerkleLevels         = block.AssetMerkleLevels
	LiquidityMerkleLevels     = block.LiquidityMerkleLevels
	NftMerkleLevels           = block.NftMerkleLevels
	AccountMerkleLevels       = block.AccountMerkleLevels

	LastAccountIndex   = 4294967295
	LastAccountAssetId = 65535
	LastPairIndex      = 65535
	LastNftIndex       = 1099511627775
)

type AccountWitnessInfo struct {
	AccountInfo            *account.Account
	AccountAssets          []*types.AccountAsset
	AssetsRelatedTxDetails []*tx.TxDetail
}

type LiquidityWitnessInfo struct {
	LiquidityInfo            *types.LiquidityInfo
	LiquidityRelatedTxDetail *tx.TxDetail
}

type NftWitnessInfo struct {
	NftInfo            *types.NftInfo
	NftRelatedTxDetail *tx.TxDetail
}
