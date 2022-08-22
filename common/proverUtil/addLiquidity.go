/*
 * Copyright © 2021 ZkBAS Protocol
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
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/util"
)

func (w *WitnessHelper) constructAddLiquidityCryptoTx(cryptoTx *CryptoTx, oTx *Tx) (*CryptoTx, error) {
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructAddLiquidityCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoAddLiquidityTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructAddLiquidityCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	cryptoTx.AddLiquidityTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("[ConstructAddLiquidityCryptoTx] invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoAddLiquidityTx(txInfo *commonTx.AddLiquidityTxInfo) (info *CryptoAddLiquidityTx, err error) {
	packedAAmount, err := util.ToPackedAmount(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ToCryptoAddLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedBAmount, err := util.ToPackedAmount(txInfo.AssetBAmount)
	if err != nil {
		logx.Errorf("[ToCryptoAddLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedLpAmount, err := util.ToPackedAmount(txInfo.LpAmount)
	if err != nil {
		logx.Errorf("[ToCryptoAddLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedTreasuryAmount, err := util.ToPackedAmount(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("[ToCryptoAddLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedKLast, err := util.ToPackedAmount(txInfo.KLast)
	if err != nil {
		logx.Errorf("[ToCryptoAddLiquidityTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedFee, err := util.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoAddLiquidityTx] unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	info = &CryptoAddLiquidityTx{
		FromAccountIndex:  txInfo.FromAccountIndex,
		PairIndex:         txInfo.PairIndex,
		AssetAId:          txInfo.AssetAId,
		AssetAAmount:      packedAAmount,
		AssetBId:          txInfo.AssetBId,
		AssetBAmount:      packedBAmount,
		LpAmount:          packedLpAmount,
		KLast:             packedKLast,
		TreasuryAmount:    packedTreasuryAmount,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
	}
	return info, nil
}
