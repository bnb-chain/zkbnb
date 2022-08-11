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
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
)

func (w *WitnessHelper) constructDepositNftCryptoTx(cryptoTx *CryptoTx, oTx *Tx) (*CryptoTx, error) {
	txInfo, err := commonTx.ParseDepositNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructDepositNftCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoDepositNftTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructDepositNftCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	cryptoTx.DepositNftTxInfo = cryptoTxInfo
	cryptoTx.Signature = std.EmptySignature()
	return cryptoTx, nil
}

func ToCryptoDepositNftTx(txInfo *commonTx.DepositNftTxInfo) (info *CryptoDepositNftTx, err error) {
	info = &CryptoDepositNftTx{
		AccountIndex:        txInfo.AccountIndex,
		NftIndex:            txInfo.NftIndex,
		NftL1Address:        txInfo.NftL1Address,
		AccountNameHash:     txInfo.AccountNameHash,
		NftContentHash:      txInfo.NftContentHash,
		NftL1TokenId:        txInfo.NftL1TokenId,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.CollectionId,
	}
	return info, nil
}
