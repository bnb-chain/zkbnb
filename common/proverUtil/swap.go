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
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/pkg/treedb"
)

func ConstructSwapCryptoTx(
	oTx *Tx,
	treeDBDriver treedb.Driver,
	treeDB database.TreeDB,
	finalityBlockNr uint64,
	accountTree bsmt.SparseMerkleTree,
	accountAssetsTree *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeSwap {
		logx.Errorf("[ConstructSwapCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructSwapCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructSwapCryptoTx] invalid params")
		return nil, errors.New("[ConstructSwapCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseSwapTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructSwapCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoSwapTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructSwapCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructSwapCryptoTx] unable to construct prover info: %s", err.Error())
		return nil, err
	}
	cryptoTx, err = ConstructWitnessInfo(
		oTx,
		accountModel,
		treeDBDriver,
		treeDB,
		finalityBlockNr,
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
		logx.Errorf("[ConstructSwapCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.SwapTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("[ConstructSwapCryptoTx] invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoSwapTx(txInfo *commonTx.SwapTxInfo) (info *CryptoSwapTx, err error) {
	packedAAmount, err := util.ToPackedAmount(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedBMinAmount, err := util.ToPackedAmount(txInfo.AssetBMinAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedBAmount, err := util.ToPackedAmount(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedFee, err := util.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	info = &CryptoSwapTx{
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
