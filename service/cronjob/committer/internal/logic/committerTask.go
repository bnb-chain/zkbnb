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
 */

package logic

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/cronjob/committer/internal/svc"
)

func CommitterTask(ctx *svc.ServiceContext, lastCommitTimeStamp *time.Time,
	accountTree *tree.Tree, liquidityTree *tree.Tree, nftTree *tree.Tree, accountAssetTrees *[]*tree.Tree) error {
	mempoolTxs, err := ctx.MempoolModel.GetMempoolTxsListForCommitter()
	if err != nil {
		if err == ErrNotFound {
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
	MaxTxsAmountPerBlock := TxsAmountPerBlock[len(TxsAmountPerBlock)-1]
	blocksSize := int(math.Ceil(float64(nTxs) / float64(MaxTxsAmountPerBlock)))

	// accountMap store the map from account index to accountInfo, decrease the duplicated query from Account Model
	var (
		accountMap   = make(map[int64]*FormatAccountInfo)
		liquidityMap = make(map[int64]*Liquidity)
		nftMap       = make(map[int64]*L2Nft)
		oldStateRoot = lastBlock.StateRoot
	)
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
			pendingUpdateAccountIndexMap   = make(map[int64]bool)
			pendingUpdateLiquidityIndexMap = make(map[int64]bool)

			pendingUpdateNftIndexMap     = make(map[int64]bool)
			pendingNewNftIndexMap        = make(map[int64]bool)
			pendingNewNftWithdrawHistory []*nft.L2NftWithdrawHistory

			// block txs
			txs []*Tx
			// final account root
			finalStateRoot string
			// pub data
			pubData []byte
			// onchain tx info
			priorityOperations              int64
			pubDataOffset                   []uint32
			pendingOnChainOperationsPubData [][]byte
			pendingOnChainOperationsHash    []byte
			pendingMempoolTxs               []*MempoolTx
			pendingDeleteMempoolTxs         []*MempoolTx
		)
		// write default string into pending onchain operations hash
		pendingOnChainOperationsHash = common.FromHex(util.EmptyStringKeccak)
		// handle each transaction
		currentBlockHeight += 1

		// compute block commitment
		createdAt := time.Now().UnixMilli()

		for j := 0; j < MaxTxsAmountPerBlock; j++ {
			// if not full block, just break
			if i*MaxTxsAmountPerBlock+j >= nTxs {
				break
			}
			var (
				pendingPriorityOperation int64
				newCollectionNonce       = commonConstant.NilCollectionId
			)
			// get mempool tx
			mempoolTx := mempoolTxs[i*MaxTxsAmountPerBlock+j]
			// handle tx pub data
			pendingPriorityOperation, pendingOnChainOperationsPubData, pendingOnChainOperationsHash, pubData, pubDataOffset, err =
				handleTxPubData(
					mempoolTx,
					pubData,
					pendingOnChainOperationsPubData,
					pendingOnChainOperationsHash,
					pubDataOffset,
				)
			if err != nil {
				logx.Errorf("[CommitterTask] unable to handle l1 tx: %s", err.Error())
				return err
			}
			// compute new priority operations
			priorityOperations += pendingPriorityOperation

			// get related account info
			if mempoolTx.AccountIndex != commonConstant.NilTxAccountIndex {
				if accountMap[mempoolTx.AccountIndex] == nil {
					accountInfo, err := ctx.AccountModel.GetAccountByAccountIndex(mempoolTx.AccountIndex)
					if err != nil {
						logx.Errorf("[CommitterTask] get account by account index: %s", err.Error())
						return err
					}
					accountMap[mempoolTx.AccountIndex], err = commonAsset.ToFormatAccountInfo(accountInfo)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to format account info: %s", err.Error())
						return err
					}
				}
				// handle registerZNS tx
				pendingUpdateAccountIndexMap[mempoolTx.AccountIndex] = true
				if accountMap[mempoolTx.AccountIndex].Status == account.AccountStatusPending {
					if mempoolTx.TxType != TxTypeRegisterZns {
						logx.Errorf("[CommitterTask] first transaction should be registerZNS")
						return errors.New("[CommitterTask] first transaction should be registerZNS")
					}
					accountMap[mempoolTx.AccountIndex].Status = account.AccountStatusConfirmed
					pendingUpdateAccountIndexMap[mempoolTx.AccountIndex] = true
					// update account tree
					if int64(len(*accountAssetTrees)) != mempoolTx.AccountIndex {
						logx.Errorf("[CommitterTask] invalid account index")
						return errors.New("[CommitterTask] invalid account index")
					}
					emptyAssetTree, err := tree.NewEmptyAccountAssetTree()
					if err != nil {
						logx.Errorf("[CommitterTask] unable to new empty account state tree")
						return err
					}
					*accountAssetTrees = append(*accountAssetTrees, emptyAssetTree)
					nAccountLeafHash, err := tree.ComputeAccountLeafHash(
						accountMap[mempoolTx.AccountIndex].AccountNameHash,
						accountMap[mempoolTx.AccountIndex].PublicKey,
						accountMap[mempoolTx.AccountIndex].Nonce,
						accountMap[mempoolTx.AccountIndex].CollectionNonce,
						(*accountAssetTrees)[mempoolTx.AccountIndex].RootNode.Value,
					)
					if err != nil {
						log.Println("[CommitterTask] unable to compute account leaf:", err)
						return err
					}
					err = accountTree.Update(mempoolTx.AccountIndex, nAccountLeafHash)
					if err != nil {
						log.Println("[CommitterTask] unable to update account tree:", err)
						return err
					}
				}
			}
			// check if the tx is still valid
			if mempoolTx.ExpiredAt != commonConstant.NilExpiredAt {
				if mempoolTx.ExpiredAt < createdAt {
					mempoolTx.Status = mempool.FailTxStatus
					mempoolTx.L2BlockHeight = currentBlockHeight
					pendingDeleteMempoolTxs = append(pendingDeleteMempoolTxs, mempoolTx)
					continue
				}
			}
			if mempoolTx.Nonce != commonConstant.NilNonce {
				// check nonce, the latest nonce should be previous nonce + 1
				if mempoolTx.Nonce != accountMap[mempoolTx.AccountIndex].Nonce+1 {
					mempoolTx.Status = mempool.FailTxStatus
					mempoolTx.L2BlockHeight = currentBlockHeight
					pendingDeleteMempoolTxs = append(pendingDeleteMempoolTxs, mempoolTx)
					continue
					//logx.Errorf("[CommitterTask] invalid nonce")
					//return errors.New("[CommitterTask] invalid nonce")
				}
			}
			// check mempool tx details are correct
			var (
				txDetails []*tx.TxDetail
			)
			for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
				if mempoolTxDetail.AccountIndex != commonConstant.NilTxAccountIndex {
					pendingUpdateAccountIndexMap[mempoolTxDetail.AccountIndex] = true
					if accountMap[mempoolTxDetail.AccountIndex] == nil {
						accountInfo, err := ctx.AccountModel.GetAccountByAccountIndex(mempoolTxDetail.AccountIndex)
						if err != nil {
							logx.Errorf("[CommitterTask] get account by account index: %s", err.Error())
							return err
						}
						accountMap[mempoolTxDetail.AccountIndex], err = commonAsset.ToFormatAccountInfo(accountInfo)
						if err != nil {
							logx.Errorf("[CommitterTask] unable to format account info: %s", err.Error())
							return err
						}
					}
				}
				var (
					baseBalance string
				)
				// check balance
				switch mempoolTxDetail.AssetType {
				case GeneralAssetType:
					if accountMap[mempoolTxDetail.AccountIndex].AssetInfo == nil {
						accountMap[mempoolTxDetail.AccountIndex].AssetInfo = make(map[int64]*commonAsset.AccountAsset)
					}
					if accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId] == nil {
						accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId] = &commonAsset.AccountAsset{
							AssetId:                  mempoolTxDetail.AssetId,
							Balance:                  ZeroBigInt,
							LpAmount:                 ZeroBigInt,
							OfferCanceledOrFinalized: ZeroBigInt,
						}
					}
					// get latest account asset info
					baseBalance = accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId].String()
					var (
						nBalance string
					)
					if mempoolTx.TxType == TxTypeFullExit {
						balanceDelta := &commonAsset.AccountAsset{
							AssetId:                  accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId].AssetId,
							Balance:                  ffmath.Neg(accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId].Balance),
							LpAmount:                 big.NewInt(0),
							OfferCanceledOrFinalized: big.NewInt(0),
						}
						// compute new balance
						nBalance, err = commonAsset.ComputeNewBalance(GeneralAssetType, baseBalance, balanceDelta.String())
						if err != nil {
							logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
							return err
						}
						mempoolTxDetail.BalanceDelta = balanceDelta.String()
						txInfo, err := commonTx.ParseFullExitTxInfo(mempoolTx.TxInfo)
						if err != nil {
							logx.Errorf("[CommitterTask] unable to parse full exit tx info: %s", err.Error())
							return err
						}
						txInfo.AssetAmount = accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId].Balance
						infoBytes, err := json.Marshal(txInfo)
						if err != nil {
							logx.Errorf("[CommitterTask] unable to marshal tx: %s", err.Error())
							return err
						}
						mempoolTx.TxInfo = string(infoBytes)
					} else {
						// compute new balance
						nBalance, err = commonAsset.ComputeNewBalance(GeneralAssetType, baseBalance, mempoolTxDetail.BalanceDelta)
						if err != nil {
							logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
							return err
						}
					}
					nAccountAsset, err := commonAsset.ParseAccountAsset(nBalance)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to parse account asset: %s", err.Error())
						return err
					}
					// check balance is valid
					if nAccountAsset.Balance.Cmp(util.ZeroBigInt) < 0 {
						// mark this transaction as invalid transaction
						mempoolTx.Status = mempool.FailTxStatus
						mempoolTx.L2BlockHeight = currentBlockHeight
						pendingDeleteMempoolTxs = append(pendingDeleteMempoolTxs, mempoolTx)
						continue
					}
					accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId] = nAccountAsset
					// update account state tree
					nAssetLeaf, err := tree.ComputeAccountAssetLeafHash(
						accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId].Balance.String(),
						accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId].LpAmount.String(),
						accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId].OfferCanceledOrFinalized.String(),
					)
					if err != nil {
						log.Println("[CommitterTask] unable to compute new account asset leaf:", err)
						return err
					}
					err = (*accountAssetTrees)[mempoolTxDetail.AccountIndex].Update(mempoolTxDetail.AssetId, nAssetLeaf)
					if err != nil {
						log.Println("[CommitterTask] unable to update asset tree:", err)
						return err
					}

					accountMap[mempoolTxDetail.AccountIndex].AssetRoot = common.Bytes2Hex(
						(*accountAssetTrees)[mempoolTxDetail.AccountIndex].RootNode.Value)

					break
				case LiquidityAssetType:
					pendingUpdateLiquidityIndexMap[mempoolTxDetail.AssetId] = true
					if liquidityMap[mempoolTxDetail.AssetId] == nil {
						liquidityMap[mempoolTxDetail.AssetId], err = ctx.LiquidityModel.GetLiquidityByPairIndex(mempoolTxDetail.AssetId)
						if err != nil {
							logx.Errorf("[CommitterTask] unable to get latest liquidity by pair index: %s", err.Error())
							return err
						}
					}
					var (
						poolInfo *commonAsset.LiquidityInfo
					)
					if mempoolTx.TxType == TxTypeCreatePair {
						poolInfo = commonAsset.EmptyLiquidityInfo(mempoolTxDetail.AssetId)
					} else {
						poolInfo, err = commonAsset.ConstructLiquidityInfo(
							liquidityMap[mempoolTxDetail.AssetId].PairIndex,
							liquidityMap[mempoolTxDetail.AssetId].AssetAId,
							liquidityMap[mempoolTxDetail.AssetId].AssetA,
							liquidityMap[mempoolTxDetail.AssetId].AssetBId,
							liquidityMap[mempoolTxDetail.AssetId].AssetB,
							liquidityMap[mempoolTxDetail.AssetId].LpAmount,
							liquidityMap[mempoolTxDetail.AssetId].KLast,
							liquidityMap[mempoolTxDetail.AssetId].FeeRate,
							liquidityMap[mempoolTxDetail.AssetId].TreasuryAccountIndex,
							liquidityMap[mempoolTxDetail.AssetId].TreasuryRate,
						)
						if err != nil {
							logx.Errorf("[CommitterTask] unable to construct pool info: %s", err.Error())
							return err
						}
					}
					baseBalance = poolInfo.String()
					// compute new balance
					nBalance, err := commonAsset.ComputeNewBalance(
						LiquidityAssetType, baseBalance, mempoolTxDetail.BalanceDelta)
					if err != nil {
						logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
						return err
					}
					nPoolInfo, err := commonAsset.ParseLiquidityInfo(nBalance)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to parse pair info: %s", err.Error())
						return err
					}
					// update liquidity info
					liquidityMap[mempoolTxDetail.AssetId] = &Liquidity{
						Model:                liquidityMap[mempoolTxDetail.AssetId].Model,
						PairIndex:            nPoolInfo.PairIndex,
						AssetAId:             liquidityMap[mempoolTxDetail.AssetId].AssetAId,
						AssetA:               nPoolInfo.AssetA.String(),
						AssetBId:             liquidityMap[mempoolTxDetail.AssetId].AssetBId,
						AssetB:               nPoolInfo.AssetB.String(),
						LpAmount:             nPoolInfo.LpAmount.String(),
						KLast:                nPoolInfo.KLast.String(),
						FeeRate:              nPoolInfo.FeeRate,
						TreasuryAccountIndex: nPoolInfo.TreasuryAccountIndex,
						TreasuryRate:         nPoolInfo.TreasuryRate,
					}

					// update account state tree
					nLiquidityAssetLeaf, err := tree.ComputeLiquidityAssetLeafHash(
						liquidityMap[mempoolTxDetail.AssetId].AssetAId,
						liquidityMap[mempoolTxDetail.AssetId].AssetA,
						liquidityMap[mempoolTxDetail.AssetId].AssetBId,
						liquidityMap[mempoolTxDetail.AssetId].AssetB,
						liquidityMap[mempoolTxDetail.AssetId].LpAmount,
						liquidityMap[mempoolTxDetail.AssetId].KLast,
						liquidityMap[mempoolTxDetail.AssetId].FeeRate,
						liquidityMap[mempoolTxDetail.AssetId].TreasuryAccountIndex,
						liquidityMap[mempoolTxDetail.AssetId].TreasuryRate,
					)
					if err != nil {
						log.Println("[CommitterTask] unable to compute new account liquidity leaf:", err)
						return err
					}
					err = liquidityTree.Update(mempoolTxDetail.AssetId, nLiquidityAssetLeaf)
					if err != nil {
						log.Println("[CommitterTask] unable to update liquidity tree:", err)
						return err
					}

					break
				case NftAssetType:
					// check if nft exists in the db
					if nftMap[mempoolTxDetail.AssetId] == nil {
						nftMap[mempoolTxDetail.AssetId], err = ctx.L2NftModel.GetNftAsset(mempoolTxDetail.AssetId)
						if err != nil {
							logx.Errorf("[CommitterTask] unable to get nft asset: %s", err.Error())
							return err
						}
					}
					// check special type
					if mempoolTx.TxType == commonTx.TxTypeDepositNft || mempoolTx.TxType == commonTx.TxTypeMintNft {
						pendingNewNftIndexMap[mempoolTxDetail.AssetId] = true
						baseBalance = commonAsset.EmptyNftInfo(nftMap[mempoolTxDetail.AssetId].NftIndex).String()
					} else {
						pendingNewNftIndexMap[mempoolTxDetail.AssetId] = true
						pendingUpdateNftIndexMap[mempoolTxDetail.AssetId] = true
						// before nft info
						baseBalance = commonAsset.ConstructNftInfo(
							nftMap[mempoolTxDetail.AssetId].NftIndex,
							nftMap[mempoolTxDetail.AssetId].CreatorAccountIndex,
							nftMap[mempoolTxDetail.AssetId].OwnerAccountIndex,
							nftMap[mempoolTxDetail.AssetId].NftContentHash,
							nftMap[mempoolTxDetail.AssetId].NftL1TokenId,
							nftMap[mempoolTxDetail.AssetId].NftL1Address,
							nftMap[mempoolTxDetail.AssetId].CreatorTreasuryRate,
							nftMap[mempoolTxDetail.AssetId].CollectionId,
						).String()
					}
					if mempoolTx.TxType == commonTx.TxTypeWithdrawNft || mempoolTx.TxType == commonTx.TxTypeFullExitNft {
						pendingNewNftWithdrawHistory = append(pendingNewNftWithdrawHistory, &nft.L2NftWithdrawHistory{
							NftIndex:            nftMap[mempoolTxDetail.AssetId].NftIndex,
							CreatorAccountIndex: nftMap[mempoolTxDetail.AssetId].CreatorAccountIndex,
							OwnerAccountIndex:   nftMap[mempoolTxDetail.AssetId].OwnerAccountIndex,
							NftContentHash:      nftMap[mempoolTxDetail.AssetId].NftContentHash,
							NftL1Address:        nftMap[mempoolTxDetail.AssetId].NftL1Address,
							NftL1TokenId:        nftMap[mempoolTxDetail.AssetId].NftL1TokenId,
							CreatorTreasuryRate: nftMap[mempoolTxDetail.AssetId].CreatorTreasuryRate,
							CollectionId:        nftMap[mempoolTxDetail.AssetId].CollectionId,
						})
					}
					// delta nft info
					nftInfo, err := commonAsset.ParseNftInfo(mempoolTxDetail.BalanceDelta)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to parse nft info: %s", err.Error())
						return err
					}
					if pendingUpdateNftIndexMap[mempoolTxDetail.AssetId] {
						// update nft info
						nftMap[mempoolTxDetail.AssetId] = &L2Nft{
							Model:               nftMap[mempoolTxDetail.AssetId].Model,
							NftIndex:            nftInfo.NftIndex,
							CreatorAccountIndex: nftInfo.CreatorAccountIndex,
							OwnerAccountIndex:   nftInfo.OwnerAccountIndex,
							NftContentHash:      nftInfo.NftContentHash,
							NftL1Address:        nftInfo.NftL1Address,
							NftL1TokenId:        nftInfo.NftL1TokenId,
							CreatorTreasuryRate: nftInfo.CreatorTreasuryRate,
							CollectionId:        nftInfo.CollectionId,
						}
					}
					// get nft asset
					nftAsset := nftMap[mempoolTxDetail.AssetId]
					// update nft tree
					nNftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
						nftAsset.CreatorAccountIndex, nftAsset.OwnerAccountIndex,
						nftAsset.NftContentHash,
						nftAsset.NftL1Address, nftAsset.NftL1TokenId,
						nftAsset.CreatorTreasuryRate,
						nftAsset.CollectionId,
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
				case CollectionNonceAssetType:
					baseBalance = strconv.FormatInt(accountMap[mempoolTxDetail.AccountIndex].CollectionNonce, 10)
					newCollectionNonce, err = strconv.ParseInt(mempoolTxDetail.BalanceDelta, 10, 64)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to parse int: %s", err.Error())
						return err
					}
					if newCollectionNonce != accountMap[mempoolTxDetail.AccountIndex].CollectionNonce+1 {
						logx.Errorf("[CommitterTask] invalid collection nonce")
						return errors.New("[CommitterTask] invalid collection nonce")
					}
					break
				default:
					logx.Error("[CommitterTask] invalid tx type")
					return errors.New("[CommitterTask] invalid tx type")
				}
				var (
					nonce, collectionNonce int64
				)
				if mempoolTxDetail.AccountIndex != commonConstant.NilTxAccountIndex {
					nonce = accountMap[mempoolTxDetail.AccountIndex].Nonce
					collectionNonce = accountMap[mempoolTxDetail.AccountIndex].CollectionNonce
				}
				txDetails = append(txDetails, &tx.TxDetail{
					AssetId:         mempoolTxDetail.AssetId,
					AssetType:       mempoolTxDetail.AssetType,
					AccountIndex:    mempoolTxDetail.AccountIndex,
					AccountName:     mempoolTxDetail.AccountName,
					Balance:         baseBalance,
					BalanceDelta:    mempoolTxDetail.BalanceDelta,
					Order:           mempoolTxDetail.Order,
					Nonce:           nonce,
					AccountOrder:    mempoolTxDetail.AccountOrder,
					CollectionNonce: collectionNonce,
				})
			}
			if mempoolTx.Nonce != commonConstant.NilNonce {
				// update nonce
				accountMap[mempoolTx.AccountIndex].Nonce = mempoolTx.Nonce
			}
			// check if we need to update nonce
			if newCollectionNonce != commonConstant.NilCollectionId {
				accountMap[mempoolTx.AccountIndex].CollectionNonce = newCollectionNonce
			}
			// update account tree
			for accountIndex, _ := range pendingUpdateAccountIndexMap {
				nAccountLeafHash, err := tree.ComputeAccountLeafHash(
					accountMap[accountIndex].AccountNameHash,
					accountMap[accountIndex].PublicKey,
					accountMap[accountIndex].Nonce,
					accountMap[accountIndex].CollectionNonce,
					(*accountAssetTrees)[accountIndex].RootNode.Value,
				)
				if err != nil {
					log.Println("[CommitterTask] unable to compute account leaf:", err)
					return err
				}
				err = accountTree.Update(accountIndex, nAccountLeafHash)
				if err != nil {
					log.Println("[CommitterTask] unable to update account tree:", err)
					return err
				}
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
			hFunc.Write(liquidityTree.RootNode.Value)
			hFunc.Write(nftTree.RootNode.Value)
			stateRoot := common.Bytes2Hex(hFunc.Sum(nil))
			finalStateRoot = stateRoot
			oTx := ConvertMempoolTxToTx(mempoolTx, txDetails, stateRoot, currentBlockHeight)
			txs = append(txs, oTx)
		}
		// construct assets history
		var (
			pendingUpdateAccounts      []*Account
			pendingNewAccountHistory   []*AccountHistory
			pendingUpdateLiquidity     []*Liquidity
			pendingNewLiquidityHistory []*LiquidityHistory
			pendingUpdateNft           []*L2Nft
			pendingNewNftHistory       []*L2NftHistory
		)
		// handle account
		for accountIndex, flag := range pendingUpdateAccountIndexMap {
			if !flag {
				continue
			}
			accountInfo, err := commonAsset.FromFormatAccountInfo(accountMap[accountIndex])
			if err != nil {
				logx.Errorf("[CommitterTask] unable to convert from format account info: %s", err.Error())
				return err
			}
			pendingUpdateAccounts = append(pendingUpdateAccounts, accountInfo)
			pendingNewAccountHistory = append(pendingNewAccountHistory, &AccountHistory{
				AccountIndex:    accountInfo.AccountIndex,
				Nonce:           accountInfo.Nonce,
				CollectionNonce: accountInfo.CollectionNonce,
				AssetInfo:       accountInfo.AssetInfo,
				AssetRoot:       accountInfo.AssetRoot,
				L2BlockHeight:   currentBlockHeight,
			})
		}
		for pairIndex, flag := range pendingUpdateLiquidityIndexMap {
			if !flag {
				continue
			}
			pendingUpdateLiquidity = append(pendingUpdateLiquidity, liquidityMap[pairIndex])
			pendingNewLiquidityHistory = append(pendingNewLiquidityHistory, &LiquidityHistory{
				PairIndex:            liquidityMap[pairIndex].PairIndex,
				AssetAId:             liquidityMap[pairIndex].AssetAId,
				AssetA:               liquidityMap[pairIndex].AssetA,
				AssetBId:             liquidityMap[pairIndex].AssetBId,
				AssetB:               liquidityMap[pairIndex].AssetB,
				LpAmount:             liquidityMap[pairIndex].LpAmount,
				KLast:                liquidityMap[pairIndex].KLast,
				FeeRate:              liquidityMap[pairIndex].FeeRate,
				TreasuryAccountIndex: liquidityMap[pairIndex].TreasuryAccountIndex,
				TreasuryRate:         liquidityMap[pairIndex].TreasuryRate,
				L2BlockHeight:        currentBlockHeight,
			})
		}
		for nftIndex, flag := range pendingNewNftIndexMap {
			if !flag {
				continue
			}
			pendingNewNftHistory = append(pendingNewNftHistory, &L2NftHistory{
				NftIndex:            nftMap[nftIndex].NftIndex,
				CreatorAccountIndex: nftMap[nftIndex].CreatorAccountIndex,
				OwnerAccountIndex:   nftMap[nftIndex].OwnerAccountIndex,
				NftContentHash:      nftMap[nftIndex].NftContentHash,
				NftL1Address:        nftMap[nftIndex].NftL1Address,
				NftL1TokenId:        nftMap[nftIndex].NftL1TokenId,
				CreatorTreasuryRate: nftMap[nftIndex].CreatorTreasuryRate,
				CollectionId:        nftMap[nftIndex].CollectionId,
				L2BlockHeight:       currentBlockHeight,
			})
		}
		for nftIndex, flag := range pendingUpdateNftIndexMap {
			if !flag {
				continue
			}
			pendingUpdateNft = append(pendingUpdateNft, nftMap[nftIndex])
		}
		var realTxsAmountPerBlock int
		for i := 0; i < len(TxsAmountPerBlock); i++ {
			if len(txs) <= TxsAmountPerBlock[i] {
				realTxsAmountPerBlock = TxsAmountPerBlock[i]
				logx.Infof("the block contains %d txs: %d real txs, %d empty txs", realTxsAmountPerBlock, len(txs), realTxsAmountPerBlock-len(txs))
				break
			}
		}
		emptyPubdata := make([]byte, (realTxsAmountPerBlock-len(txs))*32*std.PubDataSizePerTx)
		pubData = append(pubData, emptyPubdata...)
		// create commitment
		commitment := util.CreateBlockCommitment(
			currentBlockHeight,
			createdAt,
			common.FromHex(oldStateRoot),
			common.FromHex(finalStateRoot),
			pubData,
			int64(len(pubDataOffset)),
		)
		// update old state root
		oldStateRoot = finalStateRoot
		// construct block
		createAtTime := time.UnixMilli(createdAt)
		var (
			oBlock          *block.Block
			oBlockForCommit *blockForCommit.BlockForCommit
		)
		if len(txs) != 0 {
			oBlock = &Block{
				Model: gorm.Model{
					CreatedAt: createAtTime,
				},
				BlockSize:                    uint16(realTxsAmountPerBlock),
				BlockCommitment:              commitment,
				BlockHeight:                  currentBlockHeight,
				StateRoot:                    finalStateRoot,
				PriorityOperations:           priorityOperations,
				PendingOnChainOperationsHash: common.Bytes2Hex(pendingOnChainOperationsHash),
				Txs:                          txs,
				BlockStatus:                  block.StatusPending,
			}
			if pendingOnChainOperationsPubData != nil {
				onChainOperationsPubDataBytes, err := json.Marshal(pendingOnChainOperationsPubData)
				if err != nil {
					logx.Errorf("[CommitterTask] unable to marshal on chain operations pub data: %s", err.Error())
					return err
				}
				oBlock.PendingOnChainOperationsPubData = string(onChainOperationsPubDataBytes)
			}
			offsetBytes, err := json.Marshal(pubDataOffset)
			if err != nil {
				logx.Errorf("[CommitterTask] unable to marshal pub data: %s", err.Error())
				return err
			}
			oBlockForCommit = &BlockForCommit{
				BlockSize:         uint16(realTxsAmountPerBlock),
				BlockHeight:       currentBlockHeight,
				StateRoot:         finalStateRoot,
				PublicData:        common.Bytes2Hex(pubData),
				Timestamp:         createdAt,
				PublicDataOffsets: string(offsetBytes),
			}
		}

		// create block for committer
		// create block, history, update mempool txs, create new l1 amount infos
		err = ctx.BlockModel.CreateBlockForCommitter(
			oBlock,
			oBlockForCommit,
			pendingMempoolTxs,
			pendingDeleteMempoolTxs,
			pendingUpdateAccounts,
			pendingNewAccountHistory,
			pendingUpdateLiquidity,
			pendingNewLiquidityHistory,
			pendingUpdateNft,
			pendingNewNftHistory,
			pendingNewNftWithdrawHistory,
		)
		if err != nil {
			logx.Errorf("[CommitterTask] unable to create block for committer: %s", err.Error())
			return err
		}
		*lastCommitTimeStamp = time.Now()
	}
	return nil
}
