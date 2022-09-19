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
	cryptoTypes "github.com/bnb-chain/zkbnb-crypto/circuit/bn254/types"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

func (w *WitnessHelper) constructDepositNftTxWitness(cryptoTx *TxWitness, oTx *tx.Tx) (*TxWitness, error) {
	txInfo, err := types.ParseDepositNftTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoDepositNftTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.DepositNftTxInfo = cryptoTxInfo
	cryptoTx.Signature = cryptoTypes.EmptySignature()
	return cryptoTx, nil
}

func toCryptoDepositNftTx(txInfo *txtypes.DepositNftTxInfo) (info *cryptoTypes.DepositNftTx, err error) {
	info = &cryptoTypes.DepositNftTx{
		IsNewNft:            txInfo.IsNewNft,
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
