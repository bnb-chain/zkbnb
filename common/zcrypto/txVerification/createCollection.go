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
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zecrey-labs/zecrey-crypto/wasm/zecrey-legend/legendTxTypes"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"math/big"
	"strconv"
)

func VerifyCollectionTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	txInfo *CreateCollectionTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.AccountIndex] == nil ||
		accountInfoMap[txInfo.AccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.GasFeeAssetId] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Errorf("[VerifyCollectionTxInfo] invalid params")
		return nil, errors.New("[VerifyCollectionTxInfo] invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.AccountIndex].Nonce {
		log.Println("[VerifyCollectionTxInfo] invalid nonce")
		return nil, errors.New("[VerifyCollectionTxInfo] invalid nonce")
	}
	if txInfo.CollectionId != accountInfoMap[txInfo.AccountIndex].CollectionNonce {
		log.Println("[VerifyCollectionTxInfo] invalid nonce")
		return nil, errors.New("[VerifyCollectionTxInfo] invalid nonce")
	}
	// set tx info
	var (
		assetDeltaMap = make(map[int64]map[int64]*big.Int)
	)
	// init delta map
	assetDeltaMap[txInfo.AccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// from account asset Gas
	assetDeltaMap[txInfo.AccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	// gas account asset Gas
	if assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] == nil {
		assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] = txInfo.GasFeeAssetAmount
	} else {
		assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] = ffmath.Add(
			assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId],
			txInfo.GasFeeAssetAmount,
		)
	}
	// check balance
	if accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(
		assetDeltaMap[txInfo.AccountIndex][txInfo.GasFeeAssetId]) < 0 {
		logx.Errorf("[VerifyCollectionTxInfo] you don't have enough balance of asset Gas")
		return nil, errors.New("[VerifyCollectionTxInfo] you don't have enough balance of asset Gas")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash := legendTxTypes.ComputeCreateCollectionMsgHash(txInfo, hFunc)
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.AccountIndex].PublicKey)
	if err != nil {
		return nil, err
	}
	isValid, err := pk.Verify(txInfo.Sig, msgHash, hFunc)
	if err != nil {
		log.Println("[VerifyCollectionTxInfo] unable to verify signature:", err)
		return nil, err
	}
	if !isValid {
		log.Println("[VerifyCollectionTxInfo] invalid signature")
		return nil, errors.New("[VerifyCollectionTxInfo] invalid signature")
	}
	// compute tx details
	// from account collection nonce
	order := int64(0)
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      commonConstant.NilAssetId,
		AssetType:    CollectionNonceAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfoMap[txInfo.AccountIndex].AccountName,
		BalanceDelta: strconv.FormatInt(txInfo.CollectionId, 10),
		Order:        order,
	})
	// from account asset gas
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfoMap[txInfo.AccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order: order,
	})
	// gas account asset gas
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  accountInfoMap[txInfo.GasAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order: order,
	})
	return txDetails, nil
}
