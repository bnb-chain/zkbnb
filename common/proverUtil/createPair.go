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

func ConstructCreatePairCryptoTx(
	oTx *Tx,
	accountTree *Tree,
	accountAssetsTree *[]*Tree,
	liquidityTree *Tree,
	nftTree *Tree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeCreatePair {
		logx.Errorf("[ConstructCreatePairCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructCreatePairCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructCreatePairCryptoTx] invalid params")
		return nil, errors.New("[ConstructCreatePairCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseCreatePairTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructCreatePairCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoCreatePairTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructCreatePairCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructCreatePairCryptoTx] unable to construct prover info: %s", err.Error())
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
		logx.Errorf("[ConstructCreatePairCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.CreatePairTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.Signature = std.EmptySignature()
	return cryptoTx, nil
}

func ToCryptoCreatePairTx(txInfo *commonTx.CreatePairTxInfo) (info *CryptoCreatePairTx, err error) {
	info = &CryptoCreatePairTx{
		PairIndex:            txInfo.PairIndex,
		AssetAId:             txInfo.AssetAId,
		AssetBId:             txInfo.AssetBId,
		FeeRate:              txInfo.FeeRate,
		TreasuryAccountIndex: txInfo.TreasuryAccountIndex,
		TreasuryRate:         txInfo.TreasuryRate,
	}
	return info, nil
}
