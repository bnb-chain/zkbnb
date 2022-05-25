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
	"encoding/json"
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
	VerifyRemoveLiquidityTx:
	accounts order is:
	- FromAccount
		- Assets:
			- AssetA
			- AssetB
			- AssetGas
			- LpAmount
	- TreasuryAccount
		- Assets:
			- LpAmount
	- GasAccount
		- Assets:
			- AssetGas
*/
func VerifyRemoveLiquidityTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	liquidityInfo *LiquidityInfo,
	txInfo *RemoveLiquidityTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[liquidityInfo.TreasuryAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		liquidityInfo == nil ||
		liquidityInfo.AssetAId != txInfo.AssetAId ||
		liquidityInfo.AssetBId != txInfo.AssetBId ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.PairIndex] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.PairIndex].LpAmount.Cmp(ZeroBigInt) <= 0 ||
		txInfo.AssetAMinAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.AssetBMinAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.LpAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		infoBytes, _ := json.Marshal(accountInfoMap)
		log.Println(string(infoBytes))
		logx.Errorf("[VerifyRemoveLiquidityTxInfo] invalid params")
		return nil, errors.New("[VerifyRemoveLiquidityTxInfo] invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.FromAccountIndex].Nonce {
		log.Println("[VerifyRemoveLiquidityTxInfo] invalid nonce")
		return nil, errors.New("[VerifyRemoveLiquidityTxInfo] invalid nonce")
	}
	// add tx info
	var (
		assetDeltaMap             = make(map[int64]map[int64]*big.Int)
		lpDeltaForFromAccount     *big.Int
		lpDeltaForTreausryAccount *big.Int
		poolDeltaForToAccount     *LiquidityInfo
	)
	// init delta map
	assetDeltaMap[txInfo.FromAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// from account asset A
	assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetAId] = txInfo.AssetAAmountDelta
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
	// from account lp amount
	lpDeltaForFromAccount = ffmath.Neg(txInfo.LpAmount)
	// pool account pool info
	poolAssetADelta := ffmath.Neg(txInfo.AssetAAmountDelta)
	poolAssetBDelta := ffmath.Neg(txInfo.AssetBAmountDelta)
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
	// treasury account
	lpDeltaForTreausryAccount = commonAsset.ComputeSLp(
		liquidityInfo.AssetA,
		liquidityInfo.AssetB,
		liquidityInfo.KLast,
		liquidityInfo.FeeRate,
		liquidityInfo.TreasuryRate,
	)
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
	// check lp amount
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.PairIndex].LpAmount.Cmp(txInfo.LpAmount) < 0 {
		logx.Errorf("[VerifyRemoveLiquidityTxInfo] invalid lp amount")
		return nil, errors.New("[VerifyRemoveLiquidityTxInfo] invalid lp amount")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeRemoveLiquidityMsgHash(txInfo, hFunc)
	if err != nil {
		logx.Errorf("[VerifyRemoveLiquidityTxInfo] unable to compute hash: %s", err.Error())
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
		log.Println("[VerifyRemoveLiquidityTxInfo] unable to verify signature:", err)
		return nil, err
	}
	if !isValid {
		log.Println("[VerifyRemoveLiquidityTxInfo] invalid signature")
		return nil, errors.New("[VerifyRemoveLiquidityTxInfo] invalid signature")
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
			txInfo.AssetAId, txInfo.AssetAAmountDelta, ZeroBigInt, ZeroBigInt).String(),
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
	// from account lp
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.PairIndex, ZeroBigInt, lpDeltaForFromAccount, ZeroBigInt).String(),
		Order: order,
	})
	// treasury account
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    GeneralAssetType,
		AccountIndex: liquidityInfo.TreasuryAccountIndex,
		AccountName:  accountInfoMap[liquidityInfo.TreasuryAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.PairIndex, ZeroBigInt, lpDeltaForTreausryAccount, ZeroBigInt,
		).String(),
		Order: order,
	})
	// pool account pool info
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
