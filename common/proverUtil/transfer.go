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

func ConstructTransferCryptoTx(
	oTx *Tx,
	accountTree *Tree,
	accountAssetsTree *[]*Tree,
	liquidityTree *Tree,
	nftTree *Tree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeTransfer {
		logx.Errorf("[ConstructTransferCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructTransferCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructTransferCryptoTx] invalid params")
		return nil, errors.New("[ConstructTransferCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseTransferTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructTransferCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoTransferTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructTransferCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccountMap, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructTransferCryptoTx] unable to construct prover info: %s", err.Error())
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
		proverAccountMap,
		proverLiquidityInfo,
		proverNftInfo,
	)
	if err != nil {
		logx.Errorf("[ConstructTransferCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.TransferTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.ExpiredAt = oTx.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("[ConstructTransferCryptoTx] invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoTransferTx(txInfo *commonTx.TransferTxInfo) (info *CryptoTransferTx, err error) {
	packedAmount, err := util.ToPackedAmount(txInfo.AssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoTransferTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedFee, err := util.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoTransferTx] unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	info = &CryptoTransferTx{
		FromAccountIndex:  txInfo.FromAccountIndex,
		ToAccountIndex:    txInfo.ToAccountIndex,
		AssetId:           txInfo.AssetId,
		AssetAmount:       packedAmount,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: int64(packedFee),
		CallDataHash:      txInfo.CallDataHash,
	}
	return info, nil
}
