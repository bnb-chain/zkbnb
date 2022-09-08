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
	"github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/types"
)

func (w *WitnessHelper) constructRemoveLiquidityTxWitness(cryptoTx *TxWitness, oTx *Tx) (*TxWitness, error) {
	txInfo, err := types.ParseRemoveLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoRemoveLiquidityTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.RemoveLiquidityTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = txInfo.Sig
	if err != nil {
		return nil, err
	}
	return cryptoTx, nil
}

func toCryptoRemoveLiquidityTx(txInfo *types.RemoveLiquidityTxInfo) (info *CryptoRemoveLiquidityTx, err error) {
	packedAMinAmount, err := common.ToPackedAmount(txInfo.AssetAMinAmount)
	if err != nil {
		return nil, err
	}
	packedBMinAmount, err := common.ToPackedAmount(txInfo.AssetBMinAmount)
	if err != nil {
		return nil, err
	}
	packedAAmount, err := common.ToPackedAmount(txInfo.AssetAAmountDelta)
	if err != nil {
		return nil, err
	}
	packedBAmount, err := common.ToPackedAmount(txInfo.AssetBAmountDelta)
	if err != nil {
		return nil, err
	}
	packedLpAmount, err := common.ToPackedAmount(txInfo.LpAmount)
	if err != nil {
		return nil, err
	}
	packedKLast, err := common.ToPackedAmount(txInfo.KLast)
	if err != nil {
		return nil, err
	}
	packedTreasuryAmount, err := common.ToPackedAmount(txInfo.TreasuryAmount)
	if err != nil {
		return nil, err
	}
	packedFee, err := common.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		return nil, err
	}
	info = &CryptoRemoveLiquidityTx{
		FromAccountIndex:  txInfo.FromAccountIndex,
		PairIndex:         txInfo.PairIndex,
		AssetAId:          txInfo.AssetAId,
		AssetAMinAmount:   packedAMinAmount,
		AssetBId:          txInfo.AssetBId,
		AssetBMinAmount:   packedBMinAmount,
		LpAmount:          packedLpAmount,
		KLast:             packedKLast,
		TreasuryAmount:    packedTreasuryAmount,
		AssetAAmountDelta: packedAAmount,
		AssetBAmountDelta: packedBAmount,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
	}
	return info, nil
}
