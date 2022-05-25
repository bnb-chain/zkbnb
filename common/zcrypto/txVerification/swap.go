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
)

/*
	VerifySwapTx:
	accounts order is:
	- FromAccount
		- Assets:
			- AssetA
			- AssetB
			- AssetGas
	- GasAccount
		- Assets:
			- AssetGas
*/
func VerifySwapTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	liquidityInfo *LiquidityInfo,
	txInfo *SwapTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId].Balance.Cmp(ZeroBigInt) <= 0 ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(ZeroBigInt) <= 0 ||
		liquidityInfo == nil ||
		!((liquidityInfo.AssetAId == txInfo.AssetAId &&
			liquidityInfo.AssetBId == txInfo.AssetBId) ||
			(liquidityInfo.AssetBId == txInfo.AssetAId &&
				liquidityInfo.AssetAId == txInfo.AssetBId)) ||
		txInfo.AssetAAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.AssetBMinAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Errorf("[VerifySwapTxInfo] invalid params")
		return nil, errors.New("[VerifySwapTxInfo] invalid params")
	}
	// verify delta amount
	if txInfo.AssetBAmountDelta.Cmp(txInfo.AssetBMinAmount) < 0 {
		log.Println("[VerifySwapTxInfo] invalid swap amount")
		return nil, errors.New("[VerifySwapTxInfo] invalid swap amount")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.FromAccountIndex].Nonce {
		log.Println("[VerifySwapTxInfo] invalid nonce")
		return nil, errors.New("[VerifySwapTxInfo] invalid nonce")
	}
	var (
		//assetDeltaForTreasuryAccount *big.Int
		assetDeltaMap         = make(map[int64]map[int64]*big.Int)
		poolDeltaForToAccount *LiquidityInfo
	)
	// init delta map
	assetDeltaMap[txInfo.FromAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// from account asset A
	assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetAId] = ffmath.Neg(txInfo.AssetAAmount)
	// from account asset B
	assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetBId] = txInfo.AssetBAmountDelta
	// from account asset Gas
	if assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] == nil {
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	} else {
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] = ffmath.Sub(
			assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId],
			txInfo.GasFeeAssetAmount,
		)
	}
	if txInfo.AssetAAmount.Cmp(assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetAId]) < 0 {
		log.Println("[VerifySwapTxInfo] invalid treasury amount")
		return nil, errors.New("[VerifySwapTxInfo] invalid treasury amount")
	}
	// to account pool
	poolAssetADelta := txInfo.AssetAAmount
	poolAssetBDelta := ffmath.Neg(txInfo.AssetBAmountDelta)
	if txInfo.AssetAId == liquidityInfo.AssetAId {
		poolDeltaForToAccount = &LiquidityInfo{
			PairIndex:            txInfo.PairIndex,
			AssetAId:             txInfo.AssetAId,
			AssetA:               poolAssetADelta,
			AssetBId:             txInfo.AssetBId,
			AssetB:               poolAssetBDelta,
			LpAmount:             ZeroBigInt,
			KLast:                ZeroBigInt,
			FeeRate:              liquidityInfo.FeeRate,
			TreasuryAccountIndex: liquidityInfo.TreasuryAccountIndex,
			TreasuryRate:         liquidityInfo.TreasuryRate,
		}
	} else if txInfo.AssetAId == liquidityInfo.AssetBId {
		poolDeltaForToAccount = &LiquidityInfo{
			PairIndex:            txInfo.PairIndex,
			AssetAId:             txInfo.AssetBId,
			AssetA:               poolAssetBDelta,
			AssetBId:             txInfo.AssetAId,
			AssetB:               poolAssetADelta,
			LpAmount:             ZeroBigInt,
			KLast:                ZeroBigInt,
			FeeRate:              liquidityInfo.FeeRate,
			TreasuryAccountIndex: liquidityInfo.TreasuryAccountIndex,
			TreasuryRate:         liquidityInfo.TreasuryRate,
		}
	} else {
		log.Println("[VerifySwapTxInfo] invalid pool")
		return nil, errors.New("[VerifySwapTxInfo] invalid pool")
	}
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
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId].Balance.Cmp(
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetAId]) < 0 {
		logx.Errorf("[VerifySwapTxInfo] you don't have enough balance of asset A")
		return nil, errors.New("[VerifySwapTxInfo] you don't have enough balance of asset A")
	}
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId]) < 0 {
		logx.Errorf("[VerifySwapTxInfo] you don't have enough balance of asset Gas")
		return nil, errors.New("[VerifySwapTxInfo] you don't have enough balance of asset Gas")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeSwapMsgHash(txInfo, hFunc)
	if err != nil {
		logx.Errorf("[VerifySwapTxInfo] unable to compute hash: %s", err.Error())
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
		log.Println("[VerifySwapTxInfo] unable to verify signature:", err)
		return nil, err
	}
	if !isValid {
		log.Println("[VerifySwapTxInfo] invalid signature")
		return nil, errors.New("[VerifySwapTxInfo] invalid signature")
	}
	// compute tx details
	// from account asset A
	order := int64(0)
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetAId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetAId, ffmath.Neg(txInfo.AssetAAmount), ZeroBigInt, ZeroBigInt).String(),
		Order: order,
	})
	// from account asset B
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetBId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetBId, txInfo.AssetBAmountDelta, ZeroBigInt, ZeroBigInt).String(),
		Order: order,
	})
	// from account asset Gas
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order: order,
	})
	// pool info
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    LiquidityAssetType,
		AccountIndex: commonConstant.NilAccountIndex,
		AccountName:  commonConstant.NilAccountName,
		BalanceDelta: poolDeltaForToAccount.String(),
		Order:        order,
	})
	// gas account asset Gas
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
