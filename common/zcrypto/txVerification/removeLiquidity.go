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
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/asset"
	"github.com/zecrey-labs/zecrey-legend/common/util"
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
		- Liquidity:
			- LpAmount
	- ToAccount
		- Liquidity
			- AssetA
			- AssetB
	- GasAccount
		- Assets:
			- AssetGas
*/
func VerifyRemoveLiquidityTxInfo(
	accountInfoMap map[int64]*account.Account,
	assetInfoMap map[int64]map[int64]*asset.AccountAsset,
	liquidityInfoMap map[int64]map[int64]*asset.AccountLiquidity,
	txInfo *RemoveLiquidityTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[txInfo.ToAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		assetInfoMap[txInfo.FromAccountIndex] == nil ||
		assetInfoMap[txInfo.FromAccountIndex][txInfo.AssetAId] == nil ||
		assetInfoMap[txInfo.FromAccountIndex][txInfo.AssetBId] == nil ||
		assetInfoMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] == nil ||
		assetInfoMap[txInfo.GasAccountIndex] == nil ||
		assetInfoMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] == nil ||
		liquidityInfoMap[txInfo.ToAccountIndex] == nil ||
		liquidityInfoMap[txInfo.ToAccountIndex][txInfo.PairIndex] == nil ||
		!((liquidityInfoMap[txInfo.ToAccountIndex][txInfo.PairIndex].AssetAId == txInfo.AssetAId &&
			liquidityInfoMap[txInfo.ToAccountIndex][txInfo.PairIndex].AssetBId == txInfo.AssetBId) ||
			(liquidityInfoMap[txInfo.ToAccountIndex][txInfo.PairIndex].AssetBId == txInfo.AssetAId &&
				liquidityInfoMap[txInfo.ToAccountIndex][txInfo.PairIndex].AssetAId == txInfo.AssetBId)) ||
		liquidityInfoMap[txInfo.FromAccountIndex] == nil ||
		liquidityInfoMap[txInfo.FromAccountIndex][txInfo.PairIndex] == nil ||
		!((liquidityInfoMap[txInfo.FromAccountIndex][txInfo.PairIndex].AssetAId == txInfo.AssetAId &&
			liquidityInfoMap[txInfo.FromAccountIndex][txInfo.PairIndex].AssetBId == txInfo.AssetBId) ||
			(liquidityInfoMap[txInfo.FromAccountIndex][txInfo.PairIndex].AssetBId == txInfo.AssetAId &&
				liquidityInfoMap[txInfo.FromAccountIndex][txInfo.PairIndex].AssetAId == txInfo.AssetBId)) ||
		txInfo.AssetAMinAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.AssetBMinAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.LpAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
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
		assetDeltaMap         = make(map[int64]map[int64]*big.Int)
		lpDeltaForFromAccount *big.Int
		poolDeltaForToAccount *PoolInfo
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
	if txInfo.AssetAId == liquidityInfoMap[txInfo.ToAccountIndex][txInfo.PairIndex].AssetAId {
		poolDeltaForToAccount = &PoolInfo{
			AssetAAmount: poolAssetADelta,
			AssetBAmount: poolAssetBDelta,
		}
	} else if txInfo.AssetAId == liquidityInfoMap[txInfo.ToAccountIndex][txInfo.PairIndex].AssetBId {
		poolDeltaForToAccount = &PoolInfo{
			AssetAAmount: poolAssetBDelta,
			AssetBAmount: poolAssetADelta,
		}
	} else {
		log.Println("[VerifyRemoveLiquidityTxInfo] invalid pool")
		return nil, errors.New("[VerifyRemoveLiquidityTxInfo] invalid pool")
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
	// check lp amount
	lpBalance, isValid := new(big.Int).SetString(liquidityInfoMap[txInfo.FromAccountIndex][txInfo.PairIndex].LpAmount, 10)
	if !isValid {
		logx.Errorf("[VerifyRemoveLiquidityTxInfo] unable to parse balance")
		return nil, errors.New("[VerifyRemoveLiquidityTxInfo] unable to parse balance")
	}
	if lpBalance.Cmp(txInfo.LpAmount) < 0 {
		logx.Errorf("[VerifyRemoveLiquidityTxInfo] invalid lp amount")
		return nil, errors.New("[VerifyRemoveLiquidityTxInfo] invalid lp amount")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash := legendTxTypes.ComputeRemoveLiquidityMsgHash(txInfo, hFunc)
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.FromAccountIndex].PublicKey)
	if err != nil {
		return nil, err
	}
	isValid, err = pk.Verify(txInfo.Sig, msgHash, hFunc)
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
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetAId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		Balance:      assetInfoMap[txInfo.FromAccountIndex][txInfo.AssetAId].Balance,
		BalanceDelta: assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetAId].String(),
	})
	// from account asset B
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetBId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		Balance:      assetInfoMap[txInfo.FromAccountIndex][txInfo.AssetBId].Balance,
		BalanceDelta: assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetBId].String(),
	})
	// from account asset Gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		Balance:      assetInfoMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId].Balance,
		BalanceDelta: assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId].String(),
	})
	// from account lp
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    LiquidityLpAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		Balance:      liquidityInfoMap[txInfo.FromAccountIndex][txInfo.PairIndex].LpAmount,
		BalanceDelta: lpDeltaForFromAccount.String(),
	})
	// pool account pool info
	poolInfo, err := util.ConstructPoolInfo(
		liquidityInfoMap[txInfo.ToAccountIndex][txInfo.PairIndex].AssetA,
		liquidityInfoMap[txInfo.ToAccountIndex][txInfo.PairIndex].AssetB,
	)
	if err != nil {
		logx.Errorf("[VerifyRemoveLiquidityTxInfo] unable to construct pool info: %s", err.Error())
		return nil, err
	}
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    LiquidityAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  accountInfoMap[txInfo.ToAccountIndex].AccountName,
		Balance:      poolInfo.String(),
		BalanceDelta: poolDeltaForToAccount.String(),
	})
	// gas account asset Gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  accountInfoMap[txInfo.GasAccountIndex].AccountName,
		Balance:      assetInfoMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId].Balance,
		BalanceDelta: assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId].String(),
	})
	return txDetails, nil
}
