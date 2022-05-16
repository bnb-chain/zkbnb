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
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"math/big"
)

/*
	VerifyTransferNftTx:
	accounts order is:
	- FromAccount
		- Assets
			- AssetGas
		- Nft
			- nft index
	- ToAccount
		- Nft
			- nft index
	- GasAccount
		- Assets
			- AssetGas
*/
func VerifyTransferNftTxInfo(
	accountInfoMap map[int64]*commonAsset.FormatAccountInfo,
	nftInfoMap map[int64]*nft.L2Nft,
	txInfo *TransferNftTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[txInfo.ToAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance == "" ||
		nftInfoMap[txInfo.NftIndex] == nil ||
		nftInfoMap[txInfo.NftIndex].OwnerAccountIndex != txInfo.FromAccountIndex ||
		nftInfoMap[txInfo.NftIndex].NftIndex != txInfo.NftIndex ||
		nftInfoMap[txInfo.NftIndex].NftContentHash != txInfo.NftContentHash ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Errorf("[VerifyTransferNftTxInfo] invalid params")
		return nil, errors.New("[VerifyTransferNftTxInfo] invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.FromAccountIndex].Nonce {
		log.Println("[VerifyTransferNftTxInfo] invalid nonce")
		return nil, errors.New("[VerifyTransferNftTxInfo] invalid nonce")
	}
	// set tx info
	var (
		assetDeltaMap = make(map[int64]map[int64]*big.Int)
		newNftInfo    *NftInfo
	)
	// init delta map
	assetDeltaMap[txInfo.FromAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// from account asset Gas
	assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	// to account nft info
	newNftInfo = &NftInfo{
		NftIndex:            nftInfoMap[txInfo.NftIndex].NftIndex,
		CreatorAccountIndex: nftInfoMap[txInfo.NftIndex].CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
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
	assetGasBalance, isValid := new(big.Int).SetString(accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance, 10)
	if !isValid {
		logx.Errorf("[VerifyMintNftTxInfo] unable to parse balance")
		return nil, errors.New("[VerifyMintNftTxInfo] unable to parse balance")
	}
	if assetGasBalance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		logx.Errorf("[VerifyMintNftTxInfo] you don't have enough balance of asset Gas")
		return nil, errors.New("[VerifyMintNftTxInfo] you don't have enough balance of asset Gas")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash := legendTxTypes.ComputeTransferNftMsgHash(txInfo, hFunc)
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.FromAccountIndex].PublicKey)
	if err != nil {
		return nil, err
	}
	isValid, err = pk.Verify(txInfo.Sig, msgHash, hFunc)
	if err != nil {
		log.Println("[VerifyTransferNftTxInfo] unable to verify signature:", err)
		return nil, err
	}
	if !isValid {
		log.Println("[VerifyTransferNftTxInfo] invalid signature")
		return nil, errors.New("[VerifyTransferNftTxInfo] invalid signature")
	}
	// compute tx details
	// from account asset gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: ffmath.Neg(txInfo.GasFeeAssetAmount).String(),
	})
	// to account nft delta
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    NftAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  accountInfoMap[txInfo.ToAccountIndex].AccountName,
		BalanceDelta: newNftInfo.String(),
	})
	// gas account asset gas
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  accountInfoMap[txInfo.GasAccountIndex].AccountName,
		BalanceDelta: txInfo.GasFeeAssetAmount.String(),
	})
	return txDetails, nil
}
