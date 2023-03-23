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
	cryptoTypes "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

func (w *WitnessHelper) constructFullExitNftTxWitness(cryptoTx *TxWitness, oTx *tx.Tx) (*TxWitness, error) {
	txInfo, err := types.ParseFullExitNftTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoFullExitNftTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.FullExitNftTxInfo = cryptoTxInfo
	cryptoTx.Signature = cryptoTypes.EmptySignature()
	return cryptoTx, nil
}

func toCryptoFullExitNftTx(txInfo *txtypes.FullExitNftTxInfo) (info *cryptoTypes.FullExitNftTx, err error) {
	info = &cryptoTypes.FullExitNftTx{
		AccountIndex:        txInfo.AccountIndex,
		L1Address:           common.AddressStrToBytes(txInfo.L1Address),
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		CreatorL1Address:    common.AddressStrToBytes(txInfo.CreatorL1Address),
		RoyaltyRate:         txInfo.RoyaltyRate,
		NftIndex:            txInfo.NftIndex,
		CollectionId:        txInfo.CollectionId,
		NftContentHash:      txInfo.NftContentHash,
		NftContentType:      txInfo.NftContentType,
	}
	return info, nil
}
