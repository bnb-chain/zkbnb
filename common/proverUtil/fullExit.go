/*
 * Copyright © 2021 Zkbas Protocol
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

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/bas-smt/database"
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/pkg/treedb"
)

func ConstructFullExitCryptoTx(
	oTx *Tx,
	treeDBDriver treedb.Driver,
	treeDB database.TreeDB,
	accountTree bsmt.SparseMerkleTree,
	accountAssetsTree *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree,
	accountModel AccountModel,
	finalityBlockNr uint64,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeFullExit {
		logx.Errorf("[ConstructFullExitCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructFullExitCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructFullExitCryptoTx] invalid params")
		return nil, errors.New("[ConstructFullExitCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseFullExitTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructFullExitCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoFullExitTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructFullExitCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructFullExitCryptoTx] unable to construct prover info: %s", err.Error())
		return nil, err
	}
	cryptoTx, err = ConstructWitnessInfo(
		oTx,
		accountModel,
		treeDBDriver,
		treeDB,
		accountTree,
		accountAssetsTree,
		liquidityTree,
		nftTree,
		accountKeys,
		proverAccounts,
		proverLiquidityInfo,
		proverNftInfo,
		finalityBlockNr,
	)
	if err != nil {
		logx.Errorf("[ConstructFullExitCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.FullExitTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.Signature = std.EmptySignature()
	return cryptoTx, nil
}

func ToCryptoFullExitTx(txInfo *commonTx.FullExitTxInfo) (info *CryptoFullExitTx, err error) {
	info = &CryptoFullExitTx{
		AccountIndex:    txInfo.AccountIndex,
		AssetId:         txInfo.AssetId,
		AssetAmount:     txInfo.AssetAmount,
		AccountNameHash: txInfo.AccountNameHash,
	}
	return info, nil
}
