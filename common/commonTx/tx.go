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

package commonTx

import (
	"encoding/json"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
)

type (
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

type RegisterZnsTxInfo struct {
	TxType          uint8
	AccountIndex    int64
	AccountName     string
	AccountNameHash []byte
	PubKey          string
}

func ParseRegisterZnsTxInfo(txInfoStr string) (txInfo *RegisterZnsTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseRegisterZnsTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

type CreatePairTxInfo struct {
	TxType               uint8
	PairIndex            int64
	AssetAId             int64
	AssetBId             int64
	FeeRate              int64
	TreasuryAccountIndex int64
	TreasuryRate         int64
}

func ParseCreatePairTxInfo(txInfoStr string) (txInfo *CreatePairTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseCreatePairTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

type UpdatePairRateTxInfo struct {
	TxType               uint8
	PairIndex            int64
	FeeRate              int64
	TreasuryAccountIndex int64
	TreasuryRate         int64
}

func ParseUpdatePairRateTxInfo(txInfoStr string) (txInfo *UpdatePairRateTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseUpdatePairRateTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

type DepositTxInfo struct {
	TxType          uint8
	AccountIndex    int64
	AccountNameHash []byte
	AssetId         int64
	AssetAmount     *big.Int
}

func ParseDepositTxInfo(txInfoStr string) (txInfo *DepositTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseDepositTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

type DepositNftTxInfo struct {
	TxType              uint8
	AccountIndex        int64
	NftIndex            int64
	NftL1Address        string
	CreatorAccountIndex int64
	CreatorTreasuryRate int64
	NftContentHash      []byte
	NftL1TokenId        *big.Int
	AccountNameHash     []byte
	CollectionId        int64
}

func ParseDepositNftTxInfo(txInfoStr string) (txInfo *DepositNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseDepositNftTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

type FullExitTxInfo struct {
	TxType          uint8
	AccountIndex    int64
	AccountNameHash []byte
	AssetId         int64
	AssetAmount     *big.Int
}

func ParseFullExitTxInfo(txInfoStr string) (txInfo *FullExitTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseFullExitTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

type FullExitNftTxInfo struct {
	TxType                 uint8
	AccountIndex           int64
	CreatorAccountIndex    int64
	CreatorTreasuryRate    int64
	NftIndex               int64
	CollectionId           int64
	NftL1Address           string
	AccountNameHash        []byte
	CreatorAccountNameHash []byte
	NftContentHash         []byte
	NftL1TokenId           *big.Int
}

func ParseFullExitNftTxInfo(txInfoStr string) (txInfo *FullExitNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseFullExitNftTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseCreateCollectionTxInfo(txInfoStr string) (txInfo *CreateCollectionTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseCreateCollectionTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

// layer-2 transactions
func ParseTransferTxInfo(txInfoStr string) (txInfo *TransferTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseTransferTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseSwapTxInfo(txInfoStr string) (txInfo *SwapTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseSwapTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseAddLiquidityTxInfo(txInfoStr string) (txInfo *AddLiquidityTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseAddLiquidityTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseRemoveLiquidityTxInfo(txInfoStr string) (txInfo *RemoveLiquidityTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseRemoveLiquidityTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseMintNftTxInfo(txInfoStr string) (txInfo *MintNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseMintNftTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseTransferNftTxInfo(txInfoStr string) (txInfo *TransferNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseTransferNftTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseOfferTxInfo(txInfoStr string) (txInfo *OfferTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseOfferTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseAtomicMatchTxInfo(txInfoStr string) (txInfo *AtomicMatchTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseAtomicMatchTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseCancelOfferTxInfo(txInfoStr string) (txInfo *CancelOfferTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseCancelOfferTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseWithdrawTxInfo(txInfoStr string) (txInfo *WithdrawTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseWithdrawTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseWithdrawNftTxInfo(txInfoStr string) (txInfo *WithdrawNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseWithdrawNftTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}
