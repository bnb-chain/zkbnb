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
	"encoding/json"
	"errors"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-crypto/accumulators/merkleTree"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/tree"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/committer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/errcode"
	"github.com/zeromicro/go-zero/core/logx"
)

type paramInfo struct {
	pendingUpdateAccountIndexMap    map[int64]bool
	pendingUpdateLiquidityIndexMap  map[int64]bool
	pendingUpdateNftIndexMap        map[int64]bool
	pendingNewNftIndexMap           map[int64]bool
	pendingNewNftWithdrawHistory    []*nft.L2NftWithdrawHistory
	txs                             []*Tx
	finalStateRoot                  string
	pubData                         []byte
	priorityOperations              int64
	pubDataOffset                   []uint32
	pendingOnChainOperationsPubData [][]byte
	pendingOnChainOperationsHash    []byte
	pendingMempoolTxs               []*MempoolTx
	pendingDeleteMempoolTxs         []*MempoolTx
}

func CommitterTask(ctx *svc.ServiceContext,
	lastCommitTimeStamp time.Time, accountTree *tree.Tree, liquidityTree *tree.Tree, nftTree *tree.Tree, accountAssetTrees *[]*tree.Tree) error {
	mempoolTxs, err := ctx.MempoolModel.GetMempoolTxsListForCommitter()
	if err != nil {
		if err == errcode.ErrNotFound {
			return nil
		}
		logx.Errorf("[GetMempoolTxsListForCommitter] unable to get tx in mempool")
		return err
	}
	var txsLength = len(mempoolTxs)
	logx.Infof("[CommitterTask] Mempool txs number : %d", txsLength)
	currentBlockHeight, err := ctx.BlockModel.GetCurrentBlockHeight()
	if err != nil && err != errcode.ErrNotFound {
		logx.Errorf("[GetCurrentBlockHeight] err when get current block height err:%v", err)
		return err
	}
	lastBlock, err := ctx.BlockModel.GetBlockByBlockHeight(currentBlockHeight)
	if err != nil {
		logx.Errorf("[GetBlockByBlockHeight] unable to get block by height: %v", err)
		return err
	}
	blocksSize := int(math.Ceil(float64(txsLength) / float64(MaxTxsAmountPerBlock)))
	var (
		accountMap   = make(map[int64]*FormatAccountInfo)
		liquidityMap = make(map[int64]*Liquidity)
		nftMap       = make(map[int64]*L2Nft)
		oldStateRoot = lastBlock.StateRoot
	)
	for i := 0; i < blocksSize; i++ {
		var now = time.Now()
		if now.Unix()-lastCommitTimeStamp.Unix() < MaxCommitterInterval {
			if txsLength-i*MaxTxsAmountPerBlock < MaxTxsAmountPerBlock {
				logx.Infof("[CommitterTask] not enough transactions")
				return errcode.ErrNotEnoughTransactions
			}
		}
		createdAt := time.Now().UnixMilli()
		currentBlockHeight += 1
		startIndex := i * MaxTxsAmountPerBlock
		endIndex := (i + 1) * MaxTxsAmountPerBlock
		if endIndex > txsLength {
			endIndex = txsLength
		}
		param, err := getCommitTxs(ctx, accountAssetTrees, accountTree, nftTree, liquidityTree,
			currentBlockHeight, mempoolTxs[startIndex:endIndex], accountMap, liquidityMap, nftMap, createdAt)
		if err != nil {
			logx.Errorf("[getCommitTxs] err: %v", err)
			return err
		}
		if len(param.txs) == 0 {
			return nil
		}
		pendingUpdateAccounts, pendingNewAccountHistory, err := newAccount(param.pendingUpdateAccountIndexMap, accountMap, currentBlockHeight)
		if err != nil {
			logx.Errorf("[newAccount] err: %v", err)
			return err
		}
		pendingUpdateLiquidity, pendingNewLiquidityHistory := newLiquidity(param.pendingUpdateLiquidityIndexMap, liquidityMap, currentBlockHeight)
		pendingNewNftHistory := newPendingNewNftHistory(param.pendingNewNftIndexMap, nftMap, currentBlockHeight)
		var pendingUpdateNft []*L2Nft
		for nftIndex, flag := range param.pendingUpdateNftIndexMap {
			if !flag {
				continue
			}
			pendingUpdateNft = append(pendingUpdateNft, nftMap[nftIndex])
		}
		// create commitment
		commitment := util.CreateBlockCommitment(currentBlockHeight, createdAt,
			common.FromHex(oldStateRoot), common.FromHex(param.finalStateRoot), param.pubData, int64(len(param.pubDataOffset)))
		// update old state root
		oldStateRoot = param.finalStateRoot
		oBlock, err := newBlock(time.UnixMilli(createdAt), commitment, param.finalStateRoot, param.txs, param.pendingOnChainOperationsPubData,
			currentBlockHeight, param.priorityOperations, param.pendingOnChainOperationsHash)
		if err != nil {
			logx.Errorf("[newBlock] unable to marshal pub data: %v", err)
			return err
		}
		oBlockForCommit, err := newBlockForCommit(currentBlockHeight, param.finalStateRoot, param.pubData, createdAt, param.pubDataOffset)
		if err != nil {
			logx.Errorf("[newBlockForCommit] unable to marshal pub data: %v", err)
			return err
		}
		if err = ctx.BlockModel.CreateBlockForCommitter(oBlock, oBlockForCommit,
			param.pendingMempoolTxs, param.pendingDeleteMempoolTxs, pendingUpdateAccounts, pendingNewAccountHistory, pendingUpdateLiquidity,
			pendingNewLiquidityHistory, pendingUpdateNft, pendingNewNftHistory, param.pendingNewNftWithdrawHistory); err != nil {
			logx.Errorf("[CommitterTask] unable to create block for committer: %v", err)
			return err
		}
	}
	return nil
}

func getCommitTxs(ctx *svc.ServiceContext, accountAssetTrees *[]*tree.Tree, accountTree *tree.Tree, nftTree *merkleTree.Tree, liquidityTree *tree.Tree,
	currentBlockHeight int64, mempoolTxs []*mempool.MempoolTx, accountMap map[int64]*FormatAccountInfo, liquidityMap map[int64]*liquidity.Liquidity,
	nftMap map[int64]*nft.L2Nft, createdAt int64) (*paramInfo, error) {
	var err error
	param := &paramInfo{
		pendingUpdateAccountIndexMap:    make(map[int64]bool),
		pendingUpdateLiquidityIndexMap:  make(map[int64]bool),
		pendingUpdateNftIndexMap:        make(map[int64]bool),
		pendingNewNftIndexMap:           make(map[int64]bool),
		pendingNewNftWithdrawHistory:    make([]*nft.L2NftWithdrawHistory, 0),
		txs:                             make([]*Tx, 0),
		finalStateRoot:                  "",
		pubData:                         make([]byte, 0),
		priorityOperations:              0,
		pubDataOffset:                   make([]uint32, 0),
		pendingOnChainOperationsPubData: make([][]byte, 0),
		pendingOnChainOperationsHash:    make([]byte, 0),
		pendingMempoolTxs:               make([]*MempoolTx, 0),
		pendingDeleteMempoolTxs:         make([]*MempoolTx, 0),
	}
	param.pendingOnChainOperationsHash = common.FromHex(util.EmptyStringKeccak)
	for _, mempoolTx := range mempoolTxs {
		var (
			pendingPriorityOperation int64
			newCollectionNonce       = commonConstant.NilCollectionId
		)
		pendingPriorityOperation, param.pendingOnChainOperationsPubData, param.pendingOnChainOperationsHash, param.pubData, param.pubDataOffset, err =
			handleTxPubData(mempoolTx, param.pubData, param.pendingOnChainOperationsPubData, param.pendingOnChainOperationsHash, param.pubDataOffset)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to handle l1 tx: %v", err)
			return param, err
		}
		param.priorityOperations += pendingPriorityOperation
		// get related account info
		if mempoolTx.AccountIndex != commonConstant.NilTxAccountIndex {
			if accountMap[mempoolTx.AccountIndex] == nil {
				accountInfo, err := ctx.AccountModel.GetAccountByAccountIndex(mempoolTx.AccountIndex)
				if err != nil {
					logx.Errorf("[CommitterTask] get account by account index: %v", err)
					return param, err
				}
				accountMap[mempoolTx.AccountIndex], err = commonAsset.ToFormatAccountInfo(accountInfo)
				if err != nil {
					logx.Errorf("[CommitterTask] unable to format account info: %v", err)
					return param, err
				}
			}
			// handle registerZNS tx
			param.pendingUpdateAccountIndexMap[mempoolTx.AccountIndex] = true
			if accountMap[mempoolTx.AccountIndex].Status == account.AccountStatusPending {
				if mempoolTx.TxType != TxTypeRegisterZns {
					logx.Errorf("[CommitterTask] first transaction should be registerZNS")
					return param, errors.New("[CommitterTask] first transaction should be registerZNS")
				}
				accountMap[mempoolTx.AccountIndex].Status = account.AccountStatusConfirmed
				param.pendingUpdateAccountIndexMap[mempoolTx.AccountIndex] = true
				// update account tree
				if int64(len(*accountAssetTrees)) != mempoolTx.AccountIndex {
					logx.Errorf("[CommitterTask] invalid account index")
					return param, errors.New("[CommitterTask] invalid account index")
				}
				emptyAssetTree, err := tree.NewEmptyAccountAssetTree()
				if err != nil {
					logx.Errorf("[CommitterTask] unable to new empty account state tree")
					return param, err
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
					return param, err
				}
				err = accountTree.Update(mempoolTx.AccountIndex, nAccountLeafHash)
				if err != nil {
					log.Println("[CommitterTask] unable to update account tree:", err)
					return param, err
				}
			}
		}
		// check if the tx is still valid
		if mempoolTx.ExpiredAt != commonConstant.NilExpiredAt {
			if mempoolTx.ExpiredAt < createdAt {
				mempoolTx.Status = mempool.FailTxStatus
				mempoolTx.L2BlockHeight = currentBlockHeight
				param.pendingDeleteMempoolTxs = append(param.pendingDeleteMempoolTxs, mempoolTx)
				continue
			}
		}
		if mempoolTx.Nonce != commonConstant.NilNonce {
			// check nonce, the latest nonce should be previous nonce + 1
			if mempoolTx.Nonce != accountMap[mempoolTx.AccountIndex].Nonce+1 {
				mempoolTx.Status = mempool.FailTxStatus
				mempoolTx.L2BlockHeight = currentBlockHeight
				param.pendingDeleteMempoolTxs = append(param.pendingDeleteMempoolTxs, mempoolTx)
				continue
			}
		}
		// check mempool tx details are correct
		var txDetails []*tx.TxDetail
		for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
			if mempoolTxDetail.AccountIndex != commonConstant.NilTxAccountIndex {
				param.pendingUpdateAccountIndexMap[mempoolTxDetail.AccountIndex] = true
				if accountMap[mempoolTxDetail.AccountIndex] == nil {
					accountInfo, err := ctx.AccountModel.GetAccountByAccountIndex(mempoolTxDetail.AccountIndex)
					if err != nil {
						logx.Errorf("[CommitterTask] get account by account index: %s", err.Error())
						return param, err
					}
					accountMap[mempoolTxDetail.AccountIndex], err = commonAsset.ToFormatAccountInfo(accountInfo)
					if err != nil {
						logx.Errorf("[CommitterTask] unable to format account info: %s", err.Error())
						return param, err
					}
				}
			}
			var baseBalance string
			// check balance
			switch mempoolTxDetail.AssetType {
			case GeneralAssetType:
				param.pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, err = processGeneralAssetType(ctx,
					accountMap, param.pendingDeleteMempoolTxs, mempoolTx, accountAssetTrees, mempoolTxDetail, currentBlockHeight)
				if err != nil {
					if err.Error() == errcode.ErrContinueFlag.Error() {
						continue
					}
					logx.Errorf("[processGeneralAssetType] err: %v", err)
					return param, err
				}
			case LiquidityAssetType:
				param.pendingUpdateLiquidityIndexMap, liquidityMap, baseBalance, err = processLiquidityAssetType(ctx,
					param.pendingUpdateLiquidityIndexMap, mempoolTx, mempoolTxDetail, liquidityMap, liquidityTree)
				if err != nil {
					logx.Errorf("[processLiquidityAssetType] err: %v", err)
					return param, err
				}
			case NftAssetType:
				param.pendingNewNftWithdrawHistory, baseBalance, param.pendingNewNftIndexMap, param.pendingUpdateNftIndexMap, err = processNftAssetType(ctx,
					nftMap, nftTree, mempoolTxDetail, mempoolTx, param.pendingNewNftIndexMap, param.pendingUpdateNftIndexMap, param.pendingNewNftWithdrawHistory)
				if err != nil {
					logx.Errorf("[processNftAssetType] err: %v", err)
					return param, err
				}
			case CollectionNonceAssetType:
				baseBalance, newCollectionNonce, err = processCollectionNonceAssetType(accountMap, mempoolTxDetail)
				if err != nil {
					logx.Errorf("[processCollectionNonceAssetType] err: %v", err)
					return param, err
				}
			default:
				logx.Error("[CommitterTask] invalid tx type")
				return param, errors.New("[CommitterTask] invalid tx type")
			}
			tmpTx := &tx.TxDetail{
				AssetId:      mempoolTxDetail.AssetId,
				AssetType:    mempoolTxDetail.AssetType,
				AccountIndex: mempoolTxDetail.AccountIndex,
				AccountName:  mempoolTxDetail.AccountName,
				Balance:      baseBalance,
				BalanceDelta: mempoolTxDetail.BalanceDelta,
				Order:        mempoolTxDetail.Order,
				AccountOrder: mempoolTxDetail.AccountOrder,
			}
			if mempoolTxDetail.AccountIndex != commonConstant.NilTxAccountIndex {
				tmpTx.Nonce = accountMap[mempoolTxDetail.AccountIndex].Nonce
				tmpTx.CollectionNonce = accountMap[mempoolTxDetail.AccountIndex].CollectionNonce
			}
			txDetails = append(txDetails, tmpTx)
		}
		if mempoolTx.Nonce != commonConstant.NilNonce {
			accountMap[mempoolTx.AccountIndex].Nonce = mempoolTx.Nonce
		}
		if newCollectionNonce != commonConstant.NilCollectionId {
			accountMap[mempoolTx.AccountIndex].CollectionNonce = newCollectionNonce
		}
		// update account tree
		for accountIndex, _ := range param.pendingUpdateAccountIndexMap {
			nAccountLeafHash, err := tree.ComputeAccountLeafHash(accountMap[accountIndex].AccountNameHash, accountMap[accountIndex].PublicKey,
				accountMap[accountIndex].Nonce, accountMap[accountIndex].CollectionNonce, (*accountAssetTrees)[accountIndex].RootNode.Value)
			if err != nil {
				logx.Errorf("[CommitterTask] unable to compute account leaf:", err)
				return param, err
			}
			if err = accountTree.Update(accountIndex, nAccountLeafHash); err != nil {
				logx.Errorf("[accountTree.Update] err:", err)
				return param, err
			}
		}
		// update mempool tx info
		mempoolTx.L2BlockHeight = currentBlockHeight
		mempoolTx.Status = mempool.SuccessTxStatus
		param.pendingMempoolTxs = append(param.pendingMempoolTxs, mempoolTx)
		param.finalStateRoot = newStateRoot(accountTree, liquidityTree, nftTree)
		param.txs = append(param.txs, ConvertMempoolTxToTx(mempoolTx, txDetails, param.finalStateRoot, currentBlockHeight))
	}
	return param, err

}

func processGeneralAssetType(ctx *svc.ServiceContext, accountMap map[int64]*commonAsset.AccountInfo, pendingDeleteMempoolTxs []*mempool.MempoolTx,
	mempoolTx *mempool.MempoolTx, accountAssetTrees *[]*merkleTree.Tree, mempoolTxDetail *mempool.MempoolTxDetail,
	currentBlockHeight int64) ([]*mempool.MempoolTx, *mempool.MempoolTx, *mempool.MempoolTxDetail, map[int64]*commonAsset.AccountInfo, string, error) {
	var err error
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
	baseBalance := accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId].String()
	var nBalance string
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
			return pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, err
		}
		mempoolTxDetail.BalanceDelta = balanceDelta.String()
		txInfo, err := commonTx.ParseFullExitTxInfo(mempoolTx.TxInfo)
		if err != nil {
			logx.Errorf("[CommitterTask] unable to parse full exit tx info: %s", err.Error())
			return pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, err
		}
		txInfo.AssetAmount = accountMap[mempoolTxDetail.AccountIndex].AssetInfo[mempoolTxDetail.AssetId].Balance
		infoBytes, err := json.Marshal(txInfo)
		if err != nil {
			logx.Errorf("[CommitterTask] unable to marshal tx: %s", err.Error())
			return pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, err
		}
		mempoolTx.TxInfo = string(infoBytes)
	} else {
		// compute new balance
		nBalance, err = commonAsset.ComputeNewBalance(GeneralAssetType, baseBalance, mempoolTxDetail.BalanceDelta)
		if err != nil {
			logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
			return pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, err
		}
	}
	nAccountAsset, err := commonAsset.ParseAccountAsset(nBalance)
	if err != nil {
		logx.Errorf("[CommitterTask] unable to parse account asset: %s", err.Error())
		return pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, err
	}
	// check balance is valid
	if nAccountAsset.Balance.Cmp(util.ZeroBigInt) < 0 {
		// mark this transaction as invalid transaction
		mempoolTx.Status = mempool.FailTxStatus
		mempoolTx.L2BlockHeight = currentBlockHeight
		pendingDeleteMempoolTxs = append(pendingDeleteMempoolTxs, mempoolTx)
		// continue
		return pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, errcode.ErrContinueFlag
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
		return pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, err
	}
	err = (*accountAssetTrees)[mempoolTxDetail.AccountIndex].Update(mempoolTxDetail.AssetId, nAssetLeaf)
	if err != nil {
		log.Println("[CommitterTask] unable to update asset tree:", err)
		return pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, err
	}
	accountMap[mempoolTxDetail.AccountIndex].AssetRoot = common.Bytes2Hex(
		(*accountAssetTrees)[mempoolTxDetail.AccountIndex].RootNode.Value)
	return pendingDeleteMempoolTxs, mempoolTx, mempoolTxDetail, accountMap, baseBalance, nil
}

func processLiquidityAssetType(ctx *svc.ServiceContext, pendingUpdateLiquidityIndexMap map[int64]bool, mempoolTx *mempool.MempoolTx,
	mempoolTxDetail *mempool.MempoolTxDetail, liquidityMap map[int64]*liquidity.Liquidity, liquidityTree *tree.Tree) (map[int64]bool, map[int64]*liquidity.Liquidity, string, error) {
	var err error
	pendingUpdateLiquidityIndexMap[mempoolTxDetail.AssetId] = true
	if liquidityMap[mempoolTxDetail.AssetId] == nil {
		liquidityMap[mempoolTxDetail.AssetId], err = ctx.LiquidityModel.GetLiquidityByPairIndex(mempoolTxDetail.AssetId)
		if err != nil {
			logx.Errorf("[CommitterTask] unable to get latest liquidity by pair index: %s", err.Error())
			return pendingUpdateLiquidityIndexMap, liquidityMap, "", err
		}
	}
	var poolInfo *commonAsset.LiquidityInfo
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
			return pendingUpdateLiquidityIndexMap, liquidityMap, "", err
		}
	}
	baseBalance := poolInfo.String()
	// compute new balance
	nBalance, err := commonAsset.ComputeNewBalance(
		LiquidityAssetType, baseBalance, mempoolTxDetail.BalanceDelta)
	if err != nil {
		logx.Error("[CommitterTask] unable to compute new balance: %s", err.Error())
		return pendingUpdateLiquidityIndexMap, liquidityMap, "", err
	}
	nPoolInfo, err := commonAsset.ParseLiquidityInfo(nBalance)
	if err != nil {
		logx.Errorf("[CommitterTask] unable to parse pair info: %s", err.Error())
		return pendingUpdateLiquidityIndexMap, liquidityMap, baseBalance, err
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
		return pendingUpdateLiquidityIndexMap, liquidityMap, baseBalance, err
	}
	if err = liquidityTree.Update(mempoolTxDetail.AssetId, nLiquidityAssetLeaf); err != nil {
		log.Println("[CommitterTask] unable to update liquidity tree:", err)
		return pendingUpdateLiquidityIndexMap, liquidityMap, baseBalance, err
	}
	return pendingUpdateLiquidityIndexMap, liquidityMap, baseBalance, err

}

func processNftAssetType(ctx *svc.ServiceContext, nftMap map[int64]*nft.L2Nft, nftTree *merkleTree.Tree,
	mempoolTxDetail *mempool.MempoolTxDetail, mempoolTx *mempool.MempoolTx,
	pendingNewNftIndexMap map[int64]bool, pendingUpdateNftIndexMap map[int64]bool,
	pendingNewNftWithdrawHistory []*nft.L2NftWithdrawHistory) ([]*nft.L2NftWithdrawHistory, string, map[int64]bool, map[int64]bool, error) {
	// check if nft exists in the db
	var err error
	var baseBalance string
	if nftMap[mempoolTxDetail.AssetId] == nil {
		nftMap[mempoolTxDetail.AssetId], err = ctx.L2NftModel.GetNftAsset(mempoolTxDetail.AssetId)
		if err != nil {
			logx.Errorf("[L2NftModel.GetNftAsset] err: %v", err)
			return pendingNewNftWithdrawHistory, baseBalance, pendingNewNftIndexMap, pendingUpdateNftIndexMap, err
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
		return pendingNewNftWithdrawHistory, baseBalance, pendingNewNftIndexMap, pendingUpdateNftIndexMap, err
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
	nNftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
		nftAsset.CreatorAccountIndex, nftAsset.OwnerAccountIndex,
		nftAsset.NftContentHash,
		nftAsset.NftL1Address, nftAsset.NftL1TokenId,
		nftAsset.CreatorTreasuryRate,
		nftAsset.CollectionId,
	)
	if err != nil {
		logx.Errorf("[CommitterTask] unable to compute new nft asset leaf: %s", err)
		return pendingNewNftWithdrawHistory, baseBalance, pendingNewNftIndexMap, pendingUpdateNftIndexMap, err
	}
	if err = nftTree.Update(mempoolTxDetail.AssetId, nNftAssetLeaf); err != nil {
		log.Println("[CommitterTask] unable to update nft tree:", err)
		return pendingNewNftWithdrawHistory, baseBalance, pendingNewNftIndexMap, pendingUpdateNftIndexMap, err
	}
	return pendingNewNftWithdrawHistory, baseBalance, pendingNewNftIndexMap, pendingUpdateNftIndexMap, nil
}

func processCollectionNonceAssetType(accountMap map[int64]*commonAsset.AccountInfo, mempoolTxDetail *mempool.MempoolTxDetail) (string, int64, error) {
	baseBalance := strconv.FormatInt(accountMap[mempoolTxDetail.AccountIndex].CollectionNonce, 10)
	newCollectionNonce, err := strconv.ParseInt(mempoolTxDetail.BalanceDelta, 10, 64)
	if err != nil {
		logx.Errorf("[strconv.ParseInt] unable to parse int: %s", err.Error())
		return "", 0, err
	}
	if newCollectionNonce != accountMap[mempoolTxDetail.AccountIndex].CollectionNonce+1 {
		logx.Errorf("[processCollectionNonceAssetType] invalid collection nonce")
		return "", 0, errcode.ErrNotInvalidCollectionNonce
	}
	return baseBalance, newCollectionNonce, nil
}
