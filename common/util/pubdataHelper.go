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
	"bytes"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	curve "github.com/bnb-chain/zkbas-crypto/ecc/ztwistededwards/tebn254"
	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
)

func ConvertTxToRegisterZNSPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeRegisterZns {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToRegisterZNSPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseRegisterZnsTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(AccountNameToBytes32(txInfo.AccountName)))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	pk, err := ParsePubKey(txInfo.PubKey)
	if err != nil {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] unable to parse pub key: %s", err.Error())
		return nil, err
	}
	// because we can get Y from X, so we only need to store X is enough
	buf.Write(PrefixPaddingBufToChunkSize(pk.A.X.Marshal()))
	buf.Write(PrefixPaddingBufToChunkSize(pk.A.Y.Marshal()))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToCreatePairPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeCreatePair {
		logx.Errorf("[ConvertTxToCreatePairPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToCreatePairPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseCreatePairTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToCreatePairPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetAId)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetBId)))
	buf.Write(Uint16ToBytes(uint16(txInfo.FeeRate)))
	buf.Write(Uint32ToBytes(uint32(txInfo.TreasuryAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.TreasuryRate)))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToUpdatePairRatePubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeUpdatePairRate {
		logx.Errorf("[ConvertTxToUpdatePairRatePubData] invalid tx type")
		return nil, errors.New("[ConvertTxToUpdatePairRatePubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseUpdatePairRateTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToUpdatePairRatePubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.FeeRate)))
	buf.Write(Uint32ToBytes(uint32(txInfo.TreasuryAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.TreasuryRate)))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToDepositPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeDeposit {
		logx.Errorf("[ConvertTxToDepositPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToDepositPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseDepositTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(Uint128ToBytes(txInfo.AssetAmount))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToDepositNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeDepositNft {
		logx.Errorf("[ConvertTxToDepositNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToDepositNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseDepositNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToTransferPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeTransfer {
		logx.Errorf("[ConvertTxToTransferPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToTransferPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseTransferTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	packedAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	pubData = buf.Bytes()
	return pubData, nil
}

func ConvertTxToSwapPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeSwap {
		logx.Errorf("[ConvertTxToSwapPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToSwapPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseSwapTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToSwapPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountDeltaBytes, err := AmountToPackedAmountBytes(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetBAmountDeltaBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToAddLiquidityPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeAddLiquidity {
		logx.Errorf("[ConvertTxToAddLiquidityPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToAddLiquidityPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToAddLiquidityPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetBAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetBAmountBytes)
	LpAmountBytes, err := AmountToPackedAmountBytes(txInfo.LpAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(LpAmountBytes)
	KLastBytes, err := AmountToPackedAmountBytes(txInfo.KLast)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(KLastBytes)
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	treasuryAmountBytes, err := AmountToPackedAmountBytes(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(treasuryAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToRemoveLiquidityPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeRemoveLiquidity {
		logx.Errorf("[ConvertTxToRemoveLiquidityPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToRemoveLiquidityPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseRemoveLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToRemoveLiquidityPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetBAmountBytes)
	LpAmountBytes, err := AmountToPackedAmountBytes(txInfo.LpAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(LpAmountBytes)
	KLastBytes, err := AmountToPackedAmountBytes(txInfo.KLast)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(KLastBytes)
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	treasuryAmountBytes, err := AmountToPackedAmountBytes(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(treasuryAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToWithdrawPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeWithdraw {
		logx.Errorf("[ConvertTxToWithdrawPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToWithdrawPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseWithdrawTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToWithdrawPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(AddressStrToBytes(txInfo.ToAddress))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(Uint128ToBytes(txInfo.AssetAmount))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToCreateCollectionPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeCreateCollection {
		logx.Errorf("[ConvertTxToCreateCollectionPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToCreateCollectionPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseCreateCollectionTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToCreateCollectionPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CollectionId)))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToMintNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeMintNft {
		logx.Errorf("[ConvertTxToMintNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToMintNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseMintNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToMintNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(Uint16ToBytes(uint16(txInfo.NftCollectionId)))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(common.FromHex(txInfo.NftContentHash)))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToTransferNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeTransferNft {
		logx.Errorf("[ConvertTxToMintNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToMintNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseTransferNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToMintNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToAtomicMatchPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeAtomicMatch {
		logx.Errorf("[ConvertTxToAtomicMatchPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToAtomicMatchPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseAtomicMatchTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToAtomicMatchPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.BuyOffer.AccountIndex)))
	buf.Write(Uint24ToBytes(txInfo.BuyOffer.OfferId))
	buf.Write(Uint32ToBytes(uint32(txInfo.SellOffer.AccountIndex)))
	buf.Write(Uint24ToBytes(txInfo.SellOffer.OfferId))
	buf.Write(Uint40ToBytes(txInfo.BuyOffer.NftIndex))
	buf.Write(Uint16ToBytes(uint16(txInfo.SellOffer.AssetId)))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	packedAmountBytes, err := AmountToPackedAmountBytes(txInfo.BuyOffer.AssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAmountBytes)
	creatorAmountBytes, err := AmountToPackedAmountBytes(txInfo.CreatorAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(creatorAmountBytes)
	treasuryAmountBytes, err := AmountToPackedAmountBytes(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(treasuryAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToCancelOfferPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeCancelOffer {
		logx.Errorf("[ConvertTxToCancelOfferPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToCancelOfferPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseCancelOfferTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToCancelOfferPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint24ToBytes(txInfo.OfferId))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToWithdrawNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeWithdrawNft {
		logx.Errorf("[ConvertTxToWithdrawNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToWithdrawNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseWithdrawNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToWithdrawNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(AddressStrToBytes(txInfo.ToAddress))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk3 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(chunk3)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	return buf.Bytes(), nil
}

func ConvertTxToFullExitPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeFullExit {
		logx.Errorf("[ConvertTxToFullExitPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToFullExitPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseFullExitTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToFullExitPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(Uint128ToBytes(txInfo.AssetAmount))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}

func ConvertTxToFullExitNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeFullExitNft {
		logx.Errorf("[ConvertTxToFullExitNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToFullExitNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseFullExitNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToFullExitNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	return buf.Bytes(), nil
}

// create block commitment
func CreateBlockCommitment(
	currentBlockHeight int64,
	createdAt int64,
	oldStateRoot []byte,
	newStateRoot []byte,
	pubData []byte,
	onChainOpsCount int64,
) string {
	var buf bytes.Buffer
	PaddingInt64IntoBuf(&buf, currentBlockHeight)
	PaddingInt64IntoBuf(&buf, createdAt)
	buf.Write(CleanAndPaddingByteByModulus(oldStateRoot))
	buf.Write(CleanAndPaddingByteByModulus(newStateRoot))
	buf.Write(CleanAndPaddingByteByModulus(pubData))
	PaddingInt64IntoBuf(&buf, onChainOpsCount)
	// TODO Keccak256
	//hFunc := mimc.NewMiMC()
	//hFunc.Write(buf.Bytes())
	//commitment := hFunc.Sum(nil)
	commitment := KeccakHash(buf.Bytes())
	return common.Bytes2Hex(commitment)
}

func CleanAndPaddingByteByModulus(buf []byte) []byte {
	if len(buf) <= 32 {
		return ffmath.Mod(new(big.Int).SetBytes(buf), curve.Modulus).FillBytes(make([]byte, 32))
	}
	offset := 32
	var pendingBuf bytes.Buffer
	for offset <= len(buf) {
		pendingBuf.Write(ffmath.Mod(new(big.Int).SetBytes(buf[offset-32:offset]), curve.Modulus).FillBytes(make([]byte, 32)))
		offset += 32
	}
	return pendingBuf.Bytes()
}
