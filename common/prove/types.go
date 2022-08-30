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

package prove

import (
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/types"
)

type (
	Tx       = tx.Tx
	TxDetail = tx.TxDetail

	Account      = account.Account
	AccountAsset = types.AccountAsset

	PoolInfo = types.LiquidityInfo
	NftInfo  = types.NftInfo

	AccountModel        = account.AccountModel
	AccountHistoryModel = account.AccountHistoryModel

	LiquidityModel        = liquidity.LiquidityModel
	LiquidityHistoryModel = liquidity.LiquidityHistoryModel

	NftModel        = nft.L2NftModel
	NftHistoryModel = nft.L2NftHistoryModel

	TxWitness = block.Tx

	CryptoAccount            = std.Account
	CryptoAccountAsset       = std.AccountAsset
	CryptoLiquidity          = std.Liquidity
	CryptoNft                = std.Nft
	CryptoRegisterZnsTx      = std.RegisterZnsTx
	CryptoCreatePairTx       = std.CreatePairTx
	CryptoUpdatePairRateTx   = std.UpdatePairRateTx
	CryptoDepositTx          = std.DepositTx
	CryptoDepositNftTx       = std.DepositNftTx
	CryptoTransferTx         = std.TransferTx
	CryptoSwapTx             = std.SwapTx
	CryptoAddLiquidityTx     = std.AddLiquidityTx
	CryptoRemoveLiquidityTx  = std.RemoveLiquidityTx
	CryptoWithdrawTx         = std.WithdrawTx
	CryptoCreateCollectionTx = std.CreateCollectionTx
	CryptoMintNftTx          = std.MintNftTx
	CryptoTransferNftTx      = std.TransferNftTx
	CryptoOfferTx            = std.OfferTx
	CryptoAtomicMatchTx      = std.AtomicMatchTx
	CryptoCancelOfferTx      = std.CancelOfferTx
	CryptoWithdrawNftTx      = std.WithdrawNftTx
	CryptoFullExitTx         = std.FullExitTx
	CryptoFullExitNftTx      = std.FullExitNftTx
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
	AccountInfo            *Account
	AccountAssets          []*AccountAsset
	AssetsRelatedTxDetails []*TxDetail
}

type LiquidityWitnessInfo struct {
	LiquidityInfo            *PoolInfo
	LiquidityRelatedTxDetail *TxDetail
}

type NftWitnessInfo struct {
	NftInfo            *NftInfo
	NftRelatedTxDetail *TxDetail
}
