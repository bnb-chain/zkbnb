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
	VerifySwapTx:
	accounts order is:
	- FromAccount
		- Assets:
			- AssetA
			- AssetB
			- AssetGas
	- ToAccount
		- Liquidity
			- AssetA
			- AssetB
	- TreasuryAccount
		- Assets
			- AssetA
	- GasAccount
		- Assets:
			- AssetGas
*/
func VerifySwapTxInfo(
	accountInfoMap map[int64]*commonAsset.FormatAccountInfo,
	txInfo *SwapTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[txInfo.ToAccountIndex] == nil ||
		accountInfoMap[txInfo.TreasuryAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId] == "" ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId] == util.ZeroBigInt.String() ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == "" ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == util.ZeroBigInt.String() ||
		accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo == nil ||
		accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex] == nil ||
		!((accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId == txInfo.AssetAId &&
			accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId == txInfo.AssetBId) ||
			(accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId == txInfo.AssetAId &&
				accountInfoMap[txInfo.ToAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId == txInfo.AssetBId)) ||
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
		assetDeltaMap         = make(map[int64]map[int64]*big.Int)
		poolDeltaForToAccount *PoolInfo
	)
	// init delta map
	assetDeltaMap[txInfo.FromAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.TreasuryAccountIndex] == nil {
		assetDeltaMap[txInfo.TreasuryAccountIndex] = make(map[int64]*big.Int)
	}
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// verify treasury amount
	assetDeltaMap[txInfo.TreasuryAccountIndex][txInfo.AssetAId], err = util.CleanPackedFee(
		ffmath.Div(
			ffmath.Multiply(
				txInfo.AssetAAmount,
				big.NewInt(txInfo.TreasuryRate)),
			big.NewInt(int64(TenThousand))))
	if err != nil {
		logx.Errorf("[VerifySwapTxInfo] unable to compute treasury amount: %s", err.Error())
		return nil, err
	}
	if txInfo.AssetAAmount.Cmp(assetDeltaMap[txInfo.TreasuryAccountIndex][txInfo.AssetAId]) < 0 {
		log.Println("[VerifySwapTxInfo] invalid treasury amount")
		return nil, errors.New("[VerifySwapTxInfo] invalid treasury amount")
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
	// to account pool
	poolAssetADelta := ffmath.Sub(txInfo.AssetAAmount, assetDeltaMap[txInfo.TreasuryAccountIndex][txInfo.AssetAId])
	poolAssetBDelta := ffmath.Neg(txInfo.AssetBAmountDelta)
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
	assetABalance, isValid := new(big.Int).SetString(accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId], 10)
	if !isValid {
		logx.Errorf("[VerifySwapTxInfo] unable to parse balance")
		return nil, errors.New("[VerifySwapTxInfo] unable to parse balance")
	}
	assetGasBalance, isValid := new(big.Int).SetString(accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId], 10)
	if !isValid {
		logx.Errorf("[VerifySwapTxInfo] unable to parse balance")
		return nil, errors.New("[VerifySwapTxInfo] unable to parse balance")
	}
	if assetABalance.Cmp(assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetAId]) < 0 {
		logx.Errorf("[VerifySwapTxInfo] you don't have enough balance of asset A")
		return nil, errors.New("[VerifySwapTxInfo] you don't have enough balance of asset A")
	}
	if assetGasBalance.Cmp(assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId]) < 0 {
		logx.Errorf("[VerifySwapTxInfo] you don't have enough balance of asset Gas")
		return nil, errors.New("[VerifySwapTxInfo] you don't have enough balance of asset Gas")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash := legendTxTypes.ComputeSwapMsgHash(txInfo, hFunc)
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.FromAccountIndex].PublicKey)
	if err != nil {
		return nil, err
	}
	isValid, err = pk.Verify(txInfo.Sig, msgHash, hFunc)
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
	// pool account pool info
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    LiquidityAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  accountInfoMap[txInfo.ToAccountIndex].AccountName,
		BalanceDelta: poolDeltaForToAccount.String(),
	})
	// treasury account asset A
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetAId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.TreasuryAccountIndex,
		AccountName:  accountInfoMap[txInfo.TreasuryAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.TreasuryAccountIndex][txInfo.AssetAId].String(),
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
