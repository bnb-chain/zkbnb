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

package util

import (
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
)

type (
	RegisterZnsTxInfo    = commonTx.RegisterZnsTxInfo
	CreatePairTxInfo     = commonTx.CreatePairTxInfo
	UpdatePairRateTxInfo = commonTx.UpdatePairRateTxInfo
	DepositTxInfo        = commonTx.DepositTxInfo
	DepositNftTxInfo     = commonTx.DepositNftTxInfo
	FullExitTxInfo       = commonTx.FullExitTxInfo
	FullExitNftTxInfo    = commonTx.FullExitNftTxInfo
)

const (
	GeneralAssetType   = commonAsset.GeneralAssetType
	LiquidityAssetType = commonAsset.LiquidityAssetType
	NftAssetType       = commonAsset.NftAssetType

	Base        = 10
	AddressSize = 20

	FeeRateBase = 10000

	AccountAssetPrefix  = "AccountAsset::"
	PoolLiquidityPrefix = "PoolLiquidity::"
	LpPrefix            = "LP::"
	AccountNftPrefix    = "Nft::"

	EmptyStringKeccak = "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
)

const (
	CHUNK_BYTES_SIZE           = 32
	TXTYPE_BYTES_SIZE          = 1
	NFTTYPE_BYTES_SIZE         = 1
	ADDRESS_BYTES_SIZE         = 20
	ACCOUNTINDEX_BYTES_SIZE    = 4
	ACCOUNTNAME_BYTES_SIZE     = 32
	ACCOUNTNAMEHASH_BYTES_SIZE = 32
	PUBKEY_BYTES_SIZE          = 32
	ASSETID_BYTES_SIZE         = 2
	PAIRINDEX_BYTES_SIZE       = 2
	PACKEDAMOUNT_BYTES_SIZE    = 5
	STATEAMOUNT_BYTES_SIZE     = 16
	NFTINDEX_BYTES_SIZE        = 5
	NFTTOKENID_BYTES_SIZE      = 32
	NFTCONTENTHASH_BYTES_SIZE  = 32
	FEERATE_BYTES_SIZE         = 2
	COLLECTIONID_BYTES_SIZE    = 2

	RegisterZnsPubDataSize = TXTYPE_BYTES_SIZE + ACCOUNTINDEX_BYTES_SIZE + ACCOUNTNAME_BYTES_SIZE +
		ACCOUNTNAMEHASH_BYTES_SIZE + PUBKEY_BYTES_SIZE + PUBKEY_BYTES_SIZE
	CreatePairPubDataSize = TXTYPE_BYTES_SIZE + PAIRINDEX_BYTES_SIZE +
		ASSETID_BYTES_SIZE + ASSETID_BYTES_SIZE + FEERATE_BYTES_SIZE + ACCOUNTINDEX_BYTES_SIZE + FEERATE_BYTES_SIZE
	UpdatePairRatePubdataSize = TXTYPE_BYTES_SIZE + PAIRINDEX_BYTES_SIZE +
		FEERATE_BYTES_SIZE + ACCOUNTINDEX_BYTES_SIZE + FEERATE_BYTES_SIZE
	DepositPubDataSize = TXTYPE_BYTES_SIZE + ACCOUNTINDEX_BYTES_SIZE +
		ACCOUNTNAMEHASH_BYTES_SIZE + ASSETID_BYTES_SIZE + STATEAMOUNT_BYTES_SIZE
	DepositNftPubDataSize = TXTYPE_BYTES_SIZE + ACCOUNTINDEX_BYTES_SIZE + NFTINDEX_BYTES_SIZE + ADDRESS_BYTES_SIZE +
		ACCOUNTINDEX_BYTES_SIZE + FEERATE_BYTES_SIZE + NFTCONTENTHASH_BYTES_SIZE + NFTTOKENID_BYTES_SIZE +
		ACCOUNTNAMEHASH_BYTES_SIZE + COLLECTIONID_BYTES_SIZE
	FullExitPubDataSize = TXTYPE_BYTES_SIZE + ACCOUNTINDEX_BYTES_SIZE +
		ACCOUNTNAMEHASH_BYTES_SIZE + ASSETID_BYTES_SIZE + STATEAMOUNT_BYTES_SIZE
	FullExitNftPubDataSize = TXTYPE_BYTES_SIZE + ACCOUNTINDEX_BYTES_SIZE + ACCOUNTINDEX_BYTES_SIZE + FEERATE_BYTES_SIZE +
		NFTINDEX_BYTES_SIZE + COLLECTIONID_BYTES_SIZE + ADDRESS_BYTES_SIZE +
		ACCOUNTNAMEHASH_BYTES_SIZE + ACCOUNTNAMEHASH_BYTES_SIZE +
		NFTCONTENTHASH_BYTES_SIZE + NFTTOKENID_BYTES_SIZE
)

const (
	TypeAccountIndex = iota
	TypeAssetId
	TypeAccountName
	TypeAccountNameOmitSpace
	TypeAccountPk
	TypePairIndex
	TypeLimit
	TypeOffset
	TypeHash
	TypeBlockHeight
	TypeTxType
	TypeChainId
	TypeLPAmount
	TypeAssetAmount
	TypeBoolean
	TypeGasFee
)

const (
	// TODO(Gavin): these constraints is not settled yet and should be revised before production
	minAccountIndex = 0
	maxAccountIndex = (1 << 32) - 1

	minBlockHeight = 0
	maxBlockHeight = (1 << 64) - 1 //60

	minHashLength = 20
	maxHashLength = 100

	minPublicKeyLength = 20
	maxPublicKeyLength = 50 //TODO

	minAssetId = 0
	maxAssetId = (1 << 32) - 1

	maxAccountNameLength          = 30
	maxAccountNameLengthOmitSpace = 20

	minPairIndex = 0
	maxPairIndex = (1 << 16) - 1

	minLimit = 0
	maxLimit = 50

	minOffset = 0
	maxOffset = (1 << 64) - 1 //TODO

	minTxtype = 0
	maxTxtype = 15

	minLPAmount = 0
	maxLPAmount = (1 << 64) - 1

	minAssetAmount = 0
	maxAssetAmount = (1 << 64) - 1

	minGasFee = 0
	maxGasFee = (1 << 64) - 1
)
