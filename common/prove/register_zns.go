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
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/types"
)

func (w *WitnessHelper) constructRegisterZnsCryptoTx(cryptoTx *CryptoTx, oTx *Tx) (*CryptoTx, error) {
	txInfo, err := types.ParseRegisterZnsTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := toCryptoRegisterZnsTx(txInfo)
	if err != nil {
		logx.Errorf("unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	cryptoTx.Signature = std.EmptySignature()
	cryptoTx.RegisterZnsTxInfo = cryptoTxInfo
	return cryptoTx, nil
}

func toCryptoRegisterZnsTx(txInfo *legendTxTypes.RegisterZnsTxInfo) (info *CryptoRegisterZnsTx, err error) {
	accountName := make([]byte, 32)
	realName := strings.Split(txInfo.AccountName, types.AccountNameSuffix)[0]
	copy(accountName[:], realName)
	pk, err := common.ParsePubKey(txInfo.PubKey)
	if err != nil {
		logx.Errorf("unable to parse pub key:%s", err.Error())
		return nil, err
	}
	info = &CryptoRegisterZnsTx{
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     accountName,
		AccountNameHash: txInfo.AccountNameHash,
		PubKey:          pk,
	}
	return info, nil
}
