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

package proverUtil

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
)

func ConstructWithdrawNftCryptoTx(
	oTx *Tx,
	accountTree *Tree,
	accountAssetsTree *[]*Tree,
	liquidityTree *Tree,
	nftTree *Tree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeWithdrawNft {
		logx.Errorf("[ConstructWithdrawNftCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructWithdrawNftCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructWithdrawNftCryptoTx] invalid params")
		return nil, errors.New("[ConstructWithdrawNftCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseWithdrawNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructWithdrawNftCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoWithdrawNftTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructWithdrawNftCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructWithdrawNftCryptoTx] unable to construct prover info: %s", err.Error())
		return nil, err
	}
	cryptoTx, err = ConstructWitnessInfo(
		oTx,
		accountModel,
		accountTree,
		accountAssetsTree,
		liquidityTree,
		nftTree,
		accountKeys,
		proverAccounts,
		proverLiquidityInfo,
		proverNftInfo,
	)
	if err != nil {
		logx.Errorf("[ConstructWithdrawNftCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.WithdrawNftTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("[ConstructWithdrawNftCryptoTx] invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoWithdrawNftTx(txInfo *commonTx.WithdrawNftTxInfo) (info *CryptoWithdrawNftTx, err error) {
	packedFee, err := util.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	info = &CryptoWithdrawNftTx{
		AccountIndex:           txInfo.AccountIndex,
		CreatorAccountIndex:    txInfo.CreatorAccountIndex,
		CreatorAccountNameHash: txInfo.CreatorAccountNameHash,
		CreatorTreasuryRate:    txInfo.CreatorTreasuryRate,
		NftIndex:               txInfo.NftIndex,
		NftContentHash:         txInfo.NftContentHash,
		NftL1Address:           txInfo.NftL1Address,
		NftL1TokenId:           txInfo.NftL1TokenId,
		ToAddress:              txInfo.ToAddress,
		GasAccountIndex:        txInfo.GasAccountIndex,
		GasFeeAssetId:          txInfo.GasFeeAssetId,
		GasFeeAssetAmount:      packedFee,
	}
	return info, nil
}
