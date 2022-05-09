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
	"github.com/zecrey-labs/zecrey-core/common/general/model/nft"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zecrey-labs/zecrey-crypto/wasm/zecrey-legend/legendTxTypes"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"math/big"
)

/*
	VerifyBuyNftTx:
	accounts order is:
	- BuyerAccount
		- Assets
			- AssetA
			- AssetGas
		- Nft
			- empty
	- OwnerAccount
		- Assets
			- AssetA
		- Nft
			- nft index
	- TreasuryAccount
		- Assets
			- AssetA
	- GasAccount
		- Assets
			- AssetGas
*/
func VerifyBuyNftTxInfo(
	accountInfoMap map[int64]*commonAsset.FormatAccountInfo,
	nftInfoMap map[int64]*nft.L2Nft,
	txInfo *BuyNftTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.BuyerAccountIndex] == nil ||
		accountInfoMap[txInfo.OwnerAccountIndex] == nil ||
		accountInfoMap[txInfo.CreatorAccountIndex] == nil ||
		accountInfoMap[txInfo.TreasuryAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		accountInfoMap[txInfo.BuyerAccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.BuyerAccountIndex].AssetInfo[txInfo.AssetId] == "" ||
		accountInfoMap[txInfo.BuyerAccountIndex].AssetInfo[txInfo.AssetId] == util.ZeroBigInt.String() ||
		accountInfoMap[txInfo.BuyerAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == "" ||
		nftInfoMap[txInfo.OwnerAccountIndex] == nil ||
		nftInfoMap[txInfo.NftIndex].OwnerAccountIndex != txInfo.OwnerAccountIndex ||
		nftInfoMap[txInfo.NftIndex].CreatorAccountIndex != txInfo.CreatorAccountIndex ||
		nftInfoMap[txInfo.NftIndex].NftIndex != txInfo.NftIndex ||
		nftInfoMap[txInfo.NftIndex].NftContentHash != txInfo.NftContentHash ||
		txInfo.BuyerAccountIndex != txInfo.OwnerAccountIndex ||
		txInfo.AssetAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Errorf("[VerifySetNftPriceTxInfo] invalid params")
		return nil, errors.New("[VerifySetNftPriceTxInfo] invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.BuyerAccountIndex].Nonce {
		log.Println("[VerifyBuyNftTxInfo] invalid nonce")
		return nil, errors.New("[VerifyBuyNftTxInfo] invalid nonce")
	}
	// set tx info
	var (
		assetDeltaMap = make(map[int64]map[int64]*big.Int)
		newNftInfo    *NftInfo
	)
	// init delta map
	assetDeltaMap[txInfo.BuyerAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.OwnerAccountIndex] == nil {
		assetDeltaMap[txInfo.OwnerAccountIndex] = make(map[int64]*big.Int)
	}
	if assetDeltaMap[txInfo.CreatorAccountIndex] == nil {
		assetDeltaMap[txInfo.CreatorAccountIndex] = make(map[int64]*big.Int)
	}
	if assetDeltaMap[txInfo.TreasuryAccountIndex] == nil {
		assetDeltaMap[txInfo.TreasuryAccountIndex] = make(map[int64]*big.Int)
	}
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// buyer account asset A
	assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.AssetId] = ffmath.Neg(txInfo.AssetAmount)
	// buyer account asset Gas
	if assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.GasFeeAssetId] == nil {
		assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	} else {
		assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.GasFeeAssetId] = ffmath.Sub(
			assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.GasFeeAssetId],
			txInfo.GasFeeAssetAmount,
		)
	}
	// creator account asset A
	assetDeltaMap[txInfo.CreatorAccountIndex][txInfo.AssetId], err = util.CleanPackedFee(
		ffmath.Div(
			ffmath.Multiply(
				txInfo.AssetAmount,
				big.NewInt(txInfo.CreatorFeeRate)),
			big.NewInt(int64(TenThousand))))
	// treasury account asset A
	assetDeltaMap[txInfo.TreasuryAccountIndex][txInfo.AssetId], err = util.CleanPackedFee(
		ffmath.Div(
			ffmath.Multiply(
				txInfo.AssetAmount,
				big.NewInt(txInfo.TreasuryFeeRate)),
			big.NewInt(int64(TenThousand))))
	// owner account asset A
	ownerAssetADelta := ffmath.Sub(
		txInfo.AssetAmount,
		ffmath.Add(
			assetDeltaMap[txInfo.CreatorAccountIndex][txInfo.AssetId],
			assetDeltaMap[txInfo.TreasuryAccountIndex][txInfo.AssetId],
		),
	)
	if ownerAssetADelta.Cmp(ZeroBigInt) < 0 {
		logx.Errorf("[VerifyBuyNftTxInfo] invalid rate")
		return nil, errors.New("[VerifyBuyNftTxInfo] invalid rate")
	}
	if assetDeltaMap[txInfo.OwnerAccountIndex][txInfo.AssetId] == nil {
		assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.AssetId] = ownerAssetADelta
	} else {
		assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.AssetId] = ffmath.Add(
			assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.AssetId],
			ownerAssetADelta,
		)
	}
	// owner account nft info
	newNftInfo = &NftInfo{
		NftIndex:            nftInfoMap[txInfo.NftIndex].NftIndex,
		CreatorAccountIndex: nftInfoMap[txInfo.NftIndex].CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.OwnerAccountIndex,
		AssetId:             commonConstant.NilAssetId,
		AssetAmount:         commonConstant.NilAssetAmountStr,
		NftContentHash:      nftInfoMap[txInfo.NftIndex].NftContentHash,
		NftL1TokenId:        nftInfoMap[txInfo.NftIndex].NftL1TokenId,
		NftL1Address:        nftInfoMap[txInfo.NftIndex].NftL1Address,
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
	assetABalance, isValid := new(big.Int).SetString(accountInfoMap[txInfo.BuyerAccountIndex].AssetInfo[txInfo.AssetId], 10)
	if !isValid {
		logx.Errorf("[VerifySwapTxInfo] unable to parse balance")
		return nil, errors.New("[VerifySwapTxInfo] unable to parse balance")
	}
	assetGasBalance, isValid := new(big.Int).SetString(accountInfoMap[txInfo.BuyerAccountIndex].AssetInfo[txInfo.GasFeeAssetId], 10)
	if !isValid {
		logx.Errorf("[VerifySwapTxInfo] unable to parse balance")
		return nil, errors.New("[VerifySwapTxInfo] unable to parse balance")
	}
	if assetABalance.Cmp(assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.AssetId]) < 0 {
		logx.Errorf("[VerifySwapTxInfo] you don't have enough balance of asset A")
		return nil, errors.New("[VerifySwapTxInfo] you don't have enough balance of asset A")
	}
	if assetGasBalance.Cmp(assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.GasFeeAssetId]) < 0 {
		logx.Errorf("[VerifySwapTxInfo] you don't have enough balance of asset Gas")
		return nil, errors.New("[VerifySwapTxInfo] you don't have enough balance of asset Gas")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash := legendTxTypes.ComputeBuyNftMsgHash(txInfo, hFunc)
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.BuyerAccountIndex].PublicKey)
	if err != nil {
		return nil, err
	}
	isValid, err = pk.Verify(txInfo.Sig, msgHash, hFunc)
	if err != nil {
		log.Println("[VerifyBuyNftTxInfo] unable to verify signature:", err)
		return nil, err
	}
	if !isValid {
		log.Println("[VerifyBuyNftTxInfo] invalid signature")
		return nil, errors.New("[VerifyBuyNftTxInfo] invalid signature")
	}
	// compute tx details
	// buyer account asset A
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.BuyerAccountIndex,
		AccountName:  accountInfoMap[txInfo.BuyerAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.AssetId].String(),
	})
	// buyer account asset gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.BuyerAccountIndex,
		AccountName:  accountInfoMap[txInfo.BuyerAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.BuyerAccountIndex][txInfo.GasFeeAssetId].String(),
	})
	// buyer account nft delta
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    NftAssetType,
		AccountIndex: txInfo.BuyerAccountIndex,
		AccountName:  accountInfoMap[txInfo.BuyerAccountIndex].AccountName,
		BalanceDelta: newNftInfo.String(),
	})
	// creator account asset A
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		AccountName:  accountInfoMap[txInfo.CreatorAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.CreatorAccountIndex][txInfo.AssetId].String(),
	})
	// treasury account asset A
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.TreasuryAccountIndex,
		AccountName:  accountInfoMap[txInfo.TreasuryAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.TreasuryAccountIndex][txInfo.AssetId].String(),
	})
	// owner account asset A
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.OwnerAccountIndex,
		AccountName:  accountInfoMap[txInfo.OwnerAccountIndex].AccountName,
		BalanceDelta: assetDeltaMap[txInfo.OwnerAccountIndex][txInfo.AssetId].String(),
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
