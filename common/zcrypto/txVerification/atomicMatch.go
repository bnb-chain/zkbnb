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

package txVerification

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/util"
)

func VerifyAtomicMatchTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	nftInfo *NftInfo,
	txInfo *AtomicMatchTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	now := time.Now().UnixMilli()
	if accountInfoMap[txInfo.AccountIndex] == nil ||
		accountInfoMap[txInfo.AccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.GasFeeAssetId] == nil ||
		accountInfoMap[txInfo.BuyOffer.AccountIndex] == nil ||
		accountInfoMap[txInfo.SellOffer.AccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		accountInfoMap[nftInfo.CreatorAccountIndex] == nil ||
		txInfo.BuyOffer.Type != commonAsset.BuyOfferType ||
		txInfo.SellOffer.Type != commonAsset.SellOfferType ||
		txInfo.BuyOffer.NftIndex != txInfo.SellOffer.NftIndex ||
		txInfo.BuyOffer.AssetId != txInfo.SellOffer.AssetId ||
		txInfo.BuyOffer.AssetAmount.String() != txInfo.SellOffer.AssetAmount.String() ||
		txInfo.BuyOffer.ExpiredAt < now ||
		txInfo.SellOffer.ExpiredAt < now ||
		txInfo.BuyOffer.NftIndex != nftInfo.NftIndex ||
		txInfo.SellOffer.AccountIndex != nftInfo.OwnerAccountIndex ||
		txInfo.BuyOffer.AccountIndex == txInfo.SellOffer.AccountIndex ||
		accountInfoMap[txInfo.BuyOffer.AccountIndex].AssetInfo[txInfo.BuyOffer.AssetId] == nil ||
		accountInfoMap[txInfo.BuyOffer.AccountIndex].AssetInfo[txInfo.BuyOffer.AssetId].Balance.Cmp(ZeroBigInt) <= 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Error("invalid params")
		return nil, errors.New("invalid params")
	}
	buyerOfferAssetId := txInfo.BuyOffer.OfferId / OfferPerAsset
	sellerOfferAssetId := txInfo.SellOffer.OfferId / OfferPerAsset
	if accountInfoMap[txInfo.BuyOffer.AccountIndex].AssetInfo[buyerOfferAssetId] == nil {
		accountInfoMap[txInfo.BuyOffer.AccountIndex].AssetInfo[buyerOfferAssetId] = &commonAsset.AccountAsset{
			AssetId:                  buyerOfferAssetId,
			Balance:                  ZeroBigInt,
			LpAmount:                 ZeroBigInt,
			OfferCanceledOrFinalized: ZeroBigInt,
		}
	}
	if accountInfoMap[txInfo.SellOffer.AccountIndex].AssetInfo[sellerOfferAssetId] == nil {
		accountInfoMap[txInfo.SellOffer.AccountIndex].AssetInfo[sellerOfferAssetId] = &commonAsset.AccountAsset{
			AssetId:                  sellerOfferAssetId,
			Balance:                  ZeroBigInt,
			LpAmount:                 ZeroBigInt,
			OfferCanceledOrFinalized: ZeroBigInt,
		}
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.AccountIndex].Nonce {
		logx.Errorf("invalid nonce, actual: %d, expected: %d",
			txInfo.Nonce, accountInfoMap[txInfo.AccountIndex].Nonce)
		return nil, fmt.Errorf("invalid nonce, actual: %d, expected: %d",
			txInfo.Nonce, accountInfoMap[txInfo.AccountIndex].Nonce)
	}
	// set tx info
	var (
		assetDeltaMap = make(map[int64]map[int64]*big.Int)
	)
	// init delta map
	assetDeltaMap[txInfo.AccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.BuyOffer.AccountIndex] == nil {
		assetDeltaMap[txInfo.BuyOffer.AccountIndex] = make(map[int64]*big.Int)
	}
	if assetDeltaMap[txInfo.SellOffer.AccountIndex] == nil {
		assetDeltaMap[txInfo.SellOffer.AccountIndex] = make(map[int64]*big.Int)
	}
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// from account asset Gas
	assetDeltaMap[txInfo.AccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	// buyer account asset A
	if assetDeltaMap[txInfo.BuyOffer.AccountIndex][txInfo.BuyOffer.AssetId] == nil {
		assetDeltaMap[txInfo.BuyOffer.AccountIndex][txInfo.BuyOffer.AssetId] = ffmath.Neg(txInfo.BuyOffer.AssetAmount)
	} else {
		assetDeltaMap[txInfo.BuyOffer.AccountIndex][txInfo.BuyOffer.AssetId] = ffmath.Sub(
			assetDeltaMap[txInfo.BuyOffer.AccountIndex][txInfo.BuyOffer.AssetId],
			txInfo.BuyOffer.AssetAmount,
		)
	}
	// seller account asset A
	if assetDeltaMap[txInfo.SellOffer.AccountIndex][txInfo.SellOffer.AssetId] == nil {
		assetDeltaMap[txInfo.SellOffer.AccountIndex][txInfo.SellOffer.AssetId] =
			txInfo.SellOffer.AssetAmount
	} else {
		assetDeltaMap[txInfo.SellOffer.AccountIndex][txInfo.SellOffer.AssetId] = ffmath.Add(
			assetDeltaMap[txInfo.SellOffer.AccountIndex][txInfo.SellOffer.AssetId],
			txInfo.SellOffer.AssetAmount,
		)
	}
	// gas account asset Gas
	if assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] == nil {
		assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] = txInfo.GasFeeAssetAmount
	} else {
		assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] = ffmath.Add(
			assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId],
			txInfo.GasFeeAssetAmount,
		)
	}
	// check balance
	if accountInfoMap[txInfo.BuyOffer.AccountIndex].AssetInfo[txInfo.BuyOffer.AssetId].Balance.Cmp(
		new(big.Int).Abs(assetDeltaMap[txInfo.BuyOffer.AccountIndex][txInfo.BuyOffer.AssetId])) < 0 {
		logx.Errorf("not enough balance of asset %d", txInfo.BuyOffer.AssetId)
		return nil, fmt.Errorf("not enough balance of asset %d", txInfo.BuyOffer.AssetId)
	}
	if accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(
		new(big.Int).Abs(assetDeltaMap[txInfo.AccountIndex][txInfo.GasFeeAssetId])) < 0 {
		logx.Errorf("not enough balance of gas")
		return nil, errors.New("not enough balance of gas")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	// buyer sig
	msgHash, err := legendTxTypes.ComputeOfferMsgHash(txInfo.BuyOffer, hFunc)
	if err != nil {
		logx.Errorf("unable to compute tx hash: %s", err.Error())
		return nil, errors.New("internal error")
	}
	// verify signature
	if err := VerifySignature(txInfo.BuyOffer.Sig, msgHash, accountInfoMap[txInfo.BuyOffer.AccountIndex].PublicKey); err != nil {
		return nil, err
	}
	// seller sig
	hFunc.Reset()
	msgHash, err = legendTxTypes.ComputeOfferMsgHash(txInfo.SellOffer, hFunc)
	if err != nil {
		logx.Errorf("unable to compute tx hash: %s", err.Error())
		return nil, errors.New("internal error")
	}
	// verify signature
	if err := VerifySignature(txInfo.SellOffer.Sig, msgHash, accountInfoMap[txInfo.SellOffer.AccountIndex].PublicKey); err != nil {
		return nil, err
	}
	// submitter hash
	hFunc.Reset()
	msgHash, err = legendTxTypes.ComputeAtomicMatchMsgHash(txInfo, hFunc)
	if err != nil {
		logx.Errorf("unable to compute tx hash: %s", err.Error())
		return nil, errors.New("internal error")
	}
	// verify submitter signature
	if err := VerifySignature(txInfo.Sig, msgHash, accountInfoMap[txInfo.AccountIndex].PublicKey); err != nil {
		return nil, err
	}
	// compute tx details
	// from account asset gas
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfoMap[txInfo.AccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// buyer asset A
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.BuyOffer.AccountIndex,
		AccountName:  accountInfoMap[txInfo.BuyOffer.AccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, ffmath.Neg(txInfo.BuyOffer.AssetAmount), ZeroBigInt, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// buyer offer
	buyerOfferIndex := txInfo.BuyOffer.OfferId % OfferPerAsset
	if accountInfoMap[txInfo.BuyOffer.AccountIndex].AssetInfo[buyerOfferAssetId] != nil {
		oBuyerOffer := accountInfoMap[txInfo.BuyOffer.AccountIndex].AssetInfo[buyerOfferAssetId].OfferCanceledOrFinalized
		// verify whether buyer offer id is valid for use
		if oBuyerOffer.Bit(int(buyerOfferIndex)) == 1 {
			logx.Errorf("account %d offer index %d is already in use", txInfo.BuyOffer.AccountIndex, buyerOfferIndex)
			return nil, errors.New("invalid buyer offer id")
		}

		nBuyerOffer := new(big.Int).SetBit(oBuyerOffer, int(buyerOfferIndex), 1)
		order++
		txDetails = append(txDetails, &MempoolTxDetail{
			AssetId:      buyerOfferAssetId,
			AssetType:    GeneralAssetType,
			AccountIndex: txInfo.BuyOffer.AccountIndex,
			AccountName:  accountInfoMap[txInfo.BuyOffer.AccountIndex].AccountName,
			BalanceDelta: commonAsset.ConstructAccountAsset(
				buyerOfferAssetId, ZeroBigInt, ZeroBigInt, nBuyerOffer,
			).String(),
			Order:        order,
			AccountOrder: accountOrder,
		})
	}
	// seller asset A
	// treasury fee
	treasuryFee, err := util.CleanPackedAmount(ffmath.Div(
		ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(txInfo.SellOffer.TreasuryRate)),
		big.NewInt(TenThousand)))
	if err != nil {
		logx.Errorf("unable to compute treasury fee: %s", err.Error())
		return nil, errors.New("internal error")
	}
	// creator fee
	creatorFee, err := util.CleanPackedAmount(ffmath.Div(
		ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(nftInfo.CreatorTreasuryRate)),
		big.NewInt(TenThousand)))
	if err != nil {
		logx.Errorf("unable to compute creator fee: %s", err.Error())
		return nil, errors.New("internal error")
	}
	// set tx info
	txInfo.CreatorAmount = creatorFee
	txInfo.TreasuryAmount = treasuryFee
	// seller amount
	sellerDeltaAmount := ffmath.Sub(txInfo.SellOffer.AssetAmount, ffmath.Add(treasuryFee, creatorFee))
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.SellOffer.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.SellOffer.AccountIndex,
		AccountName:  accountInfoMap[txInfo.SellOffer.AccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.SellOffer.AssetId, sellerDeltaAmount, ZeroBigInt, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// seller offer
	sellerOfferIndex := txInfo.SellOffer.OfferId % OfferPerAsset
	if accountInfoMap[txInfo.SellOffer.AccountIndex].AssetInfo[sellerOfferAssetId] != nil {
		oSellerOffer := accountInfoMap[txInfo.SellOffer.AccountIndex].AssetInfo[sellerOfferAssetId].OfferCanceledOrFinalized
		// verify whether buyer offer id is valid for use
		if oSellerOffer.Bit(int(sellerOfferIndex)) == 1 {
			logx.Errorf("account %d offer index %d is already in use", txInfo.SellOffer.AccountIndex, sellerOfferIndex)
			return nil, errors.New("invalid seller offer id")
		}
		nSellerOffer := new(big.Int).SetBit(oSellerOffer, int(sellerOfferIndex), 1)
		order++
		txDetails = append(txDetails, &MempoolTxDetail{
			AssetId:      sellerOfferAssetId,
			AssetType:    GeneralAssetType,
			AccountIndex: txInfo.SellOffer.AccountIndex,
			AccountName:  accountInfoMap[txInfo.SellOffer.AccountIndex].AccountName,
			BalanceDelta: commonAsset.ConstructAccountAsset(
				sellerOfferAssetId, ZeroBigInt, ZeroBigInt, nSellerOffer,
			).String(),
			Order:        order,
			AccountOrder: accountOrder,
		})
	}
	// creator fee
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: nftInfo.CreatorAccountIndex,
		AccountName:  accountInfoMap[nftInfo.CreatorAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, creatorFee, ZeroBigInt, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// nft info
	newNftInfo := &NftInfo{
		NftIndex:            nftInfo.NftIndex,
		CreatorAccountIndex: nftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.BuyOffer.AccountIndex,
		NftContentHash:      nftInfo.NftContentHash,
		NftL1TokenId:        nftInfo.NftL1TokenId,
		NftL1Address:        nftInfo.NftL1Address,
		CreatorTreasuryRate: nftInfo.CreatorTreasuryRate,
		CollectionId:        nftInfo.CollectionId,
	}
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      nftInfo.NftIndex,
		AssetType:    NftAssetType,
		AccountIndex: commonConstant.NilTxAccountIndex,
		AccountName:  commonConstant.NilAccountName,
		BalanceDelta: newNftInfo.String(),
		Order:        order,
		AccountOrder: commonConstant.NilAccountOrder,
	})
	// gas account asset A - treasury fee
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  accountInfoMap[txInfo.GasAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, treasuryFee, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// gas account asset gas
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  accountInfoMap[txInfo.GasAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	return txDetails, nil
}
