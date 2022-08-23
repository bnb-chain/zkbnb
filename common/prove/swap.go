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

package prove

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/types"
)

func (w *WitnessHelper) constructSwapCryptoTx(cryptoTx *CryptoTx, oTx *Tx) (*CryptoTx, error) {
	txInfo, err := types.ParseSwapTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructSwapCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoSwapTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructSwapCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	cryptoTx.SwapTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("[ConstructSwapCryptoTx] invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoSwapTx(txInfo *types.SwapTxInfo) (info *CryptoSwapTx, err error) {
	packedAAmount, err := common.ToPackedAmount(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedBMinAmount, err := common.ToPackedAmount(txInfo.AssetBMinAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedBAmount, err := common.ToPackedAmount(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedFee, err := common.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	info = &CryptoSwapTx{
		FromAccountIndex:  txInfo.FromAccountIndex,
		PairIndex:         txInfo.PairIndex,
		AssetAId:          txInfo.AssetAId,
		AssetAAmount:      packedAAmount,
		AssetBId:          txInfo.AssetBId,
		AssetBMinAmount:   packedBMinAmount,
		AssetBAmountDelta: packedBAmount,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
	}
	return info, nil
}
