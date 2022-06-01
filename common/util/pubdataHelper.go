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
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zeromicro/go-zero/core/logx"
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
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(AccountNameToBytes32(txInfo.AccountName))
	buf.Write(txInfo.AccountNameHash)
	pk, err := ParsePubKey(txInfo.PubKey)
	if err != nil {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] unable to parse pub key: %s", err.Error())
		return nil, err
	}
	// because we can get Y from X, so we only need to store X is enough
	buf.Write(pk.A.X.Marshal())
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk1 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(txInfo.AccountNameHash)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk1 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk2 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(txInfo.NftContentHash)
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(txInfo.AccountNameHash)
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(txInfo.CallDataHash)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	buf.Write(packedAssetBAmountBytes)
	buf.Write(LpAmountBytes)
	KLastBytes, err := AmountToPackedAmountBytes(txInfo.KLast)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(KLastBytes)
	chunk1 := PaddingBufToChunkSize(buf.Bytes())
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
	chunk2 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	buf.Write(packedAssetBAmountBytes)
	buf.Write(LpAmountBytes)
	KLastBytes, err := AmountToPackedAmountBytes(txInfo.KLast)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(KLastBytes)
	chunk1 := PaddingBufToChunkSize(buf.Bytes())
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
	chunk2 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk1 := PaddingBufToChunkSize(buf.Bytes())
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
	chunk2 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	buf.Write(Uint64ToBytes(uint64(txInfo.NftIndex)))
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
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(common.FromHex(txInfo.NftContentHash))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	buf.Write(Uint64ToBytes(uint64(txInfo.NftIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(txInfo.CallDataHash)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	buf.Write(Uint16ToBytes(uint16(txInfo.SellOffer.AssetId)))
	chunk1 := PaddingBufToChunkSize(buf.Bytes())
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
	chunk2 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk1 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := PaddingBufToChunkSize(buf.Bytes())
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
	chunk3 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(chunk3)
	buf.Write(txInfo.NftContentHash)
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(txInfo.CreatorAccountNameHash)
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
	chunk := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(txInfo.AccountNameHash)
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
	buf.Write(PaddingBufToChunkSize([]byte{}))
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
	chunk1 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := PaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(txInfo.AccountNameHash)
	buf.Write(txInfo.CreatorAccountNameHash)
	buf.Write(txInfo.NftContentHash)
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	return buf.Bytes(), nil
}

// TODO create block commitment
func CreateBlockCommitment(lastBlockHeight, currentBlockHeight int64, pubdata []byte) string {
	var buf bytes.Buffer
	buf.Write(Int64ToBytes(lastBlockHeight))
	buf.Write(Int64ToBytes(currentBlockHeight))
	buf.Write(pubdata)
	hFunc := mimc.NewMiMC()
	hFunc.Write(buf.Bytes())
	commitment := hFunc.Sum(nil)
	return common.Bytes2Hex(commitment)
}
