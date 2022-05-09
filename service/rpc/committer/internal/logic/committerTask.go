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
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-core/common/general/model/liquidityPair"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/tree"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/committer/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"log"
	"math"
	"math/big"
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
			nftAssetsHistoryMap       = make(map[int64]*L2NftHistory)
			accountsHistoryMap        = make(map[int64]*commonAsset.FormatAccountHistoryInfo)
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

			liquidityPairMap = make(map[int64]*liquidityPair.LiquidityPair)
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
			if accountsMap[mempoolTx.AccountIndex] == nil {
				accountsMap[mempoolTx.AccountIndex], err = ctx.AccountModel.GetAccountByAccountIndex(mempoolTx.AccountIndex)
				if err != nil {
					logx.Errorf("[CommitterTask] get account by account index: %s", err.Error())
					return err
				}
			}
			if accountsHistoryMap[mempoolTx.AccountIndex] == nil {
				accountHistoryInfo, err := ctx.AccountHistoryModel.GetLatestAccountInfoByAccountIndex(mempoolTx.AccountIndex)
				if err != nil {
					if err == ErrNotFound {
						// set new account history
						accountHistoryInfo = &AccountHistory{
							AccountIndex:  accountsMap[mempoolTx.AccountIndex].AccountIndex,
							Nonce:         accountsMap[mempoolTx.AccountIndex].Nonce,
							AssetInfo:     accountsMap[mempoolTx.AccountIndex].AssetInfo,
							AssetRoot:     accountsMap[mempoolTx.AccountIndex].AssetRoot,
							LiquidityInfo: accountsMap[mempoolTx.AccountIndex].LiquidityInfo,
							LiquidityRoot: accountsMap[mempoolTx.AccountIndex].LiquidityRoot,
							Status:        account.AccountHistoryStatusConfirmed,
							L2BlockHeight: currentBlockHeight,
						}
						accountsHistoryMap[mempoolTx.AccountIndex], err = commonAsset.ToFormatAccountHistoryInfo(accountHistoryInfo)
						if err != nil {
							logx.Errorf("[CommitterTask] cannot convert to format account history info: %s", err.Error())
							return err
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
						accountsMap[mempoolTx.AccountIndex].Status = account.AccountStatusConfirmed
						pendingUpdateAccountIndex[mempoolTx.AccountIndex] = true
						// update account tree
						if int64(len(accountStateTrees)) != accountHistoryInfo.AccountIndex {
							logx.Errorf("[CommitterTask] invalid account index")
							return errors.New("[CommitterTask] invalid account index")
						}
						emptyStateTree, err := tree.NewEmptyAccountStateTree()
						if err != nil {
							logx.Errorf("[CommitterTask] unable to new empty account state tree")
							return err
						}
						accountStateTrees = append(accountStateTrees, emptyStateTree)
						nAccountLeafHash, err := tree.ComputeAccountLeafHash(
							accountHistoryInfo.AccountIndex,
							accountsMap[mempoolTx.AccountIndex].AccountName,
							accountsMap[mempoolTx.AccountIndex].PublicKey,
							accountHistoryInfo.Nonce,
							accountStateTrees[accountHistoryInfo.AccountIndex].AssetTree.RootNode.Value,
							accountStateTrees[accountHistoryInfo.AccountIndex].LiquidityTree.RootNode.Value,
						)
						if err != nil {
							log.Println("[CommitterTask] unable to compute account leaf:", err)
							return err
						}
						err = accountTree.Update(accountHistoryInfo.AccountIndex, nAccountLeafHash)
						if err != nil {
							log.Println("[CommitterTask] unable to update account tree:", err)
							return err
						}
					}
					accountsHistoryMap[mempoolTx.AccountIndex], err = commonAsset.ToFormatAccountHistoryInfo(accountHistoryInfo)
					if err != nil {
						logx.Errorf("[CommitterTask] unable convert to format account history info: %s", err.Error())
						return err
					}
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
				txDetails         []*tx.TxDetail
			)
			for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
				if accountAssetExist[mempoolTxDetail.AccountIndex] == nil {
					accountAssetExist[mempoolTxDetail.AccountIndex] = make(map[int64]bool)
				}
				if accountsMap[mempoolTxDetail.AccountIndex] == nil {
					accountsMap[mempoolTxDetail.AccountIndex], err = ctx.AccountModel.GetAccountByAccountIndex(mempoolTxDetail.AccountIndex)
					if err != nil {
						logx.Errorf("[CommitterTask] get account by account index: %s", err.Error())
						return err
					}
				}
				if accountsHistoryMap[mempoolTxDetail.AccountIndex] == nil {
					accountHistoryInfo, err := ctx.AccountHistoryModel.GetLatestAccountInfoByAccountIndex(mempoolTxDetail.AccountIndex)
					if err != nil {
						if err == ErrNotFound {
							// set new account history
							accountHistoryInfo = &AccountHistory{
								AccountIndex:  accountsMap[mempoolTxDetail.AccountIndex].AccountIndex,
								Nonce:         accountsMap[mempoolTxDetail.AccountIndex].Nonce,
								AssetInfo:     accountsMap[mempoolTxDetail.AccountIndex].AssetInfo,
								AssetRoot:     accountsMap[mempoolTxDetail.AccountIndex].AssetRoot,
								LiquidityInfo: accountsMap[mempoolTxDetail.AccountIndex].LiquidityInfo,
								LiquidityRoot: accountsMap[mempoolTxDetail.AccountIndex].LiquidityRoot,
								Status:        account.AccountHistoryStatusConfirmed,
								L2BlockHeight: currentBlockHeight,
							}
						} else {
							logx.Errorf("[CommitterTask] cannot get related account info from history table: %s", err.Error())
							return err
						}
					}
					accountsHistoryMap[mempoolTxDetail.AccountIndex], err = commonAsset.ToFormatAccountHistoryInfo(accountHistoryInfo)
					if err != nil {
						logx.Errorf("[CommitterTask] cannot convert to format account history info: %s", err.Error())
						return err
					}
				}
				var (
					baseBalance string
				)
				// check balance
				switch mempoolTxDetail.AssetType {
				case GeneralAssetType:
					if accountsHistoryMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId] == "" {
						accountsHistoryMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId] = ZeroBigIntString
					}
					// get latest account asset info
					baseBalance = accountsHistoryMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId]
					// compute new balance
					nBalance, err := util.ComputeNewBalance(GeneralAssetType, baseBalance, mempoolTxDetail.BalanceDelta)
					if err != nil {
						logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
						return err
					}
					nBalanceInt, isValid := new(big.Int).SetString(nBalance, 10)
					if !isValid {
						logx.Errorf("[CommitterTask] unable to parse big int")
						return errors.New("[CommitterTask] unable to parse big int")
					}
					// check balance is valid
					if nBalanceInt.Cmp(util.ZeroBigInt) < 0 {
						// mark this transaction as invalid transaction
						mempoolTx.Status = mempool.FailTxStatus
						mempoolTx.L2BlockHeight = currentBlockHeight
						pendingMempoolTxs = append(pendingMempoolTxs, mempoolTx)
						continue
					}
					accountsHistoryMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId] = nBalance
					accountsHistoryMap[mempoolTxDetail.AccountIndex].L2BlockHeight = currentBlockHeight
					pendingNewAccountIndex[mempoolTxDetail.AccountIndex] = true
					// update account state tree
					nAssetLeaf, err := tree.ComputeAccountAssetLeafHash(
						mempoolTxDetail.AssetId,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId])
					if err != nil {
						log.Println("[CommitterTask] unable to compute new account asset leaf:", err)
						return err
					}
					err = accountStateTrees[mempoolTxDetail.AccountIndex].AssetTree.Update(mempoolTxDetail.AssetId, nAssetLeaf)
					if err != nil {
						log.Println("[CommitterTask] unable to update asset tree:", err)
						return err
					}

					accountsHistoryMap[mempoolTxDetail.AccountIndex].AssetRoot = common.Bytes2Hex(
						accountStateTrees[mempoolTxDetail.AccountIndex].AssetTree.RootNode.Value)

					// update account tree
					nAccountLeafHash, err := tree.ComputeAccountLeafHash(
						mempoolTxDetail.AccountIndex,
						accountsMap[mempoolTxDetail.AccountIndex].AccountName,
						accountsMap[mempoolTxDetail.AccountIndex].PublicKey,
						accountsMap[mempoolTxDetail.AccountIndex].Nonce,
						accountStateTrees[mempoolTxDetail.AccountIndex].AssetTree.RootNode.Value,
						accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.RootNode.Value,
					)
					if err != nil {
						log.Println("[CommitterTask] unable to compute account leaf:", err)
						return err
					}
					err = accountTree.Update(mempoolTxDetail.AccountIndex, nAccountLeafHash)
					if err != nil {
						log.Println("[CommitterTask] unable to update account tree:", err)
						return err
					}

					break
				case LiquidityAssetType:
					if accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId] == nil {
						// get pair info from liquidityPair
						if liquidityPairMap[mempoolTxDetail.AssetId] == nil {
							liquidityPairMap[mempoolTxDetail.AssetId], err = ctx.LiquidityPairModel.GetLiquidityPairByIndex(mempoolTxDetail.AssetId)
							if err != nil {
								logx.Errorf("[CommitterTask] cannot get liquidity pair by index: %s", err.Error())
								return err
							}
						}
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId] = &commonAsset.Liquidity{
							PairIndex: mempoolTxDetail.AssetId,
							AssetAId:  liquidityPairMap[mempoolTxDetail.AssetId].AssetAId,
							AssetA:    ZeroBigIntString,
							AssetBId:  liquidityPairMap[mempoolTxDetail.AssetId].AssetBId,
							AssetB:    ZeroBigIntString,
							LpAmount:  ZeroBigIntString,
						}
					}
					poolInfo, err := util.ConstructPoolInfo(
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetA,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetB,
					)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to construct pool info: %s", err.Error())
						return err
					}
					baseBalance = poolInfo.String()
					// compute new balance
					nBalance, err := util.ComputeNewBalance(
						LiquidityAssetType, baseBalance, mempoolTxDetail.BalanceDelta)
					if err != nil {
						logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
						return err
					}
					newPoolInfo, err := util.ParsePoolInfo(nBalance)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to parse pair info: %s", err.Error())
						return err
					}
					accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetA =
						newPoolInfo.AssetAAmount.String()
					accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetB =
						newPoolInfo.AssetBAmount.String()
					accountsHistoryMap[mempoolTxDetail.AccountIndex].L2BlockHeight = currentBlockHeight
					pendingNewAccountIndex[mempoolTxDetail.AccountIndex] = true

					// update account state tree
					nLiquidityAssetLeaf, err := tree.ComputeAccountLiquidityAssetLeafHash(
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].PairIndex,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetAId,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetA,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetBId,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetB,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].LpAmount)
					if err != nil {
						log.Println("[CommitterTask] unable to compute new account liquidity leaf:", err)
						return err
					}
					err = accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.Update(mempoolTxDetail.AssetId, nLiquidityAssetLeaf)
					if err != nil {
						log.Println("[CommitterTask] unable to update liquidity tree:", err)
						return err
					}
					// update root node
					accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityRoot = common.Bytes2Hex(
						accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.RootNode.Value)

					// update account tree
					nAccountLeafHash, err := tree.ComputeAccountLeafHash(
						mempoolTxDetail.AccountIndex, accountsMap[mempoolTxDetail.AccountIndex].AccountName, accountsMap[mempoolTxDetail.AccountIndex].PublicKey, accountsMap[mempoolTxDetail.AccountIndex].Nonce,
						accountStateTrees[mempoolTxDetail.AccountIndex].AssetTree.RootNode.Value,
						accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.RootNode.Value,
					)
					if err != nil {
						log.Println("[UpdateDepositAccount] unable to compute account leaf:", err)
						return err
					}
					err = accountTree.Update(mempoolTxDetail.AccountIndex, nAccountLeafHash)
					if err != nil {
						log.Println("[UpdateDepositAccount] unable to update account tree:", err)
						return err
					}
					break
				case LiquidityLpAssetType:
					if accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId] == nil {
						// get pair info from liquidityPair
						if liquidityPairMap[mempoolTxDetail.AssetId] == nil {
							liquidityPairMap[mempoolTxDetail.AssetId], err = ctx.LiquidityPairModel.GetLiquidityPairByIndex(mempoolTxDetail.AssetId)
							if err != nil {
								logx.Errorf("[CommitterTask] cannot get liquidity pair by index: %s", err.Error())
								return err
							}
						}
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId] = &commonAsset.Liquidity{
							PairIndex: mempoolTxDetail.AssetId,
							AssetAId:  liquidityPairMap[mempoolTxDetail.AssetId].AssetAId,
							AssetA:    ZeroBigIntString,
							AssetBId:  liquidityPairMap[mempoolTxDetail.AssetId].AssetBId,
							AssetB:    ZeroBigIntString,
							LpAmount:  ZeroBigIntString,
						}
					}
					baseBalance = accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].LpAmount
					// compute new balance
					nBalance, err := util.ComputeNewBalance(
						LiquidityLpAssetType, baseBalance, mempoolTxDetail.BalanceDelta)
					if err != nil {
						logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
						return err
					}
					accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].LpAmount = nBalance
					accountsHistoryMap[mempoolTxDetail.AccountIndex].L2BlockHeight = currentBlockHeight
					pendingNewAccountIndex[mempoolTxDetail.AccountIndex] = true

					// update account state tree
					nLiquidityAssetLeaf, err := tree.ComputeAccountLiquidityAssetLeafHash(
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].PairIndex,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetAId,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetA,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetBId,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].AssetB,
						accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityInfo[mempoolTxDetail.AssetId].LpAmount)
					if err != nil {
						log.Println("[CommitterTask] unable to compute new account liquidity leaf:", err)
						return err
					}
					err = accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.Update(mempoolTxDetail.AssetId, nLiquidityAssetLeaf)
					if err != nil {
						log.Println("[CommitterTask] unable to update liquidity tree:", err)
						return err
					}

					// update root node
					accountsHistoryMap[mempoolTxDetail.AccountIndex].LiquidityRoot = common.Bytes2Hex(
						accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.RootNode.Value)

					// update account tree
					nAccountLeafHash, err := tree.ComputeAccountLeafHash(
						mempoolTxDetail.AccountIndex, accountsMap[mempoolTxDetail.AccountIndex].AccountName, accountsMap[mempoolTxDetail.AccountIndex].PublicKey, accountsMap[mempoolTxDetail.AccountIndex].Nonce,
						accountStateTrees[mempoolTxDetail.AccountIndex].AssetTree.RootNode.Value,
						accountStateTrees[mempoolTxDetail.AccountIndex].LiquidityTree.RootNode.Value,
					)
					if err != nil {
						log.Println("[UpdateDepositAccount] unable to compute account leaf:", err)
						return err
					}
					err = accountTree.Update(mempoolTxDetail.AccountIndex, nAccountLeafHash)
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
					// compute new balance
					nBalance, err := util.ComputeNewBalance(
						NftAssetType, nftInfo.String(), mempoolTxDetail.BalanceDelta)
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
				txDetails = append(txDetails, &tx.TxDetail{
					AssetId:      mempoolTxDetail.AssetId,
					AssetType:    mempoolTxDetail.AssetType,
					AccountIndex: mempoolTxDetail.AccountIndex,
					AccountName:  mempoolTxDetail.AccountName,
					Balance:      baseBalance,
					BalanceDelta: mempoolTxDetail.BalanceDelta,
				})
			}
			// add into mempool tx
			pendingMempoolTxs = append(pendingMempoolTxs, mempoolTx)
			// update mempool tx info
			mempoolTx.L2BlockHeight = currentBlockHeight
			mempoolTx.Status = mempool.SuccessTxStatus
			// construct tx
			// account root
			hFunc := mimc.NewMiMC()
			hFunc.Write(accountTree.RootNode.Value)
			hFunc.Write(nftTree.RootNode.Value)
			accountRoot := common.Bytes2Hex(hFunc.Sum(nil))
			finalAccountRoot = accountRoot
			oTx := ConvertMempoolTxToTx(mempoolTx, txDetails, accountRoot, currentBlockHeight)
			txs = append(txs, oTx)
		}
		// construct assets history
		var (
			pendingNewNftAssetsHistory   []*L2NftHistory
			pendingNewAccountHistory     []*AccountHistory
			pendingUpdateAccounts        []*Account
			pendingUpdatedAccountHistory []*AccountHistory
		)
		for _, nftAssetHistory := range nftAssetsHistoryMap {
			pendingNewNftAssetsHistory = append(pendingNewNftAssetsHistory, nftAssetHistory)
		}
		for accountIndex, flag := range pendingNewAccountIndex {
			if !flag {
				continue
			}
			accountHistoryInfo, err := commonAsset.FromFormatAccountHistoryInfo(accountsHistoryMap[accountIndex])
			if err != nil {
				logx.Errorf("[CommitterTask] unable to ")
				return err
			}
			newAccountHistoryInfo := &account.AccountHistory{
				AccountIndex:  accountHistoryInfo.AccountIndex,
				Nonce:         accountHistoryInfo.Nonce,
				AssetInfo:     accountHistoryInfo.AssetInfo,
				AssetRoot:     accountHistoryInfo.AssetRoot,
				LiquidityInfo: accountHistoryInfo.LiquidityInfo,
				LiquidityRoot: accountHistoryInfo.LiquidityRoot,
				Status:        accountHistoryInfo.Status,
				L2BlockHeight: accountHistoryInfo.L2BlockHeight,
			}
			pendingNewAccountHistory = append(pendingNewAccountHistory, newAccountHistoryInfo)
		}
		for accountIndex, flag := range pendingUpdateAccountIndex {
			if !flag {
				continue
			}
			accountHistoryInfo, err := commonAsset.FromFormatAccountHistoryInfo(accountsHistoryMap[accountIndex])
			if err != nil {
				logx.Errorf("[CommitterTask] unable to ")
				return err
			}
			pendingUpdatedAccountHistory = append(pendingUpdatedAccountHistory, accountHistoryInfo)
			pendingUpdateAccounts = append(pendingUpdateAccounts, accountsMap[accountIndex])
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
			pendingUpdateAccounts,
			pendingNewAccountHistory, pendingUpdatedAccountHistory)
		if err != nil {
			logx.Errorf("[CommitterTask] unable to create block for committer: %s", err.Error())
			return err
		}
	}
	return nil
}

/**
handleTxPubdata: handle different layer-1 txs
*/
func handleTxPubdata(mempoolTx *MempoolTx, oldPendingOnchainOperationsHash []byte) (
	priorityOperation int64,
	newPendingOnchainOperationsHash []byte,
	pubData []byte,
	err error,
) {
	priorityOperation = 0
	newPendingOnchainOperationsHash = oldPendingOnchainOperationsHash
	switch mempoolTx.TxType {
	case TxTypeRegisterZns:
		pubData, err = util.ConvertTxToRegisterZNSPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to registerZNS pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		priorityOperation++
		break
	case TxTypeDeposit:
		pubData, err = util.ConvertTxToDepositPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to deposit pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		priorityOperation++
		break
	case TxTypeDepositNft:
		pubData, err = util.ConvertTxToDepositNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to deposit nft pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		priorityOperation++
		break
	case TxTypeTransfer:
		pubData, err = util.ConvertTxToTransferPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to transfer pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeSwap:
		pubData, err = util.ConvertTxToSwapPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to swap pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeAddLiquidity:
		pubData, err = util.ConvertTxToAddLiquidityPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to add liquidity pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeRemoveLiquidity:
		pubData, err = util.ConvertTxToRemoveLiquidityPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to remove liquidity pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeMintNft:
		pubData, err = util.ConvertTxToMintNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to mint nft pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeSetNftPrice:
		pubData, err = util.ConvertTxToSetNftPricePubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to set nft price pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeBuyNft:
		pubData, err = util.ConvertTxToBuyNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to buy nft pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		break
	case TxTypeWithdraw:
		pubData, err = util.ConvertTxToWithdrawPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to withdraw pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		newPendingOnchainOperationsHash = util.ConcatKeccakHash(oldPendingOnchainOperationsHash, pubData)
		break
	case TxTypeWithdrawNft:
		pubData, err = util.ConvertTxToWithdrawNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to withdraw nft pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		newPendingOnchainOperationsHash = util.ConcatKeccakHash(oldPendingOnchainOperationsHash, pubData)
		break
	case TxTypeFullExit:
		pubData, err = util.ConvertTxToFullExitPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubdata] unable to convert tx to full exit pub data")
			return priorityOperation, newPendingOnchainOperationsHash, pubData, err
		}
		priorityOperation++
		newPendingOnchainOperationsHash = util.ConcatKeccakHash(oldPendingOnchainOperationsHash, pubData)
		break
	case TxTypeFullExitNft:
		pubData, err = util.ConvertTxToFullExitNftPubData(mempoolTx)
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
