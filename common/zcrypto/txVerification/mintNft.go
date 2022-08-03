/*
 * Copyright © 2021 Zkbas Protocol
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
	"fmt"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/util"
)

func VerifyMintNftTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	txInfo *MintNftTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	realNftContentHash, err := util.FromHex(txInfo.NftContentHash)
	if err != nil || len(realNftContentHash) != 32 {
		logx.Errorf("invalid NftContentHash")
		return nil, errors.New("invalid NftContentHash")
	}
	// verify params
	if accountInfoMap[txInfo.CreatorAccountIndex] == nil ||
		accountInfoMap[txInfo.ToAccountIndex] == nil ||
		accountInfoMap[txInfo.CreatorAccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.CreatorAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Error("invalid params")
		return nil, errors.New("invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.CreatorAccountIndex].Nonce {
		logx.Errorf("invalid nonce, actual: %d, expected: %d",
			txInfo.Nonce, accountInfoMap[txInfo.CreatorAccountIndex].Nonce)
		return nil, fmt.Errorf("invalid nonce, actual: %d, expected: %d",
			txInfo.Nonce, accountInfoMap[txInfo.CreatorAccountIndex].Nonce)
	}
	// set tx info
	var (
		assetDeltaMap = make(map[int64]map[int64]*big.Int)
		newNftInfo    *NftInfo
	)
	// init delta map
	assetDeltaMap[txInfo.CreatorAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// from account asset Gas
	assetDeltaMap[txInfo.CreatorAccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	// to account nft info
	newNftInfo = &NftInfo{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		NftL1TokenId:        commonConstant.NilL1TokenId,
		NftL1Address:        commonConstant.NilL1Address,
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.NftCollectionId,
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
	if accountInfoMap[txInfo.CreatorAccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(
		new(big.Int).Abs(assetDeltaMap[txInfo.CreatorAccountIndex][txInfo.GasFeeAssetId])) < 0 {
		logx.Errorf("not enough balance of gas")
		return nil, errors.New("not enough balance of gas")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeMintNftMsgHash(txInfo, hFunc)
	if err != nil {
		logx.Errorf("unable to compute tx hash: %s", err.Error())
		return nil, errors.New("internal error")
	}
	// verify signature
	if err := VerifySignature(txInfo.Sig, msgHash, accountInfoMap[txInfo.CreatorAccountIndex].PublicKey); err != nil {
		return nil, err
	}
	// compute tx details
	// from account asset gas
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		AccountName:  accountInfoMap[txInfo.CreatorAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// to account empty delta
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  accountInfoMap[txInfo.ToAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ZeroBigInt, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// nft info
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    NftAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  accountInfoMap[txInfo.ToAccountIndex].AccountName,
		BalanceDelta: newNftInfo.String(),
		Order:        order,
		AccountOrder: commonConstant.NilAccountOrder,
	})
	// gas account asset gas
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
