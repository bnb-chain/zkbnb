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
 */

package logic

import (
	"errors"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/asset"
	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/tree"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/committer/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"log"
	"math"
	"time"
)

func CommitterTask(
	ctx *svc.ServiceContext,
	lastCommitTimeStamp time.Time,
	accountTree *tree.Tree,
	nftTree *tree.Tree,
	accountStateTrees []*tree.AccountStateTree,
) error {
	// Get Txs from Mempool
	mempoolTxs, err := ctx.MempoolModel.GetMempoolTxsListForCommitter()
	if err != nil {
		if err == ErrNotFound {
			logx.Info("[CommitterTask] no tx in mempool")
			return nil
		} else {
			logx.Error("[CommitterTask] unable to get tx in mempool")
			return err
		}
	}

	var nTxs = len(mempoolTxs)
	logx.Infof("[CommitterTask] Mempool txs number : %d", nTxs)

	// get current block height
	currentBlockHeight, err := ctx.BlockModel.GetCurrentBlockHeight()
	if err != nil && err != ErrNotFound {
		logx.Error("[CommitterTask] err when get current block height")
		return err
	}
	// get last block info
	lastBlock, err := ctx.BlockModel.GetBlockByBlockHeight(currentBlockHeight)
	if err != nil {
		logx.Errorf("[CommitterTask] unable to get block by height: %s", err.Error())
		return err
	}
	// handle txs
	// check how many blocks
	blocksSize := int(math.Ceil(float64(nTxs) / float64(MaxTxsAmountPerBlock)))

	// accountsMap store the map from account index to accountInfo, decrease the duplicated query from Account Model
	var accountsMap = make(map[int64]*Account)

	for i := 0; i < blocksSize; i++ {
		// Check time stamp
		var now = time.Now()
		if now.Unix()-lastCommitTimeStamp.Unix() < MaxCommitterInterval {
			// if time is less than MaxCommitterInterval (15 minutes for now)
			// and remaining txs number( equals to "nTxs - (i + 1) * MaxTxsAmountPerBlock") is less than MaxTxsAmountPerBlock
			if nTxs-i*MaxTxsAmountPerBlock < MaxTxsAmountPerBlock {
				logx.Infof("[CommitterTask] not enough transactions")
				return errors.New("[CommitterTask] not enough transactions")
			}
		}

		var (
			assetsHistoryMap          = make(map[string]*AccountAssetHistory)
			liquidityAssetsHistoryMap = make(map[string]*AccountLiquidityHistory)
			nftAssetsHistoryMap       = make(map[int64]*L2NftHistory)
			accountsHistoryMap        = make(map[int64]*AccountHistory)
			pendingUpdateAccountIndex = make(map[int64]bool)
			pendingNewAccountIndex    = make(map[int64]bool)

			// block txs
			txs []*Tx
			// final account root
			finalAccountRoot string
			// pub data
			pubdata []byte
			// onchain tx info
			priorityOperations           int64
			pendingOnchainOperationsHash []byte
			pendingMempoolTxs            []*MempoolTx
		)
		// write default string into pending onchain operations hash
		pendingOnchainOperationsHash = common.FromHex(util.EmptyStringKeccak)
		// handle each transaction
		currentBlockHeight += 1

		for j := 0; j < MaxTxsAmountPerBlock; j++ {
			// if not full block, just break
			if i*MaxTxsAmountPerBlock+j >= nTxs {
				break
			}
			var (
				pendingPriorityOperation int64
				pendingPubdata           []byte
			)
			// get mempool tx
			mempoolTx := mempoolTxs[i*MaxTxsAmountPerBlock+j]
			pendingMempoolTxs = append(pendingMempoolTxs, mempoolTx)
			// handle tx pub data
			pendingPriorityOperation, pendingOnchainOperationsHash, pendingPubdata, err = handleTxPubdata(mempoolTx, pendingOnchainOperationsHash)
			if err != nil {
				logx.Errorf("[CommitterTask] unable to handle l1 tx: %s", err.Error())
				return err
			}
			// compute new priority operations
			priorityOperations += pendingPriorityOperation
			// add pub data from tx
			pubdata = append(pubdata, pendingPubdata...)

			// get related account info
			if accountsHistoryMap[mempoolTx.AccountIndex] == nil {
				accountHistoryInfo, err := ctx.AccountHistoryModel.GetLatestAccountInfoByAccountIndex(mempoolTx.AccountIndex)
				if err != nil {
					if err == ErrNotFound {
						accountInfo, err := ctx.AccountModel.GetAccountByAccountIndex(mempoolTx.AccountIndex)
						// if we cannot get any info from account table, return error
						if err != nil {
							logx.Errorf("[CommitterTask] unable to get account info: %s", err.Error())
							return err
						}
						// set new account history
						accountsHistoryMap[mempoolTx.AccountIndex] = &AccountHistory{
							AccountIndex:    accountInfo.AccountIndex,
							AccountName:     accountInfo.AccountName,
							AccountNameHash: accountInfo.AccountNameHash,
							PublicKey:       accountInfo.PublicKey,
							L1Address:       accountInfo.L1Address,
							Nonce:           accountInfo.Nonce,
							Status:          account.AccountHistoryStatusConfirmed,
							L2BlockHeight:   currentBlockHeight,
						}
					} else {
						logx.Errorf("[CommitterTask] cannot get related account info from history table: %s", err.Error())
						return err
					}
				} else {
					// it means that just make the register, haven't confirmed by committer, need up
					if accountHistoryInfo.Status == account.AccountHistoryStatusPending {
						if mempoolTx.TxType != TxTypeRegisterZns {
							logx.Errorf("[CommitterTask] first transaction should be registerZNS")
							return errors.New("[CommitterTask] first transaction should be registerZNS")
						}
						accountHistoryInfo.Status = account.AccountHistoryStatusConfirmed
						accountHistoryInfo.L2BlockHeight = currentBlockHeight
						pendingUpdateAccountIndex[mempoolTx.AccountIndex] = true
					}
					accountsHistoryMap[mempoolTx.AccountIndex] = accountHistoryInfo
				}
			}
			// check if we need to update nonce(create new account history)
			if mempoolTx.Nonce != -1 {
				// check nonce, the latest nonce should be previous nonce + 1
				if mempoolTx.Nonce != accountsHistoryMap[mempoolTx.AccountIndex].Nonce+1 {
					logx.Errorf("[CommitterTask] invalid nonce")
					return errors.New("[CommitterTask] invalid nonce")
				}
				// update nonce first
				accountsHistoryMap[mempoolTx.AccountIndex].Nonce = mempoolTx.Nonce
				// check for update or create
				if !pendingUpdateAccountIndex[mempoolTx.AccountIndex] {
					pendingNewAccountIndex[mempoolTx.AccountIndex] = true
				}
			}

			// check mempool tx details are correct
			var (
				accountAssetExist = make(map[int64]map[int64]bool)
			)
			for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
				if accountAssetExist[mempoolTxDetail.AccountIndex] == nil {
					accountAssetExist[mempoolTxDetail.AccountIndex] = make(map[int64]bool)
				}
				// check balance
				switch mempoolTxDetail.AssetType {
				case GeneralAssetType:
					key := util.GetAccountAssetUniqueKey(mempoolTxDetail.AccountIndex, mempoolTxDetail.AssetId)
					// query for related assetInfo
					// in order to get the latest asset info
					if assetsHistoryMap[key] == nil {
						var resAccountSingleAsset *AccountAsset
						rowsAffected, assetHistory, err := ctx.AccountAssetHistoryModel.GetLatestAccountAssetByIndexAndAssetId(mempoolTxDetail.AccountIndex, mempoolTxDetail.AssetId)
						if err != nil {
							logx.Errorf("[CommitterTask] unable to get single account assetInfo: %s", err.Error())
							return err
						}
						if rowsAffected == 0 {
							resAccountSingleAsset, err = ctx.AccountAssetModel.GetSingleAccountAsset(mempoolTxDetail.AccountIndex, mempoolTxDetail.AssetId)

							if err != nil {
								if err != asset.ErrNotFound {
									errInfo := fmt.Sprintf("[CommitterTask] unable to get single account assetInfo %s. Invalid accountIndex/assetId %v/%v",
										err.Error(), mempoolTxDetail.AccountIndex, mempoolTxDetail.AssetId)
									logx.Error(errInfo)
									return errors.New(errInfo)
								} else {
									// init nil AccountSingleAsset
									resAccountSingleAsset = &asset.AccountAsset{
										AccountIndex: mempoolTxDetail.AccountIndex,
										AssetId:      mempoolTxDetail.AssetId,
										Balance:      ZeroBigIntString,
									}
								}
							}
						} else {
							resAccountSingleAsset = AssetHistoryToAsset(assetHistory)
						}
						assetsHistoryMap[key] = AssetToAssetHistory(resAccountSingleAsset, currentBlockHeight)
					}
					// update assetInfo history
					assetInfo := assetsHistoryMap[key]
					// check balance
					if !accountAssetExist[mempoolTxDetail.AccountIndex][mempoolTxDetail.AssetId] {
						accountAssetExist[mempoolTxDetail.AccountIndex][mempoolTxDetail.AssetId] = true
						if assetInfo.Balance != mempoolTxDetail.Balance {
							logx.Errorf("[CommitterTask] invalid balance")
							return errors.New("[CommitterTask] invalid balance")
						}
					}
					// compute new balance
					nBalance, err := util.ComputeNewBalance(GeneralAssetType, mempoolTxDetail.Balance, mempoolTxDetail.BalanceDelta)
					if err != nil {
						logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
						return err
					}
					assetInfo.Balance = nBalance

					// update account state tree
					nAssetLeaf, err := tree.ComputeAccountAssetLeafHash(assetInfo.AssetId, assetInfo.Balance)
					if err != nil {
						log.Println("[CommitterTask] unable to compute new account asset leaf:", err)
						return err
					}
					err = accountStateTrees[mempoolTxDetail.AccountIndex].AssetTree.Update(mempoolTxDetail.AssetId, nAssetLeaf)
					if err != nil {
						log.Println("[CommitterTask] unable to update asset tree:", err)
						return err
					}

					// update account tree
					// if accountInfo not exists, query from account Model
					if accountsMap[mempoolTxDetail.AccountIndex] == nil {
						accInfo, err := ctx.AccountHistoryModel.GetLatestAccountInfoByAccountIndex(mempoolTxDetail.AccountIndex)
						if err != nil {
							log.Println("[CommitterTask] GetAccountByAccountIndex error: ", err)
							return err
						}
						accountsMap[mempoolTxDetail.AccountIndex] = &Account{
							AccountIndex: accInfo.AccountIndex,
							AccountName:  accInfo.AccountName,
							PublicKey:    accInfo.PublicKey,
							L1Address:    accInfo.L1Address,
							Nonce:        accInfo.Nonce,
						}
					}
					nAccountLeafHash, err := tree.ComputeAccountLeafHash(
						assetInfo.AccountIndex, accountsMap[mempoolTxDetail.AccountIndex].AccountName, accountsMap[mempoolTxDetail.AccountIndex].PublicKey, accountsMap[mempoolTxDetail.AccountIndex].Nonce,
						accountStateTrees[mempoolTxDetail.AccountIndex].AssetTree.RootNode.Value,
						accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.RootNode.Value,
					)
					if err != nil {
						log.Println("[UpdateDepositAccount] unable to compute account leaf:", err)
						return err
					}
					err = accountTree.Update(assetInfo.AccountIndex, nAccountLeafHash)
					if err != nil {
						log.Println("[UpdateDepositAccount] unable to update account tree:", err)
						return err
					}

					break
				case LiquidityAssetType:
					key := util.GetPoolLiquidityUniqueKey(
						mempoolTxDetail.AccountIndex, mempoolTxDetail.AssetId)
					// query for related assetInfo
					if liquidityAssetsHistoryMap[key] == nil {
						var liquidityAsset *AccountLiquidity
						rowsAffected, assetHistory, err := ctx.LiquidityAssetHistoryModel.GetLatestLiquidityAsset(
							uint32(mempoolTxDetail.AccountIndex), uint32(mempoolTxDetail.AssetId),
						)
						if err != nil {
							logx.Errorf("[CommitterTask] unable to get single account assetInfo: %s", err.Error())
							return err
						}
						if rowsAffected == 0 {
							liquidityAsset, err = ctx.LiquidityAssetModel.GetLiquidityByAccountIndexAndPairIndex(
								uint32(mempoolTxDetail.AccountIndex), uint32(mempoolTxDetail.AssetId),
							)
							if err != nil {
								if err == ErrNotFound {
									liquidityAsset = &AccountLiquidity{
										AccountIndex: mempoolTxDetail.AccountIndex,
										PairIndex:    mempoolTxDetail.AssetId,
										AssetAId:     mempoolTx.AssetAId,
										AssetA:       "0",
										AssetBId:     mempoolTx.AssetBId,
										AssetB:       "0",
										LpAmount:     "0",
									}
								} else {
									logx.Errorf("[CommitterTask] unable to get account liquidity: %s", err.Error())
									return err
								}
							}
						} else {
							liquidityAsset = LiquidityAssetHistoryToLiquidityAsset(assetHistory)
						}
						liquidityAssetsHistoryMap[key] = LiquidityAssetToLiquidityAssetHistory(liquidityAsset, currentBlockHeight)
					}
					// update assetInfo history
					liquidityAsset := liquidityAssetsHistoryMap[key]
					// special design for deposit
					// check balance
					poolInfo, err := util.ConstructPoolInfo(liquidityAsset.AssetA, liquidityAsset.AssetB)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to construct pair info: %s", err.Error())
						return err
					}
					nPoolInfo, err := util.ParsePoolInfo(mempoolTxDetail.Balance)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to parse pair info: %s", err.Error())
						return err
					}
					isEqualPair := util.IsEqualPoolInfo(poolInfo, nPoolInfo)
					if !isEqualPair {
						logx.Errorf("[CommitterTask] not equal pair info")
						return errors.New("[CommitterTask] not equal pair info")
					}
					// compute new balance
					nBalance, err := util.ComputeNewBalance(
						LiquidityAssetType, mempoolTxDetail.Balance, mempoolTxDetail.BalanceDelta)
					if err != nil {
						logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
						return err
					}
					newPoolInfo, err := util.ParsePoolInfo(nBalance)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to parse pair info: %s", err.Error())
						return err
					}
					liquidityAsset.AssetA = newPoolInfo.AssetAAmount.String()
					liquidityAsset.AssetB = newPoolInfo.AssetBAmount.String()

					// update account state tree
					nLiquidityAssetLeaf, err := tree.ComputeAccountLiquidityAssetLeafHash(
						liquidityAsset.PairIndex,
						liquidityAsset.AssetAId, liquidityAsset.AssetA,
						liquidityAsset.AssetBId, liquidityAsset.AssetB,
						liquidityAsset.LpAmount)
					if err != nil {
						log.Println("[CommitterTask] unable to compute new account liquidity leaf:", err)
						return err
					}
					err = accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.Update(mempoolTxDetail.AssetId, nLiquidityAssetLeaf)
					if err != nil {
						log.Println("[CommitterTask] unable to update liquidity tree:", err)
						return err
					}

					// update account tree
					// if accountInfo not exists, query from account Model
					if accountsMap[mempoolTxDetail.AccountIndex] == nil {
						// if account is in register list
						accInfo, err := ctx.AccountHistoryModel.GetLatestAccountInfoByAccountIndex(mempoolTxDetail.AccountIndex)
						if err != nil {
							log.Println("[CommitterTask] GetAccountByAccountIndex error: ", err)
							return err
						}
						accountsMap[mempoolTxDetail.AccountIndex] = &account.Account{
							AccountIndex: accInfo.AccountIndex,
							AccountName:  accInfo.AccountName,
							PublicKey:    accInfo.PublicKey,
							L1Address:    accInfo.L1Address,
							Nonce:        accInfo.Nonce,
						}
					}
					nAccountLeafHash, err := tree.ComputeAccountLeafHash(
						mempoolTxDetail.AccountIndex, accountsMap[mempoolTxDetail.AccountIndex].AccountName, accountsMap[mempoolTxDetail.AccountIndex].PublicKey, accountsMap[mempoolTxDetail.AccountIndex].Nonce,
						accountStateTrees[mempoolTxDetail.AccountIndex].AssetTree.RootNode.Value,
						accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.RootNode.Value,
					)
					if err != nil {
						log.Println("[UpdateDepositAccount] unable to compute account leaf:", err)
						return err
					}
					err = accountTree.Update(liquidityAsset.AccountIndex, nAccountLeafHash)
					if err != nil {
						log.Println("[UpdateDepositAccount] unable to update account tree:", err)
						return err
					}

					break
				case LiquidityLpAssetType:
					key := util.GetAccountLPUniqueKey(
						mempoolTxDetail.AccountIndex, mempoolTxDetail.AssetId)
					// query for related assetInfo
					if liquidityAssetsHistoryMap[key] == nil {
						var liquidityAsset *AccountLiquidity
						rowsAffected, assetHistory, err := ctx.LiquidityAssetHistoryModel.GetLatestLiquidityAsset(
							uint32(mempoolTxDetail.AccountIndex), uint32(mempoolTxDetail.AssetId),
						)
						if err != nil {
							logx.Errorf("[CommitterTask] unable to get single account assetInfo: %s", err.Error())
							return err
						}
						if rowsAffected == 0 {
							liquidityAsset, err = ctx.LiquidityAssetModel.GetLiquidityByAccountIndexAndPairIndex(
								uint32(mempoolTxDetail.AccountIndex), uint32(mempoolTxDetail.AssetId),
							)
							if err != nil {
								if err == ErrNotFound {
									liquidityAsset = &AccountLiquidity{
										AccountIndex: mempoolTxDetail.AccountIndex,
										PairIndex:    mempoolTxDetail.AssetId,
										AssetAId:     mempoolTx.AssetAId,
										AssetA:       "0",
										AssetBId:     mempoolTx.AssetBId,
										AssetB:       "0",
										LpAmount:     "0",
									}
								} else {
									logx.Errorf("[CommitterTask] unable to get account liquidity: %s", err.Error())
									return err
								}
							}
						} else {
							liquidityAsset = LiquidityAssetHistoryToLiquidityAsset(assetHistory)
						}
						liquidityAssetsHistoryMap[key] = LiquidityAssetToLiquidityAssetHistory(liquidityAsset, currentBlockHeight)
					}
					// update assetInfo history
					liquidityAsset := liquidityAssetsHistoryMap[key]
					// special design for deposit
					// check balance
					if liquidityAsset.LpAmount != mempoolTxDetail.Balance {
						logx.Errorf("[CommitterTask] invalid lp amount")
						return errors.New("[CommitterTask] invalid lp amount")
					}
					// compute new balance
					nBalance, err := util.ComputeNewBalance(
						LiquidityLpAssetType, mempoolTxDetail.Balance, mempoolTxDetail.BalanceDelta)
					if err != nil {
						logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
						return err
					}
					liquidityAsset.LpAmount = nBalance

					// update account state tree
					nLiquidityAssetLeaf, err := tree.ComputeAccountLiquidityAssetLeafHash(
						liquidityAsset.PairIndex,
						liquidityAsset.AssetAId, liquidityAsset.AssetA,
						liquidityAsset.AssetBId, liquidityAsset.AssetB,
						liquidityAsset.LpAmount)
					if err != nil {
						log.Println("[CommitterTask] unable to compute new account liquidity leaf:", err)
						return err
					}
					err = accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.Update(mempoolTxDetail.AssetId, nLiquidityAssetLeaf)
					if err != nil {
						log.Println("[CommitterTask] unable to update liquidity tree:", err)
						return err
					}

					// update account tree
					// if accountInfo not exists, query from account Model
					if accountsMap[mempoolTxDetail.AccountIndex] == nil {
						// if account is in register list
						accInfo, err := ctx.AccountHistoryModel.GetLatestAccountInfoByAccountIndex(mempoolTxDetail.AccountIndex)
						if err != nil {
							log.Println("[CommitterTask] GetAccountByAccountIndex error: ", err)
							return err
						}
						accountsMap[mempoolTxDetail.AccountIndex] = &account.Account{
							AccountIndex: accInfo.AccountIndex,
							AccountName:  accInfo.AccountName,
							PublicKey:    accInfo.PublicKey,
							L1Address:    accInfo.L1Address,
							Nonce:        accInfo.Nonce,
						}
					}
					nAccountLeafHash, err := tree.ComputeAccountLeafHash(
						mempoolTxDetail.AccountIndex, accountsMap[mempoolTxDetail.AccountIndex].AccountName, accountsMap[mempoolTxDetail.AccountIndex].PublicKey, accountsMap[mempoolTxDetail.AccountIndex].Nonce,
						accountStateTrees[mempoolTxDetail.AccountIndex].AssetTree.RootNode.Value,
						accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.RootNode.Value,
					)
					if err != nil {
						log.Println("[UpdateDepositAccount] unable to compute account leaf:", err)
						return err
					}
					err = accountTree.Update(liquidityAsset.AccountIndex, nAccountLeafHash)
					if err != nil {
						log.Println("[UpdateDepositAccount] unable to update account tree:", err)
						return err
					}
					break
				case NftAssetType:
					if nftAssetsHistoryMap[mempoolTxDetail.AssetId] == nil {
						var nftAsset *L2Nft
						assetHistory, err := ctx.L2NftHistoryModel.GetLatestNftAsset(mempoolTxDetail.AssetId)
						if err != nil {
							if err != ErrNotFound {
								logx.Errorf("[CommitterTask] unable to get latest nft asset: %s", err.Error())
								return err
							} else {
								nftAsset, err = ctx.L2NftModel.GetNftAsset(mempoolTxDetail.AssetId)
								if err != nil {
									if err == ErrNotFound {
										emptyNftInfo := util.EmptyNftInfo(mempoolTxDetail.AssetId)
										nftAssetsHistoryMap[mempoolTxDetail.AssetId] = &L2NftHistory{
											NftIndex:            mempoolTxDetail.AssetId,
											CreatorAccountIndex: emptyNftInfo.CreatorAccountIndex,
											OwnerAccountIndex:   emptyNftInfo.OwnerAccountIndex,
											AssetId:             emptyNftInfo.AssetId,
											AssetAmount:         emptyNftInfo.AssetAmount,
											NftContentHash:      emptyNftInfo.NftContentHash,
											NftL1TokenId:        emptyNftInfo.NftL1TokenId,
											NftL1Address:        emptyNftInfo.NftL1Address,
											CollectionId:        commonConstant.NilCollectionId,
											L2BlockHeight:       currentBlockHeight,
										}
									} else {
										logx.Errorf("[CommitterTask] unable to get nft asset: %s", err.Error())
										return err
									}
								} else {
									nftAssetsHistoryMap[mempoolTxDetail.AssetId] = NftAssetToNftAssetHistory(nftAsset, currentBlockHeight)
								}
							}
						} else {
							nftAssetsHistoryMap[mempoolTxDetail.AssetId] = assetHistory
						}
					}
					// update nft asset history
					nftAsset := nftAssetsHistoryMap[mempoolTxDetail.AssetId]
					nftInfo, err := util.ConstructNftInfo(
						nftAsset.NftIndex, nftAsset.CreatorAccountIndex, nftAsset.OwnerAccountIndex, nftAsset.AssetId,
						nftAsset.AssetAmount, nftAsset.NftContentHash, nftAsset.NftL1TokenId, nftAsset.NftL1Address,
					)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to construct nft info: %s", err.Error())
						return err
					}
					// check balance
					if nftInfo.String() != mempoolTxDetail.Balance {
						logx.Errorf("[CommitterTask] invalid nft info")
						return errors.New("[CommitterTask] invalid nft info")
					}
					// compute new balance
					nBalance, err := util.ComputeNewBalance(
						NftAssetType, mempoolTxDetail.Balance, mempoolTxDetail.BalanceDelta)
					if err != nil {
						logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
						return err
					}
					newNftInfo, err := util.ParseNftInfo(nBalance)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to parse pair info: %s", err.Error())
						return err
					}
					nftAsset = &L2NftHistory{
						Model:               nftAsset.Model,
						NftIndex:            newNftInfo.NftIndex,
						CreatorAccountIndex: newNftInfo.CreatorAccountIndex,
						OwnerAccountIndex:   newNftInfo.OwnerAccountIndex,
						AssetId:             newNftInfo.AssetId,
						AssetAmount:         newNftInfo.AssetAmount,
						NftContentHash:      newNftInfo.NftContentHash,
						NftL1TokenId:        newNftInfo.NftL1TokenId,
						NftL1Address:        newNftInfo.NftL1Address,
						CollectionId:        nftAsset.CollectionId,
						Status:              nftAsset.Status,
						L2BlockHeight:       nftAsset.L2BlockHeight,
					}

					// update nft tree
					nNftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
						nftAsset.NftIndex, nftAsset.CreatorAccountIndex, nftAsset.NftContentHash,
						nftAsset.AssetId, nftAsset.AssetAmount, nftAsset.NftL1Address, nftAsset.NftL1TokenId,
					)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to compute new nft asset leaf: %s", err)
						return err
					}
					err = nftTree.Update(mempoolTxDetail.AssetId, nNftAssetLeaf)
					if err != nil {
						log.Println("[CommitterTask] unable to update nft tree:", err)
						return err
					}
					break
				default:
					logx.Error("[CommitterTask] invalid tx type")
					return errors.New("[CommitterTask] invalid tx type")
				}
			}
			// update mempool tx info
			mempoolTx.L2BlockHeight = currentBlockHeight
			mempoolTx.Status = MempoolTxHandledTxStatus
			// construct tx
			// account root
			accountRoot := common.Bytes2Hex(accountTree.RootNode.Value)
			finalAccountRoot = accountRoot
			oTx := ConvertMempoolTxToTx(mempoolTx, accountRoot, currentBlockHeight)
			txs = append(txs, oTx)
		}
		// construct assets history
		var (
			pendingNewAssetsHistory          []*AccountAssetHistory
			pendingNewLiquidityAssetsHistory []*AccountLiquidityHistory
			pendingNewNftAssetsHistory       []*L2NftHistory
			pendingNewAccountHistory         []*AccountHistory
			pendingUpdatedAccountHistory     []*AccountHistory
		)
		for _, assetHistory := range assetsHistoryMap {
			pendingNewAssetsHistory = append(pendingNewAssetsHistory, assetHistory)
		}
		for _, liquidityAssetHistory := range liquidityAssetsHistoryMap {
			pendingNewLiquidityAssetsHistory = append(pendingNewLiquidityAssetsHistory, liquidityAssetHistory)
		}
		for _, nftAssetHistory := range nftAssetsHistoryMap {
			pendingNewNftAssetsHistory = append(pendingNewNftAssetsHistory, nftAssetHistory)
		}
		for accountIndex, flag := range pendingNewAccountIndex {
			if !flag {
				continue
			}
			pendingNewAccountHistory = append(pendingNewAccountHistory, accountsHistoryMap[accountIndex])
		}
		for accountIndex, flag := range pendingUpdateAccountIndex {
			if !flag {
				continue
			}
			pendingUpdatedAccountHistory = append(pendingUpdatedAccountHistory, accountsHistoryMap[accountIndex])
		}

		// compute block commitment
		createAt := time.Now().UnixMilli()
		// TODO commitment
		commitment := util.CreateBlockCommitment(lastBlock.BlockHeight, currentBlockHeight, pubdata)
		// construct block
		createAtTime := time.UnixMilli(createAt)
		if len(txs) == 0 {
			logx.Errorf("[CommitterTask] error with txs size")
			return errors.New("[CommitterTask] error with txs size")
		}
		hFunc := mimc.NewMiMC()
		hFunc.Write(accountTree.RootNode.Value)
		hFunc.Write(nftTree.RootNode.Value)
		finalAccountRoot = common.Bytes2Hex(hFunc.Sum(nil))
		oBlock := &Block{
			Model: gorm.Model{
				CreatedAt: createAtTime,
			},
			BlockCommitment:              commitment,
			BlockHeight:                  currentBlockHeight,
			AccountRoot:                  finalAccountRoot,
			PriorityOperations:           priorityOperations,
			PendingOnchainOperationsHash: common.Bytes2Hex(pendingOnchainOperationsHash),
			Txs:                          txs,
			BlockStatus:                  block.StatusPending,
		}

		// create block for committer
		//create block, history, update mempool txs, create new l1 amount infos
		err = ctx.BlockModel.CreateBlockForCommitter(
			oBlock, pendingMempoolTxs,
			pendingNewAssetsHistory, pendingNewLiquidityAssetsHistory,
			pendingNewAccountHistory, pendingUpdatedAccountHistory)
		if err != nil {
			logx.Errorf("[CommitterTask] unable to create block for committer: %s", err.Error())
			return err
		}

		// TODO reset global map
		//_, err = ctx.GlobalRPC.ResetGlobalMap(context.Background(), &globalRPCProto.ReqResetGlobalMap{})
		//if err != nil {
		//	logx.Errorf("[CommitterTask] unable to reset global map")
		//}
	}
	return nil
}

/**
handleTxPubdata: handle different layer-1 txs
*/
func handleTxPubdata(mempoolTx *MempoolTx, oldPendingOnchainOperationsHash []byte) (
	priorityOperation int64,
	newPendingOnchainOperationsHash []byte,
	pubdata []byte,
	err error,
) {
	priorityOperation = 0
	newPendingOnchainOperationsHash = oldPendingOnchainOperationsHash
	switch mempoolTx.TxType {
	case TxTypeRegisterZns:
		pubData, err := util.ConvertTxToRegisterZNSPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to registerZNS pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		priorityOperation++
		break
	case TxTypeDeposit:
		pubData, err := util.ConvertTxToDepositPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to deposit pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		priorityOperation++
		break
	case TxTypeDepositNft:
		pubData, err := util.ConvertTxToDepositNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to deposit nft pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		priorityOperation++
		break
	case TxTypeTransfer:
		pubData, err := util.ConvertTxToTransferPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to transfer pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeSwap:
		pubData, err := util.ConvertTxToSwapPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to swap pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeAddLiquidity:
		pubData, err := util.ConvertTxToAddLiquidityPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to add liquidity pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeRemoveLiquidity:
		pubData, err := util.ConvertTxToRemoveLiquidityPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to remove liquidity pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeMintNft:
		pubData, err := util.ConvertTxToMintNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to mint nft pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeSetNftPrice:
		pubData, err := util.ConvertTxToSetNftPricePubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to set nft price pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeBuyNft:
		pubData, err := util.ConvertTxToBuyNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to buy nft pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeWithdraw:
		pubData, err := util.ConvertTxToWithdrawPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to withdraw pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		newPendingOnchainOperationsHash = util.ConcatKeccakHash(oldPendingOnchainOperationsHash, pubData)
		break
	case TxTypeWithdrawNft:
		pubData, err := util.ConvertTxToWithdrawNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to withdraw nft pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		newPendingOnchainOperationsHash = util.ConcatKeccakHash(oldPendingOnchainOperationsHash, pubData)
		break
	case TxTypeFullExit:
		pubData, err := util.ConvertTxToFullExitPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to full exit pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		priorityOperation++
		newPendingOnchainOperationsHash = util.ConcatKeccakHash(oldPendingOnchainOperationsHash, pubData)
		break
	case TxTypeFullExitNft:
		pubData, err := util.ConvertTxToFullExitNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to full exit nft pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		priorityOperation++
		newPendingOnchainOperationsHash = util.ConcatKeccakHash(oldPendingOnchainOperationsHash, pubData)
		break
	default:
		logx.Errorf("[handleTxPubdata] invalid tx type")
		return priorityOperation, newPendingOnchainOperationsHash, nil, errors.New("[handleTxPubdata] invalid tx type")
	}
	return priorityOperation, newPendingOnchainOperationsHash, nil, nil
}
