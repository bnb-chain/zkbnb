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
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"math/big"
)

/*
	VerifyAddLiquidityTx:
	accounts order is:
	- FromAccount
		- Assets:
			- AssetA
			- AssetB
			- AssetGas
		- Liquidity:
			- LpAmount
	- ToAccount
		- Liquidity
			- AssetA
			- AssetB
	- GasAccount
		- Assets
			- AssetGas
*/
func VerifyAddLiquidityTxInfo(
	accountInfoMap map[int64]*commonAsset.FormatAccountInfo,
	txInfo *AddLiquidityTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[txInfo.ToAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId] == "" ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId] == util.ZeroBigInt.String() ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetBId] == "" ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetBId] == util.ZeroBigInt.String() ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == "" ||
		accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo == nil ||
		accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex] == nil ||
		!((accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId == txInfo.AssetAId &&
			accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId == txInfo.AssetBId) ||
			(accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId == txInfo.AssetAId &&
				accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId == txInfo.AssetBId)) ||
		accountInfoMap[txInfo.FromAccountIndex].LiquidityInfo == nil ||
		accountInfoMap[txInfo.FromAccountIndex].LiquidityInfo[txInfo.PairIndex] == nil ||
		!((accountInfoMap[txInfo.FromAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId == txInfo.AssetAId &&
			accountInfoMap[txInfo.FromAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId == txInfo.AssetBId) ||
			(accountInfoMap[txInfo.FromAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId == txInfo.AssetAId &&
				accountInfoMap[txInfo.FromAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId == txInfo.AssetBId)) ||
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
		assetDeltaMap         = make(map[int64]map[int64]*big.Int)
		poolDeltaForToAccount *PoolInfo
		lpDeltaForFromAccount *big.Int
	)
	// init delta map
	assetDeltaMap[txInfo.FromAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.ToAccountIndex] == nil {
		assetDeltaMap[txInfo.ToAccountIndex] = make(map[int64]*big.Int)
	}
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
	// from account lp
	lpDeltaForFromAccount, err = util.CleanPackedAmount(new(big.Int).Sqrt(ffmath.Multiply(txInfo.AssetAAmount, txInfo.AssetBAmount)))
	if err != nil {
		logx.Errorf("[VerifyAddLiquidityTxInfo] unable to compute lp delta: %s", err.Error())
		return nil, err
	}
	// pool account pool info
	poolAssetADelta := txInfo.AssetAAmount
	poolAssetBDelta := txInfo.AssetBAmount
	if txInfo.AssetAId == accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId {
		poolDeltaForToAccount = &PoolInfo{
			AssetAAmount: poolAssetADelta,
			AssetBAmount: poolAssetBDelta,
		}
	} else if txInfo.AssetAId == accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId {
		poolDeltaForToAccount = &PoolInfo{
			AssetAAmount: poolAssetBDelta,
			AssetBAmount: poolAssetADelta,
		}
	} else {
		log.Println("[VerifyAddLiquidityTxInfo] invalid pool")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] invalid pool")
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
	assetABalance, isValid := new(big.Int).SetString(accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId], 10)
	if !isValid {
		logx.Errorf("[VerifyAddLiquidityTxInfo] unable to parse balance")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] unable to parse balance")
	}
	assetBBalance, isValid := new(big.Int).SetString(accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetBId], 10)
	if !isValid {
		logx.Errorf("[VerifyAddLiquidityTxInfo] unable to parse balance")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] unable to parse balance")
	}
	// check asset A
	if assetABalance.Cmp(txInfo.AssetAAmount) < 0 {
		logx.Errorf("[VerifyAddLiquidityTxInfo] you don't have enough balance of asset A")
		return nil, errors.New("[VerifyAddLiquidityTxInfo] you don't have enough balance of asset A")
	}
	// check asset B
	if assetBBalance.Cmp(txInfo.AssetBAmount) < 0 {
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
	msgHash := legendTxTypes.ComputeAddLiquidityMsgHash(txInfo, hFunc)
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.FromAccountIndex].PublicKey)
	if err != nil {
		return nil, err
	}
	isValid, err = pk.Verify(txInfo.Sig, msgHash, hFunc)
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
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetAId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetAId].String(),
	})
	// from account asset B
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetBId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetBId].String(),
	})
	// from account asset Gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId].String(),
	})
	// from account lp
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    LiquidityLpAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: lpDeltaForFromAccount.String(),
	})
	// pool account pool info
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    LiquidityAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  accountInfoMap[txInfo.ToAccountIndex].AccountName,
		BalanceDelta: poolDeltaForToAccount.String(),
	})
	// gas account asset Gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  accountInfoMap[txInfo.GasAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId].String(),
	})
	return txDetails, nil
}
