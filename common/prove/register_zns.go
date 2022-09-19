/*
 * Copyright © 2021 ZkBNB Protocol
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

	cryptoTypes "github.com/bnb-chain/zkbnb-crypto/circuit/bn254/types"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

func (w *WitnessHelper) constructRegisterZnsTxWitness(cryptoTx *TxWitness, oTx *tx.Tx) (*TxWitness, error) {
	txInfo, err := types.ParseRegisterZnsTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoRegisterZnsTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.Signature = cryptoTypes.EmptySignature()
	cryptoTx.RegisterZnsTxInfo = cryptoTxInfo
	return cryptoTx, nil
}

func toCryptoRegisterZnsTx(txInfo *txtypes.RegisterZnsTxInfo) (info *cryptoTypes.RegisterZnsTx, err error) {
	accountName := make([]byte, 32)
	realName := strings.Split(txInfo.AccountName, types.AccountNameSuffix)[0]
	copy(accountName[:], realName)
	pk, err := common.ParsePubKey(txInfo.PubKey)
	if err != nil {
		return nil, err
	}
	info = &cryptoTypes.RegisterZnsTx{
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     accountName,
		AccountNameHash: txInfo.AccountNameHash,
		PubKey:          pk,
	}
	return info, nil
}
