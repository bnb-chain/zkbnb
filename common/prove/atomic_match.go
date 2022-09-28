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

	cryptoTypes "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

func (w *WitnessHelper) constructAtomicMatchTxWitness(cryptoTx *TxWitness, oTx *tx.Tx) (*TxWitness, error) {
	txInfo, err := types.ParseAtomicMatchTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoAtomicMatchTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.AtomicMatchTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		return nil, err
	}
	return cryptoTx, nil
}

func toCryptoAtomicMatchTx(txInfo *txtypes.AtomicMatchTxInfo) (info *cryptoTypes.AtomicMatchTx, err error) {
	packedFee, err := common.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		return nil, err
	}
	packedAmount, err := common.ToPackedAmount(txInfo.BuyOffer.AssetAmount)
	if err != nil {
		return nil, err
	}
	packedCreatorAmount, err := common.ToPackedAmount(txInfo.CreatorAmount)
	if err != nil {
		return nil, err
	}
	packedTreasuryAmount, err := common.ToPackedAmount(txInfo.TreasuryAmount)
	if err != nil {
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
	info = &cryptoTypes.AtomicMatchTx{
		AccountIndex: txInfo.AccountIndex,
		BuyOffer: &cryptoTypes.OfferTx{
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
		SellOffer: &cryptoTypes.OfferTx{
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
