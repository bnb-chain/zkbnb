/*
 * Copyright © 2021 Zkbas Protocol
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
	"fmt"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
)

func VerifyCancelOfferTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	txInfo *CancelOfferTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.AccountIndex] == nil ||
		accountInfoMap[txInfo.AccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.GasFeeAssetId] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Error("invalid params")
		return nil, errors.New("invalid params")
	}
	if accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.OfferId/OfferPerAsset] == nil {
		accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.OfferId/OfferPerAsset] = &commonAsset.AccountAsset{
			AssetId:                  txInfo.OfferId / OfferPerAsset,
			Balance:                  ZeroBigInt,
			LpAmount:                 ZeroBigInt,
			OfferCanceledOrFinalized: ZeroBigInt,
		}
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.AccountIndex].Nonce {
		logx.Errorf("invalid nonce, actual: %d, expected: %d",
			txInfo.Nonce, accountInfoMap[txInfo.AccountIndex].Nonce)
		return nil, fmt.Errorf("invalid nonce, actual: %d, expected: %d",
			txInfo.Nonce, accountInfoMap[txInfo.AccountIndex].Nonce)
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
		new(big.Int).Abs(assetDeltaMap[txInfo.AccountIndex][txInfo.GasFeeAssetId])) < 0 {
		logx.Errorf("not enough balance of gas")
		return nil, errors.New("not enough balance of gas")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeCancelOfferMsgHash(txInfo, hFunc)
	if err != nil {
		logx.Errorf("unable to compute tx hash: %s", err.Error())
		return nil, errors.New("internal error")
	}
	// verify signature
	if err := VerifySignature(txInfo.Sig, msgHash, accountInfoMap[txInfo.AccountIndex].PublicKey); err != nil {
		return nil, err
	}
	// compute tx details
	// from account asset gas
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfoMap[txInfo.AccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// from account offer id
	offerAssetId := txInfo.OfferId / OfferPerAsset
	offerIndex := txInfo.OfferId % OfferPerAsset
	oOffer := accountInfoMap[txInfo.AccountIndex].AssetInfo[offerAssetId].OfferCanceledOrFinalized
	// verify whether account offer id is valid for use
	if oOffer.Bit(int(offerIndex)) == 1 {
		logx.Errorf("account %d offer index %d is already in use", txInfo.AccountIndex, offerIndex)
		return nil, errors.New("invalid offer id")
	}
	nOffer := new(big.Int).SetBit(oOffer, int(offerIndex), 1)
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      offerAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfoMap[txInfo.AccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			offerAssetId, ZeroBigInt, ZeroBigInt, nOffer,
		).String(),
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
