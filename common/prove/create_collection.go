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

	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/types"
)

func constructCreateCollectionTxWitness(cryptoTx *TxWitness, oTx *Tx) (*TxWitness, error) {
	txInfo, err := types.ParseCreateCollectionTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoCreateCollectionTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.CreateCollectionTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		return nil, err
	}
	return cryptoTx, nil
}

func toCryptoCreateCollectionTx(txInfo *types.CreateCollectionTxInfo) (info *CryptoCreateCollectionTx, err error) {
	packedFee, err := common.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		return nil, err
	}
	info = &CryptoCreateCollectionTx{
		AccountIndex:      txInfo.AccountIndex,
		CollectionId:      txInfo.CollectionId,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
		ExpiredAt:         txInfo.ExpiredAt,
		Nonce:             txInfo.Nonce,
	}
	return info, nil
}
