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

func constructWithdrawNftTxWitness(cryptoTx *TxWitness, oTx *Tx) (*TxWitness, error) {
	txInfo, err := types.ParseWithdrawNftTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoWithdrawNftTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.WithdrawNftTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		return nil, err
	}
	return cryptoTx, nil
}

func toCryptoWithdrawNftTx(txInfo *types.WithdrawNftTxInfo) (info *CryptoWithdrawNftTx, err error) {
	packedFee, err := common.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		return nil, err
	}
	info = &CryptoWithdrawNftTx{
		AccountIndex:           txInfo.AccountIndex,
		CreatorAccountIndex:    txInfo.CreatorAccountIndex,
		CreatorAccountNameHash: txInfo.CreatorAccountNameHash,
		CreatorTreasuryRate:    txInfo.CreatorTreasuryRate,
		NftIndex:               txInfo.NftIndex,
		NftContentHash:         txInfo.NftContentHash,
		NftL1Address:           txInfo.NftL1Address,
		NftL1TokenId:           txInfo.NftL1TokenId,
		ToAddress:              txInfo.ToAddress,
		GasAccountIndex:        txInfo.GasAccountIndex,
		GasFeeAssetId:          txInfo.GasFeeAssetId,
		GasFeeAssetAmount:      packedFee,
		CollectionId:           txInfo.CollectionId,
	}
	return info, nil
}
