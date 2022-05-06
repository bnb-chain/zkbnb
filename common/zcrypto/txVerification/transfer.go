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
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

/*
	VerifyTransferTx:
	accounts order is:
	- FromAccount
		- Assets
			- AssetA
			- AssetGas
	- ToAccount
		- Assets
			- AssetA
	- GasAccount
		- Assets
			- AssetGas
*/
func VerifyTransferTxInfo(
	accountInfoMap map[int64]*account.Account,
	assetInfoMap map[int64]map[int64]*asset.AccountAsset,
	txInfo *TransferTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[txInfo.ToAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		assetInfoMap[txInfo.FromAccountIndex] == nil ||
		assetInfoMap[txInfo.FromAccountIndex][txInfo.AssetId] == nil ||
		assetInfoMap[txInfo.ToAccountIndex] == nil ||
		assetInfoMap[txInfo.ToAccountIndex][txInfo.AssetId] == nil ||
		assetInfoMap[txInfo.GasAccountIndex] == nil ||
		assetInfoMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] == nil ||
		txInfo.AssetAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		return nil, errors.New("[VerifyTransferTxInfo] invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.FromAccountIndex].Nonce {
		logx.Errorf("[VerifyTransferTxInfo] invalid nonce")
		return nil, errors.New("[VerifyTransferTxInfo] invalid nonce")
	}
	// check asset assetABalance
	if txInfo.AssetId != assetInfoMap[txInfo.FromAccountIndex][txInfo.AssetId].AssetId ||
		txInfo.FromAccountIndex != accountInfoMap[txInfo.FromAccountIndex].AccountIndex ||
		txInfo.GasFeeAssetId != assetInfoMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId].AssetId {
		logx.Errorf("[VerifyTransferTxInfo] invalid params")
		return nil, errors.New("[VerifyTransferTxInfo] invalid params")
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
	// to account asset A
	if assetDeltaMap[txInfo.ToAccountIndex][txInfo.AssetId] != nil {
		assetDeltaMap[txInfo.ToAccountIndex][txInfo.AssetId] = ffmath.Add(assetDeltaMap[txInfo.ToAccountIndex][txInfo.AssetId], txInfo.AssetAmount)
	} else {
		assetDeltaMap[txInfo.ToAccountIndex][txInfo.AssetId] = txInfo.AssetAmount
	}
	// gas account asset gas
	if assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] != nil {
		assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] = ffmath.Add(assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId], txInfo.GasFeeAssetAmount)
	} else {
		assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] = txInfo.GasFeeAssetAmount
	}
	// check if from account has enough assetABalance
	// asset A
	assetABalance, isValid := new(big.Int).SetString(assetInfoMap[txInfo.FromAccountIndex][txInfo.AssetId].Balance, 10)
	if !isValid {
		logx.Errorf("[VerifyTransferTxInfo] invalid assetABalance")
		return nil, errors.New("[VerifyTransferTxInfo] invalid assetABalance")
	}
	if assetABalance.Cmp(assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetId]) < 0 {
		logx.Errorf("[VerifyTransferTxInfo] you don't have enough assetABalance")
		return nil, errors.New("[VerifyTransferTxInfo] you don't have enough assetABalance")
	}
	// asset Gas
	assetGasBalance, isValid := new(big.Int).SetString(assetInfoMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId].Balance, 10)
	if !isValid {
		logx.Errorf("[VerifyTransferTxInfo] invalid assetGasBalance")
		return nil, errors.New("[VerifyTransferTxInfo] invalid assetGasBalance")
	}
	if assetGasBalance.Cmp(assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId]) < 0 {
		logx.Errorf("[VerifyTransferTxInfo] you don't have enough assetGasBalance")
		return nil, errors.New("[VerifyTransferTxInfo] you don't have enough assetGasBalance")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	hFunc.Write([]byte(txInfo.CallData))
	callDataHash := hFunc.Sum(nil)
	txInfo.CallDataHash = callDataHash
	msgHash := legendTxTypes.ComputeTransferMsgHash(txInfo, hFunc)
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.FromAccountIndex].PublicKey)
	if err != nil {
		return nil, err
	}
	isValid, err = pk.Verify(txInfo.Sig, msgHash, hFunc)
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
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		Balance:      assetInfoMap[txInfo.FromAccountIndex][txInfo.AssetId].Balance,
		BalanceDelta: assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetId].String(),
	})
	// from account asset gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		Balance:      assetInfoMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId].Balance,
		BalanceDelta: assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId].String(),
	})
	// to account asset a
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  accountInfoMap[txInfo.ToAccountIndex].AccountName,
		Balance:      assetInfoMap[txInfo.ToAccountIndex][txInfo.AssetId].Balance,
		BalanceDelta: assetDeltaMap[txInfo.ToAccountIndex][txInfo.AssetId].String(),
	})
	// gas account asset gas
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
