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

	"github.com/bnb-chain/zkbnb-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

func (w *WitnessHelper) constructSwapTxWitness(cryptoTx *TxWitness, oTx *tx.Tx) (*TxWitness, error) {
	txInfo, err := types.ParseSwapTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoSwapTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.SwapTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		return nil, err
	}
	return cryptoTx, nil
}

func toCryptoSwapTx(txInfo *legendTxTypes.SwapTxInfo) (info *std.SwapTx, err error) {
	packedAAmount, err := common.ToPackedAmount(txInfo.AssetAAmount)
	if err != nil {
		return nil, err
	}
	packedBMinAmount, err := common.ToPackedAmount(txInfo.AssetBMinAmount)
	if err != nil {
		return nil, err
	}
	packedBAmount, err := common.ToPackedAmount(txInfo.AssetBAmountDelta)
	if err != nil {
		return nil, err
	}
	packedFee, err := common.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		return nil, err
	}
	info = &std.SwapTx{
		FromAccountIndex:  txInfo.FromAccountIndex,
		PairIndex:         txInfo.PairIndex,
		AssetAId:          txInfo.AssetAId,
		AssetAAmount:      packedAAmount,
		AssetBId:          txInfo.AssetBId,
		AssetBMinAmount:   packedBMinAmount,
		AssetBAmountDelta: packedBAmount,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
	}
	return info, nil
}
