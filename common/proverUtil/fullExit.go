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

package proverUtil

import (
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
)

func (w *WitnessHelper) constructFullExitCryptoTx(cryptoTx *CryptoTx, oTx *Tx) (*CryptoTx, error) {
	txInfo, err := commonTx.ParseFullExitTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructFullExitCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoFullExitTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructFullExitCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	cryptoTx.FullExitTxInfo = cryptoTxInfo
	cryptoTx.Signature = std.EmptySignature()
	return cryptoTx, nil
}

func ToCryptoFullExitTx(txInfo *legendTxTypes.FullExitTxInfo) (info *CryptoFullExitTx, err error) {
	info = &CryptoFullExitTx{
		AccountIndex:    txInfo.AccountIndex,
		AssetId:         txInfo.AssetId,
		AssetAmount:     txInfo.AssetAmount,
		AccountNameHash: txInfo.AccountNameHash,
	}
	return info, nil
}
