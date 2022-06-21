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

package txVerification

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

func VerifyTransferTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	txInfo *TransferTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.ToAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetId] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetId].Balance.Cmp(ZeroBigInt) <= 0 ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(ZeroBigInt) <= 0 ||
		txInfo.AssetAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		return nil, errors.New("[VerifyTransferTxInfo] invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.FromAccountIndex].Nonce {
		logx.Errorf("[VerifyTransferTxInfo] invalid nonce")
		return nil, errors.New("[VerifyTransferTxInfo] invalid nonce")
	}
	// init delta map
	var (
		assetDeltaMap = make(map[int64]map[int64]*big.Int)
	)
	assetDeltaMap[txInfo.FromAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.ToAccountIndex] == nil {
		assetDeltaMap[txInfo.ToAccountIndex] = make(map[int64]*big.Int)
	}
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// compute deltas
	// from account asset A
	assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetId] = ffmath.Neg(txInfo.AssetAmount)
	// from account asset Gas
	if assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] != nil {
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] = ffmath.Sub(assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId], txInfo.GasFeeAssetAmount)
	} else {
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	}
	// check if from account has enough assetABalance
	// asset A
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetId].Balance.Cmp(
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetId]) < 0 {
		logx.Errorf("[VerifyTransferTxInfo] you don't have enough assetABalance")
		return nil, errors.New("[VerifyTransferTxInfo] you don't have enough assetABalance")
	}
	// asset Gas
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId]) < 0 {
		logx.Errorf("[VerifyTransferTxInfo] you don't have enough assetGasBalance")
		return nil, errors.New("[VerifyTransferTxInfo] you don't have enough assetGasBalance")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	hFunc.Write([]byte(txInfo.CallData))
	callDataHash := hFunc.Sum(nil)
	txInfo.CallDataHash = callDataHash
	msgHash, err := legendTxTypes.ComputeTransferMsgHash(txInfo, hFunc)
	if err != nil {
		logx.Errorf("[VerifyTransferTxInfo] unable to compute hash: %s", err.Error())
		return nil, err
	}
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.FromAccountIndex].PublicKey)
	if err != nil {
		return nil, err
	}
	isValid, err := pk.Verify(txInfo.Sig, msgHash, hFunc)
	if err != nil {
		logx.Errorf("[VerifyTransferTxInfo] unable to verify signature:", err)
		return nil, err
	}
	if !isValid {
		logx.Errorf("[VerifyTransferTxInfo] invalid signature")
		return nil, errors.New("[VerifyTransferTxInfo] invalid signature")
	}
	// compute tx details
	// from account asset A
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetId, ffmath.Neg(txInfo.AssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	order++
	// from account asset gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// to account asset a
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  accountInfoMap[txInfo.ToAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetId, txInfo.AssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// gas account asset gas
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  accountInfoMap[txInfo.GasAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	return txDetails, nil
}
