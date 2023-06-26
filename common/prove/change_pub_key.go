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
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"

	cryptoTypes "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

func (w *WitnessHelper) constructChangePubKeyTxWitness(cryptoTx *TxWitness, oTx *tx.Tx) (*TxWitness, error) {
	txInfo, err := types.ParseChangePubKeyTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoChangePubKeyTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.ChangePubKeyTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = oTx.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		return nil, err
	}
	return cryptoTx, nil
}

func toCryptoChangePubKeyTx(txInfo *txtypes.ChangePubKeyInfo) (info *cryptoTypes.ChangePubKeyTx, err error) {
	pk := new(eddsa.PublicKey)
	pk.A.X.SetBytes(txInfo.PubKeyX)
	pk.A.Y.SetBytes(txInfo.PubKeyY)

	packedFee, err := common.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		return nil, err
	}
	info = &cryptoTypes.ChangePubKeyTx{
		AccountIndex:      txInfo.AccountIndex,
		L1Address:         common.AddressStrToBytes(txInfo.L1Address),
		PubKey:            pk,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
	}
	return info, nil
}
