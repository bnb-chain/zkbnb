/*
 * Copyright © 2021 Zecrey Protocol
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
	"errors"
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/zeromicro/go-zero/core/logx"
)

func ConstructUpdatePairRateCryptoTx(
	oTx *Tx,
	accountTree *Tree,
	accountAssetsTree *[]*Tree,
	liquidityTree *Tree,
	nftTree *Tree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeUpdatePairRate {
		logx.Errorf("[ConstructUpdatePairRateCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructUpdatePairRateCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructUpdatePairRateCryptoTx] invalid params")
		return nil, errors.New("[ConstructUpdatePairRateCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseUpdatePairRateTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructUpdatePairRateCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoUpdatePairRateTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructUpdatePairRateCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructUpdatePairRateCryptoTx] unable to construct prover info: %s", err.Error())
		return nil, err
	}
	cryptoTx, err = ConstructWitnessInfo(
		oTx,
		accountModel,
		accountTree,
		accountAssetsTree,
		liquidityTree,
		nftTree,
		accountKeys,
		proverAccounts,
		proverLiquidityInfo,
		proverNftInfo,
	)
	if err != nil {
		logx.Errorf("[ConstructUpdatePairRateCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.UpdatePairRateTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.Signature = std.EmptySignature()
	return cryptoTx, nil
}

func ToCryptoUpdatePairRateTx(txInfo *commonTx.UpdatePairRateTxInfo) (info *CryptoUpdatePairRateTx, err error) {
	info = &CryptoUpdatePairRateTx{
		PairIndex:            txInfo.PairIndex,
		FeeRate:              txInfo.FeeRate,
		TreasuryAccountIndex: txInfo.TreasuryAccountIndex,
		TreasuryRate:         txInfo.TreasuryRate,
	}
	return info, nil
}
