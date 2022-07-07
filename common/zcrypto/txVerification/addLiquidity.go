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
	"log"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zecrey-labs/zecrey-crypto/wasm/zecrey-legend/legendTxTypes"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
)

func VerifyAddLiquidityTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	liquidityInfo *LiquidityInfo,
	txInfo *AddLiquidityTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[liquidityInfo.TreasuryAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId].Balance.Cmp(ZeroBigInt) <= 0 ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetBId] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetBId].Balance.Cmp(ZeroBigInt) <= 0 ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == nil ||
		liquidityInfo == nil ||
		liquidityInfo.AssetAId != txInfo.AssetAId ||
		liquidityInfo.AssetBId != txInfo.AssetBId ||
		txInfo.AssetAAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.AssetBAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.LpAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Errorf("[VerifyAddLiquidityTxInfo] invalid params")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.FromAccountIndex].Nonce {
		log.Println("[VerifyAddLiquidityTxInfo] invalid nonce")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] invalid nonce")
	}
	// add tx info
	var (
		assetDeltaMap             = make(map[int64]map[int64]*big.Int)
		poolDeltaForToAccount     *LiquidityInfo
		lpDeltaForFromAccount     *big.Int
		lpDeltaForTreasuryAccount *big.Int
	)
	// init delta map
	assetDeltaMap[txInfo.FromAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// from account asset A
	assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetAId] = ffmath.Neg(txInfo.AssetAAmount)
	// from account asset B
	assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetBId] = ffmath.Neg(txInfo.AssetBAmount)
	// from account asset Gas
	if assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] == nil {
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	} else {
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] = ffmath.Sub(
			assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId],
			txInfo.GasFeeAssetAmount,
		)
	}
	poolAssetADelta := txInfo.AssetAAmount
	poolAssetBDelta := txInfo.AssetBAmount
	// from account lp
	lpDeltaForTreasuryAccount, err = util.ComputeSLp(liquidityInfo.AssetA, liquidityInfo.AssetB, liquidityInfo.KLast, liquidityInfo.FeeRate, liquidityInfo.TreasuryRate)
	if err != nil {
		logx.Errorf("[ComputeSLp] err: %v", err)
		return nil, err
	}
	poolLp := ffmath.Sub(liquidityInfo.LpAmount, lpDeltaForTreasuryAccount)
	// lp = \Delta{x}/x * poolLp
	if liquidityInfo.AssetA.Cmp(ZeroBigInt) == 0 {
		lpDeltaForFromAccount, err = util.CleanPackedAmount(new(big.Int).Sqrt(ffmath.Multiply(txInfo.AssetAAmount, txInfo.AssetBAmount)))
		if err != nil {
			logx.Errorf("[VerifyAddLiquidityTxInfo] unable to compute lp delta: %s", err.Error())
			return nil, err
		}
	} else {
		lpDeltaForFromAccount, err = util.CleanPackedAmount(ffmath.Div(ffmath.Multiply(poolAssetADelta, poolLp), liquidityInfo.AssetA))
		if err != nil {
			logx.Errorf("[VerifyAddLiquidityTxInfo] unable to compute lp delta: %s", err.Error())
			return nil, err
		}
	}
	// pool account pool info
	finalPoolA := ffmath.Add(liquidityInfo.AssetA, poolAssetADelta)
	finalPoolB := ffmath.Add(liquidityInfo.AssetB, poolAssetBDelta)
	poolDeltaForToAccount = &LiquidityInfo{
		PairIndex:            txInfo.PairIndex,
		AssetAId:             txInfo.AssetAId,
		AssetA:               poolAssetADelta,
		AssetBId:             txInfo.AssetBId,
		AssetB:               poolAssetBDelta,
		LpAmount:             lpDeltaForFromAccount,
		KLast:                ffmath.Multiply(finalPoolA, finalPoolB),
		FeeRate:              liquidityInfo.FeeRate,
		TreasuryAccountIndex: liquidityInfo.TreasuryAccountIndex,
		TreasuryRate:         liquidityInfo.TreasuryRate,
	}
	// set tx info
	txInfo.KLast, err = util.CleanPackedAmount(ffmath.Multiply(finalPoolA, finalPoolB))
	if err != nil {
		return nil, err
	}
	txInfo.TreasuryAmount = lpDeltaForTreasuryAccount
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
	// check asset A
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId].Balance.Cmp(txInfo.AssetAAmount) < 0 {
		logx.Errorf("[VerifyAddLiquidityTxInfo] you don't have enough balance of asset A")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] you don't have enough balance of asset A")
	}
	// check asset B
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetBId].Balance.Cmp(txInfo.AssetBAmount) < 0 {
		logx.Errorf("[VerifyAddLiquidityTxInfo] you don't have enough balance of asset B")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] you don't have enough balance of asset B")
	}
	// check lp amount
	if lpDeltaForFromAccount.Cmp(txInfo.LpAmount) < 0 {
		logx.Errorf("[VerifyAddLiquidityTxInfo] invalid lp amount")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] invalid lp amount")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeAddLiquidityMsgHash(txInfo, hFunc)
	if err != nil {
		logx.Errorf("[VerifyAddLiquidityTxInfo] unable to compute hash: %s", err.Error())
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
		log.Println("[VerifyAddLiquidityTxInfo] unable to verify signature:", err)
		return nil, err
	}
	if !isValid {
		log.Println("[VerifyAddLiquidityTxInfo] invalid signature")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] invalid signature")
	}
	// compute tx details
	// from account asset A
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetAId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetAId, ffmath.Neg(txInfo.AssetAAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// from account asset B
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetBId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetBId, ffmath.Neg(txInfo.AssetBAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
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
		Order:        order,
		AccountOrder: accountOrder,
	})
	// from account lp
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.PairIndex, ZeroBigInt, lpDeltaForFromAccount, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// pool info
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    LiquidityAssetType,
		AccountIndex: commonConstant.NilTxAccountIndex,
		AccountName:  commonConstant.NilAccountName,
		BalanceDelta: poolDeltaForToAccount.String(),
		Order:        order,
		AccountOrder: commonConstant.NilAccountOrder,
	})
	// treasury account
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    GeneralAssetType,
		AccountIndex: liquidityInfo.TreasuryAccountIndex,
		AccountName:  accountInfoMap[liquidityInfo.TreasuryAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.PairIndex, ZeroBigInt, lpDeltaForTreasuryAccount, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// gas account asset Gas
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
