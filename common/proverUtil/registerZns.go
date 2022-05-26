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
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-crypto/zecrey-legend/circuit/bn254/std"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zeromicro/go-zero/core/logx"
)

func ConstructRegisterZnsCryptoTx(
	oTx *Tx,
	accountTree *Tree,
	accountAssetsTree *[]*Tree,
	liquidityTree *Tree,
	nftTree *Tree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeRegisterZns {
		logx.Errorf("[ConstructCreatePairCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructCreatePairCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructRegisterZnsCryptoTx] invalid params")
		return nil, errors.New("[ConstructRegisterZnsCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseRegisterZnsTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructRegisterZnsCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoRegisterZnsTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructRegisterZnsCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructRegisterZnsCryptoTx] unable to construct prover info: %s", err.Error())
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
		logx.Errorf("[ConstructRegisterZnsCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.RegisterZnsTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.Signature = std.EmptySignature()
	return cryptoTx, nil
}

func ToCryptoRegisterZnsTx(txInfo *commonTx.RegisterZnsTxInfo) (info *CryptoRegisterZnsTx, err error) {
	accountName := make([]byte, 32)
	copy(accountName[:], txInfo.AccountName)
	pk, err := util.ParsePubKey(txInfo.PubKey)
	if err != nil {
		logx.Errorf("[ToCryptoRegisterZnsTx] unable to parse pub key:%s", err.Error())
		return nil, err
	}
	info = &CryptoRegisterZnsTx{
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     accountName,
		AccountNameHash: common.FromHex(txInfo.AccountNameHash),
		PubKey:          pk,
	}
	return info, nil
}
