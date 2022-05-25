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
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zeromicro/go-zero/core/logx"
)

func ConstructProverInfo(
	oTx *Tx,
	accountModel AccountModel,
) (
	accountKeys []int64,
	proverAccountMap map[int64]*ProverAccountInfo,
	proverLiquidityInfo *ProverLiquidityInfo,
	proverNftInfo *ProverNftInfo,
	err error,
) {
	var (
		// init account asset map, because if we have the same asset detail, the before will be the after of the last one
		accountAssetMap = make(map[int64]map[int64]*AccountAsset)
		isKeyExist      = make(map[int64]bool)
	)
	// init prover account map
	proverAccountMap = make(map[int64]*ProverAccountInfo)
	if oTx.AccountIndex != commonConstant.NilTxAccountIndex {
		// get account info
		if proverAccountMap[oTx.AccountIndex] == nil {
			accountInfo, err := accountModel.GetConfirmedAccountByAccountIndex(oTx.AccountIndex)
			if err != nil {
				logx.Errorf("[ConstructProverInfo] unable to get valid account by index: %s", err.Error())
				return nil, nil, nil, nil, err
			}
			// if tx type == registerZNS, it means that the account should be empty
			if oTx.TxType != commonTx.TxTypeRegisterZns {
				proverAccountMap[oTx.AccountIndex] = new(ProverAccountInfo)
				// get current nonce
				if oTx.Nonce != commonConstant.NilNonce {
					accountInfo.Nonce = oTx.Nonce - 1
				}
				proverAccountMap[oTx.AccountIndex].AccountInfo = accountInfo
			}
		}
		// set account key
		accountKeys = append(accountKeys, oTx.AccountIndex)
		isKeyExist[oTx.AccountIndex] = true
	}
	for _, txDetail := range oTx.TxDetails {
		switch txDetail.AssetType {
		case commonAsset.GeneralAssetType:
			if !isKeyExist[txDetail.AccountIndex] {
				accountKeys = append(accountKeys, txDetail.AccountIndex)
				isKeyExist[txDetail.AccountIndex] = true
			}
			// get account info
			if proverAccountMap[txDetail.AccountIndex] == nil {
				accountInfo, err := accountModel.GetConfirmedAccountByAccountIndex(txDetail.AccountIndex)
				if err != nil {
					logx.Errorf("[ConstructProverInfo] unable to get valid account by index: %s", err.Error())
					return nil, nil, nil, nil, err
				}
				proverAccountMap[txDetail.AccountIndex] = new(ProverAccountInfo)
				// get current nonce
				accountInfo.Nonce = txDetail.Nonce
				proverAccountMap[txDetail.AccountIndex].AccountInfo = accountInfo
			}
			if accountAssetMap[txDetail.AccountIndex] == nil {
				accountAssetMap[txDetail.AccountIndex] = make(map[int64]*AccountAsset)
			}
			if accountAssetMap[txDetail.AccountIndex][txDetail.AssetId] == nil {
				// set account before info
				oAsset, err := commonAsset.ParseAccountAsset(txDetail.Balance)
				if err != nil {
					logx.Errorf("[ConstructProverInfo] unable to parse account asset:%s", err.Error())
					return nil, nil, nil, nil, err
				}
				proverAccountMap[txDetail.AccountIndex].AccountAssets = append(
					proverAccountMap[txDetail.AccountIndex].AccountAssets,
					oAsset,
				)
			} else {
				// set account before info
				proverAccountMap[txDetail.AccountIndex].AccountAssets = append(
					proverAccountMap[txDetail.AccountIndex].AccountAssets,
					&AccountAsset{
						AssetId:                  accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].AssetId,
						Balance:                  accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].Balance,
						LpAmount:                 accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].LpAmount,
						OfferCanceledOrFinalized: accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].OfferCanceledOrFinalized,
					},
				)
			}
			// set tx detail
			proverAccountMap[txDetail.AccountIndex].AssetsRelatedTxDetails = append(
				proverAccountMap[txDetail.AccountIndex].AssetsRelatedTxDetails,
				txDetail,
			)
			// update asset info
			newBalance, err := commonAsset.ComputeNewBalance(txDetail.AssetType, txDetail.Balance, txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[ConstructProverInfo] unable to compute new balance: %s", err.Error())
				return nil, nil, nil, nil, err
			}
			nAsset, err := commonAsset.ParseAccountAsset(newBalance)
			if err != nil {
				logx.Errorf("[ConstructProverInfo] unable to parse account asset:%s", err.Error())
				return nil, nil, nil, nil, err
			}
			accountAssetMap[txDetail.AccountIndex][txDetail.AssetId] = nAsset
			break
		case commonAsset.LiquidityAssetType:
			proverLiquidityInfo = new(ProverLiquidityInfo)
			proverLiquidityInfo.LiquidityRelatedTxDetail = txDetail
			poolInfo, err := commonAsset.ParseLiquidityInfo(txDetail.Balance)
			if err != nil {
				logx.Errorf("[ConstructProverInfo] unable to parse pool info: %s", err.Error())
				return nil, nil, nil, nil, err
			}
			proverLiquidityInfo.LiquidityInfo = poolInfo
			break
		case commonAsset.NftAssetType:
			proverNftInfo = new(ProverNftInfo)
			proverNftInfo.NftRelatedTxDetail = txDetail
			nftInfo, err := commonAsset.ParseNftInfo(txDetail.Balance)
			if err != nil {
				logx.Errorf("[ConstructProverInfo] unable to parse nft info: %s", err.Error())
				return nil, nil, nil, nil, err
			}
			proverNftInfo.NftInfo = nftInfo
			break
		case commonAsset.CollectionNonceAssetType:
			if !isKeyExist[txDetail.AccountIndex] {
				accountKeys = append(accountKeys, txDetail.AccountIndex)
				isKeyExist[txDetail.AccountIndex] = true
			}
			// get account info
			if proverAccountMap[txDetail.AccountIndex] == nil {
				accountInfo, err := accountModel.GetConfirmedAccountByAccountIndex(txDetail.AccountIndex)
				if err != nil {
					logx.Errorf("[ConstructProverInfo] unable to get valid account by index: %s", err.Error())
					return nil, nil, nil, nil, err
				}
				proverAccountMap[txDetail.AccountIndex] = new(ProverAccountInfo)
				// get current nonce
				accountInfo.Nonce = txDetail.Nonce
				accountInfo.CollectionNonce = txDetail.CollectionNonce
				proverAccountMap[txDetail.AccountIndex].AccountInfo = accountInfo
			} else {
				proverAccountMap[txDetail.AccountIndex].AccountInfo.Nonce = txDetail.Nonce
				proverAccountMap[txDetail.AccountIndex].AccountInfo.CollectionNonce = txDetail.CollectionNonce
			}
			break
		default:
			logx.Errorf("[ConstructProverInfo] invalid asset type")
			return nil, nil, nil, nil,
				errors.New("[ConstructProverInfo] invalid asset type")
		}
	}
	return accountKeys, proverAccountMap, proverLiquidityInfo, proverNftInfo, nil
}
