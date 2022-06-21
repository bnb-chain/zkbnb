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
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/zeromicro/go-zero/core/logx"
)

func ConstructRemoveLiquidityCryptoTx(
	oTx *Tx,
	accountTree *Tree,
	accountAssetsTree *[]*Tree,
	liquidityTree *Tree,
	nftTree *Tree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeRemoveLiquidity {
		logx.Errorf("[ConstructRemoveLiquidityCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructRemoveLiquidityCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructRemoveLiquidityCryptoTx] invalid params")
		return nil, errors.New("[ConstructRemoveLiquidityCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseRemoveLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructRemoveLiquidityCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoRemoveLiquidityTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructRemoveLiquidityCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructRemoveLiquidityCryptoTx] unable to construct prover info: %s", err.Error())
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
		logx.Errorf("[ConstructRemoveLiquidityCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.RemoveLiquidityTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("[ConstructRemoveLiquidityCryptoTx] invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoRemoveLiquidityTx(txInfo *commonTx.RemoveLiquidityTxInfo) (info *CryptoRemoveLiquidityTx, err error) {
	packedAMinAmount, err := util.ToPackedAmount(txInfo.AssetAMinAmount)
	if err != nil {
		logx.Errorf("[ToCryptoRemoveLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedBMinAmount, err := util.ToPackedAmount(txInfo.AssetBMinAmount)
	if err != nil {
		logx.Errorf("[ToCryptoRemoveLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedAAmount, err := util.ToPackedAmount(txInfo.AssetAAmountDelta)
	if err != nil {
		logx.Errorf("[ToCryptoRemoveLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedBAmount, err := util.ToPackedAmount(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ToCryptoRemoveLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedLpAmount, err := util.ToPackedAmount(txInfo.LpAmount)
	if err != nil {
		logx.Errorf("[ToCryptoRemoveLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedKLast, err := util.ToPackedAmount(txInfo.KLast)
	if err != nil {
		logx.Errorf("[ToCryptoAddLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedTreasuryAmount, err := util.ToPackedAmount(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("[ToCryptoAddLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedFee, err := util.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoRemoveLiquidityTx] unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	info = &CryptoRemoveLiquidityTx{
		FromAccountIndex:  txInfo.FromAccountIndex,
		PairIndex:         txInfo.PairIndex,
		AssetAId:          txInfo.AssetAId,
		AssetAMinAmount:   packedAMinAmount,
		AssetBId:          txInfo.AssetBId,
		AssetBMinAmount:   packedBMinAmount,
		LpAmount:          packedLpAmount,
		KLast:             packedKLast,
		TreasuryAmount:    packedTreasuryAmount,
		AssetAAmountDelta: packedAAmount,
		AssetBAmountDelta: packedBAmount,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
	}
	return info, nil
}
