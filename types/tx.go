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

package types

import (
	"encoding/json"

	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
)

const (
	TxTypeEmpty = iota
	TxTypeRegisterZns
	TxTypeCreatePair
	TxTypeUpdatePairRate
	TxTypeDeposit
	TxTypeDepositNft
	TxTypeTransfer
	TxTypeSwap
	TxTypeAddLiquidity
	TxTypeRemoveLiquidity
	TxTypeWithdraw
	TxTypeCreateCollection
	TxTypeMintNft
	TxTypeTransferNft
	TxTypeAtomicMatch
	TxTypeCancelOffer
	TxTypeWithdrawNft
	TxTypeFullExit
	TxTypeFullExitNft
	TxTypeOffer
)

func IsL2Tx(txType int64) bool {
	if txType == TxTypeTransfer ||
		txType == TxTypeSwap ||
		txType == TxTypeAddLiquidity ||
		txType == TxTypeRemoveLiquidity ||
		txType == TxTypeWithdraw ||
		txType == TxTypeCreateCollection ||
		txType == TxTypeMintNft ||
		txType == TxTypeTransferNft ||
		txType == TxTypeAtomicMatch ||
		txType == TxTypeCancelOffer ||
		txType == TxTypeWithdrawNft {
		return true
	}
	return false
}

type (
	RegisterZnsTxInfo    = legendTxTypes.RegisterZnsTxInfo
	CreatePairTxInfo     = legendTxTypes.CreatePairTxInfo
	UpdatePairRateTxInfo = legendTxTypes.UpdatePairRateTxInfo
	DepositTxInfo        = legendTxTypes.DepositTxInfo
	DepositNftTxInfo     = legendTxTypes.DepositNftTxInfo
	FullExitTxInfo       = legendTxTypes.FullExitTxInfo
	FullExitNftTxInfo    = legendTxTypes.FullExitNftTxInfo

	TransferTxInfo         = legendTxTypes.TransferTxInfo
	SwapTxInfo             = legendTxTypes.SwapTxInfo
	AddLiquidityTxInfo     = legendTxTypes.AddLiquidityTxInfo
	RemoveLiquidityTxInfo  = legendTxTypes.RemoveLiquidityTxInfo
	WithdrawTxInfo         = legendTxTypes.WithdrawTxInfo
	CreateCollectionTxInfo = legendTxTypes.CreateCollectionTxInfo
	MintNftTxInfo          = legendTxTypes.MintNftTxInfo
	TransferNftTxInfo      = legendTxTypes.TransferNftTxInfo
	OfferTxInfo            = legendTxTypes.OfferTxInfo
	AtomicMatchTxInfo      = legendTxTypes.AtomicMatchTxInfo
	CancelOfferTxInfo      = legendTxTypes.CancelOfferTxInfo
	WithdrawNftTxInfo      = legendTxTypes.WithdrawNftTxInfo
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

const (
	AddressSize       = 20
	FeeRateBase       = 10000
	EmptyStringKeccak = "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
)

func ParseRegisterZnsTxInfo(txInfoStr string) (txInfo *legendTxTypes.RegisterZnsTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseCreatePairTxInfo(txInfoStr string) (txInfo *legendTxTypes.CreatePairTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseUpdatePairRateTxInfo(txInfoStr string) (txInfo *legendTxTypes.UpdatePairRateTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseDepositTxInfo(txInfoStr string) (txInfo *legendTxTypes.DepositTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseDepositNftTxInfo(txInfoStr string) (txInfo *legendTxTypes.DepositNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseFullExitTxInfo(txInfoStr string) (txInfo *legendTxTypes.FullExitTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseFullExitNftTxInfo(txInfoStr string) (txInfo *legendTxTypes.FullExitNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseCreateCollectionTxInfo(txInfoStr string) (txInfo *legendTxTypes.CreateCollectionTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseTransferTxInfo(txInfoStr string) (txInfo *legendTxTypes.TransferTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseSwapTxInfo(txInfoStr string) (txInfo *legendTxTypes.SwapTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseAddLiquidityTxInfo(txInfoStr string) (txInfo *legendTxTypes.AddLiquidityTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseRemoveLiquidityTxInfo(txInfoStr string) (txInfo *legendTxTypes.RemoveLiquidityTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseMintNftTxInfo(txInfoStr string) (txInfo *legendTxTypes.MintNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseTransferNftTxInfo(txInfoStr string) (txInfo *legendTxTypes.TransferNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseAtomicMatchTxInfo(txInfoStr string) (txInfo *legendTxTypes.AtomicMatchTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseCancelOfferTxInfo(txInfoStr string) (txInfo *legendTxTypes.CancelOfferTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseWithdrawTxInfo(txInfoStr string) (txInfo *legendTxTypes.WithdrawTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func ParseWithdrawNftTxInfo(txInfoStr string) (txInfo *legendTxTypes.WithdrawNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}
