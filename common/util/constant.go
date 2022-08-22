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

package util

import (
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
)

type (
	RegisterZnsTxInfo    = legendTxTypes.RegisterZnsTxInfo
	CreatePairTxInfo     = legendTxTypes.CreatePairTxInfo
	UpdatePairRateTxInfo = legendTxTypes.UpdatePairRateTxInfo
	DepositTxInfo        = legendTxTypes.DepositTxInfo
	DepositNftTxInfo     = legendTxTypes.DepositNftTxInfo
	FullExitTxInfo       = legendTxTypes.FullExitTxInfo
	FullExitNftTxInfo    = legendTxTypes.FullExitNftTxInfo
)

const (
	AddressSize = 20

	FeeRateBase = 10000

	EmptyStringKeccak = "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
)

const (
	TxTypeBytesSize          = 1
	AddressBytesSize         = 20
	AccountIndexBytesSize    = 4
	AccountNameBytesSize     = 32
	AccountNameHashBytesSize = 32
	PubkeyBytesSize          = 32
	AssetIdBytesSize         = 2
	PairIndexBytesSize       = 2
	StateAmountBytesSize     = 16
	NftIndexBytesSize        = 5
	NftTokenIdBytesSize      = 32
	NftContentHashBytesSize  = 32
	FeeRateBytesSize         = 2
	CollectionIdBytesSize    = 2

	RegisterZnsPubDataSize = TxTypeBytesSize + AccountIndexBytesSize + AccountNameBytesSize +
		AccountNameHashBytesSize + PubkeyBytesSize + PubkeyBytesSize
	CreatePairPubDataSize = TxTypeBytesSize + PairIndexBytesSize +
		AssetIdBytesSize + AssetIdBytesSize + FeeRateBytesSize + AccountIndexBytesSize + FeeRateBytesSize
	UpdatePairRatePubdataSize = TxTypeBytesSize + PairIndexBytesSize +
		FeeRateBytesSize + AccountIndexBytesSize + FeeRateBytesSize
	DepositPubDataSize = TxTypeBytesSize + AccountIndexBytesSize +
		AccountNameHashBytesSize + AssetIdBytesSize + StateAmountBytesSize
	DepositNftPubDataSize = TxTypeBytesSize + AccountIndexBytesSize + NftIndexBytesSize + AddressBytesSize +
		AccountIndexBytesSize + FeeRateBytesSize + NftContentHashBytesSize + NftTokenIdBytesSize +
		AccountNameHashBytesSize + CollectionIdBytesSize
	FullExitPubDataSize = TxTypeBytesSize + AccountIndexBytesSize +
		AccountNameHashBytesSize + AssetIdBytesSize + StateAmountBytesSize
	FullExitNftPubDataSize = TxTypeBytesSize + AccountIndexBytesSize + AccountIndexBytesSize + FeeRateBytesSize +
		NftIndexBytesSize + CollectionIdBytesSize + AddressBytesSize +
		AccountNameHashBytesSize + AccountNameHashBytesSize +
		NftContentHashBytesSize + NftTokenIdBytesSize
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
