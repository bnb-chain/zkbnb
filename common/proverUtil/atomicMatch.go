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
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/pkg/treedb"
)

func ConstructAtomicMatchCryptoTx(
	oTx *Tx,
	treeCtx *treedb.Context,
	finalityBlockNr uint64,
	accountTree bsmt.SparseMerkleTree,
	accountAssetsTree *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeAtomicMatch {
		logx.Errorf("[ConstructAtomicMatchCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructAtomicMatchCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructAtomicMatchCryptoTx] invalid params")
		return nil, errors.New("[ConstructAtomicMatchCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseAtomicMatchTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructAtomicMatchCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoAtomicMatchTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructAtomicMatchCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructAtomicMatchCryptoTx] unable to construct prover info: %s", err.Error())
		return nil, err
	}
	cryptoTx, err = ConstructWitnessInfo(
		oTx,
		accountModel,
		treeCtx,
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
		logx.Errorf("[ConstructAtomicMatchCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.AtomicMatchTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("[ConstructAtomicMatchCryptoTx] invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoAtomicMatchTx(txInfo *commonTx.AtomicMatchTxInfo) (info *CryptoAtomicMatchTx, err error) {
	packedFee, err := util.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	packedAmount, err := util.ToPackedAmount(txInfo.BuyOffer.AssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedCreatorAmount, err := util.ToPackedAmount(txInfo.CreatorAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	packedTreasuryAmount, err := util.ToPackedAmount(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed amount: %s", err.Error())
		return nil, err
	}
	buySig := new(eddsa.Signature)
	_, err = buySig.SetBytes(txInfo.BuyOffer.Sig)
	if err != nil {
		return nil, err
	}
	sellSig := new(eddsa.Signature)
	_, err = sellSig.SetBytes(txInfo.SellOffer.Sig)
	if err != nil {
		return nil, err
	}
	info = &CryptoAtomicMatchTx{
		AccountIndex: txInfo.AccountIndex,
		BuyOffer: &CryptoOfferTx{
			Type:         txInfo.BuyOffer.Type,
			OfferId:      txInfo.BuyOffer.OfferId,
			AccountIndex: txInfo.BuyOffer.AccountIndex,
			NftIndex:     txInfo.BuyOffer.NftIndex,
			AssetId:      txInfo.BuyOffer.AssetId,
			AssetAmount:  packedAmount,
			ListedAt:     txInfo.BuyOffer.ListedAt,
			ExpiredAt:    txInfo.BuyOffer.ExpiredAt,
			TreasuryRate: txInfo.BuyOffer.TreasuryRate,
			Sig:          buySig,
		},
		SellOffer: &CryptoOfferTx{
			Type:         txInfo.SellOffer.Type,
			OfferId:      txInfo.SellOffer.OfferId,
			AccountIndex: txInfo.SellOffer.AccountIndex,
			NftIndex:     txInfo.SellOffer.NftIndex,
			AssetId:      txInfo.SellOffer.AssetId,
			AssetAmount:  packedAmount,
			ListedAt:     txInfo.SellOffer.ListedAt,
			ExpiredAt:    txInfo.SellOffer.ExpiredAt,
			TreasuryRate: txInfo.SellOffer.TreasuryRate,
			Sig:          sellSig,
		},
		CreatorAmount:     packedCreatorAmount,
		TreasuryAmount:    packedTreasuryAmount,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
	}
	return info, nil
}
