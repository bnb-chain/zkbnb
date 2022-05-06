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
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/asset"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"math/big"
)

/*
	VerifyMintNftTx:
	accounts order is:
	- FromAccount
		- Assets
			- AssetGas
	- GasAccount
		- Assets
			- AssetGas
*/
func VerifyMintNftTxInfo(
	accountInfoMap map[int64]*account.Account,
	assetInfoMap map[int64]map[int64]*asset.AccountAsset,
	txInfo *MintNftTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.CreatorAccountIndex] == nil ||
		accountInfoMap[txInfo.ToAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		assetInfoMap[txInfo.CreatorAccountIndex] == nil ||
		assetInfoMap[txInfo.CreatorAccountIndex][txInfo.GasFeeAssetId] == nil ||
		assetInfoMap[txInfo.GasAccountIndex] == nil ||
		assetInfoMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] == nil ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Errorf("[VerifyMintNftTxInfo] invalid params")
		return nil, errors.New("[VerifyMintNftTxInfo] invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.CreatorAccountIndex].Nonce {
		log.Println("[VerifyMintNftTxInfo] invalid nonce")
		return nil, errors.New("[VerifyMintNftTxInfo] invalid nonce")
	}
	// set tx info
	var (
		assetDeltaMap          = make(map[int64]map[int64]*big.Int)
		oldNftInfo, newNftInfo *NftInfo
	)
	// init delta map
	assetDeltaMap[txInfo.CreatorAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// from account asset Gas
	assetDeltaMap[txInfo.CreatorAccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	// to account nft info
	oldNftInfo = util.EmptyNftInfo(txInfo.NftIndex)
	newNftInfo = &NftInfo{
		NftIndex:       txInfo.NftIndex,
		AssetId:        commonConstant.NilAssetId,
		AssetAmount:    commonConstant.NilAssetAmountStr,
		NftContentHash: txInfo.NftContentHash,
		NftL1TokenId:   commonConstant.NilL1TokenId,
		NftL1Address:   commonConstant.NilL1Address,
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
	assetGasBalance, isValid := new(big.Int).SetString(assetInfoMap[txInfo.CreatorAccountIndex][txInfo.GasFeeAssetId].Balance, 10)
	if !isValid {
		logx.Errorf("[VerifyMintNftTxInfo] unable to parse balance")
		return nil, errors.New("[VerifyMintNftTxInfo] unable to parse balance")
	}
	if assetGasBalance.Cmp(assetDeltaMap[txInfo.CreatorAccountIndex][txInfo.GasFeeAssetId]) < 0 {
		logx.Errorf("[VerifyMintNftTxInfo] you don't have enough balance of asset Gas")
		return nil, errors.New("[VerifyMintNftTxInfo] you don't have enough balance of asset Gas")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash := legendTxTypes.ComputeMintNftMsgHash(txInfo, hFunc)
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.CreatorAccountIndex].PublicKey)
	if err != nil {
		return nil, err
	}
	isValid, err = pk.Verify(txInfo.Sig, msgHash, hFunc)
	if err != nil {
		log.Println("[VerifyMintNftTxInfo] unable to verify signature:", err)
		return nil, err
	}
	if !isValid {
		log.Println("[VerifyMintNftTxInfo] invalid signature")
		return nil, errors.New("[VerifyMintNftTxInfo] invalid signature")
	}
	// compute tx details
	// from account asset gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		AccountName:  accountInfoMap[txInfo.CreatorAccountIndex].AccountName,
		Balance:      assetInfoMap[txInfo.CreatorAccountIndex][txInfo.GasFeeAssetId].Balance,
		BalanceDelta: assetDeltaMap[txInfo.CreatorAccountIndex][txInfo.GasFeeAssetId].String(),
	})
	// to account nft info
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    NftAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  accountInfoMap[txInfo.ToAccountIndex].AccountName,
		Balance:      oldNftInfo.String(),
		BalanceDelta: newNftInfo.String(),
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
