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
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"

	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/types"
)

func constructWithdrawTxWitness(cryptoTx *TxWitness, oTx *Tx) (*TxWitness, error) {
	txInfo, err := types.ParseWithdrawTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoWithdrawTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.WithdrawTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = oTx.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		return nil, err
	}
	return cryptoTx, nil
}

func toCryptoWithdrawTx(txInfo *types.WithdrawTxInfo) (info *CryptoWithdrawTx, err error) {
	packedFee, err := common.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
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
