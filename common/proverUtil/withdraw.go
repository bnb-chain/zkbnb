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
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/util"
)

func (w *WitnessHelper) constructWithdrawCryptoTx(cryptoTx *CryptoTx, oTx *Tx) (*CryptoTx, error) {
	txInfo, err := commonTx.ParseWithdrawTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructWithdrawCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoWithdrawTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructWithdrawCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	cryptoTx.WithdrawTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = oTx.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("[ConstructWithdrawCryptoTx] invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoWithdrawTx(txInfo *commonTx.WithdrawTxInfo) (info *CryptoWithdrawTx, err error) {
	packedFee, err := util.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoWithdrawTx] unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	addrBytes := legendTxTypes.PaddingAddressToBytes32(txInfo.ToAddress)
	info = &CryptoWithdrawTx{
		FromAccountIndex:  txInfo.FromAccountIndex,
		AssetId:           txInfo.AssetId,
		AssetAmount:       txInfo.AssetAmount,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
		ToAddress:         new(big.Int).SetBytes(addrBytes),
	}
	return info, nil
}
