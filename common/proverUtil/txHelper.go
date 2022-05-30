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
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zeromicro/go-zero/core/logx"
)

func ConstructProverInfo(
	oTx *Tx,
	accountModel AccountModel,
) (
	accountKeys []int64,
	proverAccounts []*ProverAccountInfo,
	proverLiquidityInfo *ProverLiquidityInfo,
	proverNftInfo *ProverNftInfo,
	err error,
) {
	var (
		// init account asset map, because if we have the same asset detail, the before will be the after of the last one
		accountAssetMap  = make(map[int64]map[int64]*AccountAsset)
		accountMap       = make(map[int64]*Account)
		lastAccountOrder = int64(-2)
		accountCount     = -1
	)
	// init prover account map
	if oTx.TxType == commonTx.TxTypeRegisterZns {
		accountKeys = append(accountKeys, oTx.AccountIndex)
	}
	for _, txDetail := range oTx.TxDetails {
		switch txDetail.AssetType {
		case commonAsset.GeneralAssetType:
			// get account info
			if accountMap[txDetail.AccountIndex] == nil {
				accountInfo, err := accountModel.GetConfirmedAccountByAccountIndex(txDetail.AccountIndex)
				if err != nil {
					logx.Errorf("[ConstructProverInfo] unable to get valid account by index: %s", err.Error())
					return nil, nil, nil, nil, err
				}
				// get current nonce
				accountInfo.Nonce = txDetail.Nonce
				accountMap[txDetail.AccountIndex] = accountInfo
			} else {
				if lastAccountOrder != txDetail.AccountOrder {
					if oTx.AccountIndex == txDetail.AccountIndex {
						accountMap[txDetail.AccountIndex].Nonce = oTx.Nonce
					}
				}
			}
			if lastAccountOrder != txDetail.AccountOrder {
				accountKeys = append(accountKeys, txDetail.AccountIndex)
				lastAccountOrder = txDetail.AccountOrder
				proverAccounts = append(proverAccounts, &ProverAccountInfo{
					AccountInfo: &Account{
						AccountIndex:    accountMap[txDetail.AccountIndex].AccountIndex,
						AccountName:     accountMap[txDetail.AccountIndex].AccountName,
						PublicKey:       accountMap[txDetail.AccountIndex].PublicKey,
						AccountNameHash: accountMap[txDetail.AccountIndex].AccountNameHash,
						L1Address:       accountMap[txDetail.AccountIndex].L1Address,
						Nonce:           accountMap[txDetail.AccountIndex].Nonce,
						CollectionNonce: txDetail.CollectionNonce,
						AssetInfo:       accountMap[txDetail.AccountIndex].AssetInfo,
						AssetRoot:       accountMap[txDetail.AccountIndex].AssetRoot,
						Status:          accountMap[txDetail.AccountIndex].Status,
					},
				})
				accountCount++
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
				proverAccounts[accountCount].AccountAssets = append(
					proverAccounts[accountCount].AccountAssets,
					oAsset,
				)
			} else {
				// set account before info
				proverAccounts[accountCount].AccountAssets = append(
					proverAccounts[accountCount].AccountAssets,
					&AccountAsset{
						AssetId:                  accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].AssetId,
						Balance:                  accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].Balance,
						LpAmount:                 accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].LpAmount,
						OfferCanceledOrFinalized: accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].OfferCanceledOrFinalized,
					},
				)
			}
			// set tx detail
			proverAccounts[accountCount].AssetsRelatedTxDetails = append(
				proverAccounts[accountCount].AssetsRelatedTxDetails,
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
			// get account info
			if accountMap[txDetail.AccountIndex] == nil {
				accountInfo, err := accountModel.GetConfirmedAccountByAccountIndex(txDetail.AccountIndex)
				if err != nil {
					logx.Errorf("[ConstructProverInfo] unable to get valid account by index: %s", err.Error())
					return nil, nil, nil, nil, err
				}
				// get current nonce
				accountInfo.Nonce = txDetail.Nonce
				accountInfo.CollectionNonce = txDetail.CollectionNonce
				accountMap[txDetail.AccountIndex] = accountInfo
			} else {
				accountMap[txDetail.AccountIndex].Nonce = txDetail.Nonce
				accountMap[txDetail.AccountIndex].CollectionNonce = txDetail.CollectionNonce
			}
			break
		default:
			logx.Errorf("[ConstructProverInfo] invalid asset type")
			return nil, nil, nil, nil,
				errors.New("[ConstructProverInfo] invalid asset type")
		}
	}
	return accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, nil
}
