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
	"github.com/zecrey-labs/zecrey-crypto/wasm/zecrey-legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

type (
	TransferTxInfo        = legendTxTypes.TransferTxInfo
	SwapTxInfo            = legendTxTypes.SwapTxInfo
	AddLiquidityTxInfo    = legendTxTypes.AddLiquidityTxInfo
	RemoveLiquidityTxInfo = legendTxTypes.RemoveLiquidityTxInfo
	WithdrawTxInfo        = legendTxTypes.WithdrawTxInfo
	MintNftTxInfo         = legendTxTypes.MintNftTxInfo
	TransferNftTxInfo     = legendTxTypes.TransferNftTxInfo
	SetNftPriceTxInfo     = legendTxTypes.SetNftPriceTxInfo
	BuyNftTxInfo          = legendTxTypes.BuyNftTxInfo
	WithdrawNftTxInfo     = legendTxTypes.WithdrawNftTxInfo
)

type RegisterZnsTxInfo struct {
	TxType      uint8
	AccountName string
	PubKey      string
}

func ParseRegisterZnsTxInfo(txInfoStr string) (txInfo *RegisterZnsTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseRegisterZnsTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

type DepositTxInfo struct {
	TxType          uint8
	AccountIndex    uint32
	AccountNameHash string
	AssetId         uint16
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
	TxType          uint8
	AccountIndex    uint32
	AccountNameHash string
	NftType         uint8
	NftIndex        uint64
	NftContentHash  []byte
	NftL1Address    string
	NftL1TokenId    *big.Int
	Amount          uint32
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
	AccountIndex    uint32
	AccountNameHash string
	AssetId         uint16
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
	TxType          uint8
	AccountIndex    uint32
	AccountNameHash string
	NftL1Address    string
	ToAddress       string
	ProxyAddress    string
	NftType         uint8
	NftL1TokenId    *big.Int
	Amount          uint32
	NftContentHash  []byte
	NftIndex        uint64
}

func ParseFullExitNftTxInfo(txInfoStr string) (txInfo *FullExitNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseFullExitNftTxInfo] unable to parse tx info: %s", err.Error())
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
		logx.Errorf("[ParseAddLiquidityTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseMintNftTxInfo(txInfoStr string) (txInfo *MintNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseAddLiquidityTxInfo] unable to parse tx info: %s", err.Error())
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

func ParseSetNftPriceTxInfo(txInfoStr string) (txInfo *SetNftPriceTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseSetNftPriceTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}

func ParseBuyNftTxInfo(txInfoStr string) (txInfo *BuyNftTxInfo, err error) {
	err = json.Unmarshal([]byte(txInfoStr), &txInfo)
	if err != nil {
		logx.Errorf("[ParseBuyNftTxInfo] unable to parse tx info: %s", err.Error())
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
		logx.Errorf("[ParseWithdrawTxInfo] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	return txInfo, nil
}
