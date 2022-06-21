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
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/service/cronjob/mempoolMonitor/internal/svc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"math/big"
)

func MonitorMempool(
	ctx *svc.ServiceContext,
) error {
	// get oTx from l2txEventMonitor
	txs, err := ctx.L2TxEventMonitorModel.GetL2TxEventMonitorsByStatus(PendingStatus)
	if err != nil {
		if err == l2TxEventMonitor.ErrNotFound {
			logx.Info("[MonitorMempool] no l2 oTx event monitors")
			return err
		} else {
			logx.Error("[MonitorMempool] unable to get l2 oTx event monitors")
			return err
		}
	}
	// initialize mempool txs
	//nextAccountIndex, err := ctx.AccountModel.GetLatestAccountIndex()
	//if err != nil {
	//	if err == account.ErrNotFound {
	//		nextAccountIndex = -1
	//	} else {
	//		errInfo := fmt.Sprintf("[mempoolMoniter.MonitorMempool]<=>[accountModel.GetMaxAccountIndex] %s", err.Error())
	//		logx.Error(errInfo)
	//		return err
	//	}
	//}

	var (
		pendingNewAccounts       []*account.Account
		pendingNewMempoolTxs     []*mempool.MempoolTx
		pendingNewLiquidityInfos []*liquidity.Liquidity
		pendingNewNfts           []*nft.L2Nft
		newAccountInfoMap        = make(map[string]*account.Account)
		newNftInfoMap            = make(map[int64]*commonAsset.NftInfo)
		newLiquidityInfoMap      = make(map[int64]*liquidity.Liquidity)
		relatedAccountIndex      = make(map[int64]bool)
	)
	// get last handled request id
	currentRequestId, err := ctx.L2TxEventMonitorModel.GetLastHandledRequestId()
	if err != nil {
		logx.Errorf("[MonitorMempool] unable to get last handled request id: %s", err.Error())
		return err
	}
	for _, oTx := range txs {
		// set oTx as handled
		oTx.Status = l2TxEventMonitor.HandledStatus
		// request id must be in order
		if oTx.RequestId != currentRequestId+1 {
			logx.Errorf("[MonitorMempool] invalid request id")
			return errors.New("[MonitorMempool] invalid request id")
		}
		currentRequestId++

		// handle oTx based on oTx type
		switch oTx.TxType {
		case TxTypeRegisterZns:
			// parse oTx info
			txInfo, err := util.ParseRegisterZnsPubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse registerZNS pub data: %s", err.Error())
				return err
			}
			// check if the account name has been registered
			_, err = ctx.AccountModel.GetAccountByAccountName(txInfo.AccountName)
			if err != ErrNotFound {
				logx.Errorf("[MonitorMempool] account name has been registered")
				return errors.New("[MonitorMempool] account name has been registered")
			}
			// set correct account index
			//nextAccountIndex++
			//txInfo.AccountIndex = nextAccountIndex
			// create new account and account history
			accountInfo := &account.Account{
				AccountIndex:    txInfo.AccountIndex,
				AccountName:     txInfo.AccountName,
				PublicKey:       txInfo.PubKey,
				AccountNameHash: common.Bytes2Hex(txInfo.AccountNameHash),
				L1Address:       oTx.SenderAddress,
				Nonce:           commonConstant.NilNonce,
				CollectionNonce: commonConstant.NilNonce,
				AssetInfo:       commonConstant.NilAssetInfo,
				AssetRoot:       common.Bytes2Hex(tree.NilAccountAssetRoot),
				Status:          account.StatusPending,
			}
			pendingNewAccounts = append(pendingNewAccounts, accountInfo)
			accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
			newAccountInfoMap[accountNameHash] = accountInfo
			// create mempool oTx
			// serialize oTx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize oTx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:        RandomTxHash(),
				TxType:        int64(txInfo.TxType),
				GasFeeAssetId: commonConstant.NilAssetId,
				GasFee:        commonConstant.NilAssetAmountStr,
				NftIndex:      commonConstant.NilTxNftIndex,
				PairIndex:     commonConstant.NilPairIndex,
				AssetId:       commonConstant.NilAssetId,
				TxAmount:      commonConstant.NilAssetAmountStr,
				NativeAddress: oTx.SenderAddress,
				TxInfo:        string(txInfoBytes),
				AccountIndex:  txInfo.AccountIndex,
				Nonce:         commonConstant.NilNonce,
				ExpiredAt:     commonConstant.NilExpiredAt,
				L2BlockHeight: commonConstant.NilBlockHeight,
				Status:        mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			break
		case TxTypeCreatePair:
			// parse oTx info
			txInfo, err := util.ParseCreatePairPubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse registerZNS pub data: %s", err.Error())
				return err
			}
			// liquidity info
			liquidityInfo := &liquidity.Liquidity{
				PairIndex:            txInfo.PairIndex,
				AssetAId:             txInfo.AssetAId,
				AssetA:               ZeroBigIntString,
				AssetBId:             txInfo.AssetBId,
				AssetB:               ZeroBigIntString,
				LpAmount:             ZeroBigIntString,
				KLast:                ZeroBigIntString,
				TreasuryAccountIndex: txInfo.TreasuryAccountIndex,
				FeeRate:              txInfo.FeeRate,
				TreasuryRate:         txInfo.TreasuryRate,
			}
			newLiquidityInfoMap[txInfo.PairIndex] = liquidityInfo
			pendingNewLiquidityInfos = append(pendingNewLiquidityInfos, liquidityInfo)
			// tx detail
			poolInfo := &commonAsset.LiquidityInfo{
				PairIndex:            txInfo.PairIndex,
				AssetAId:             txInfo.AssetAId,
				AssetA:               big.NewInt(0),
				AssetBId:             txInfo.AssetBId,
				AssetB:               big.NewInt(0),
				LpAmount:             big.NewInt(0),
				KLast:                big.NewInt(0),
				FeeRate:              txInfo.FeeRate,
				TreasuryAccountIndex: txInfo.TreasuryAccountIndex,
				TreasuryRate:         txInfo.TreasuryRate,
			}
			txDetail := &mempool.MempoolTxDetail{
				AssetId:      txInfo.PairIndex,
				AssetType:    commonAsset.LiquidityAssetType,
				AccountIndex: commonConstant.NilTxAccountIndex,
				AccountName:  commonConstant.NilAccountName,
				BalanceDelta: poolInfo.String(),
				Order:        0,
				AccountOrder: commonConstant.NilAccountOrder,
			}
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize oTx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFeeAssetId:  commonConstant.NilAssetId,
				GasFee:         commonConstant.NilAssetAmountStr,
				NftIndex:       commonConstant.NilTxNftIndex,
				PairIndex:      txInfo.PairIndex,
				AssetId:        commonConstant.NilAssetId,
				TxAmount:       commonConstant.NilAssetAmountStr,
				NativeAddress:  commonConstant.NilL1Address,
				MempoolDetails: []*mempool.MempoolTxDetail{txDetail},
				TxInfo:         string(txInfoBytes),
				AccountIndex:   commonConstant.NilTxAccountIndex,
				Nonce:          commonConstant.NilNonce,
				ExpiredAt:      commonConstant.NilExpiredAt,
				L2BlockHeight:  commonConstant.NilBlockHeight,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			break
		case TxTypeUpdatePairRate:
			// create mempool oTx
			txInfo, err := util.ParseUpdatePairRatePubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse update pair rate pub data: %s", err.Error())
				return err
			}
			var liquidityInfo *liquidity.Liquidity
			if newLiquidityInfoMap[txInfo.PairIndex] != nil {
				liquidityInfo = newLiquidityInfoMap[txInfo.PairIndex]
			} else {
				liquidityInfo, err = ctx.LiquidityModel.GetLiquidityByPairIndex(txInfo.PairIndex)
				if err != nil {
					logx.Errorf("[MonitorMempool] unable to get liquidity by pair index: %s", err.Error())
					return err
				}
			}
			liquidityInfo.FeeRate = txInfo.FeeRate
			liquidityInfo.TreasuryAccountIndex = txInfo.TreasuryAccountIndex
			liquidityInfo.TreasuryRate = txInfo.TreasuryRate
			// construct mempool tx
			poolInfo, err := commonAsset.ConstructLiquidityInfo(
				liquidityInfo.PairIndex,
				liquidityInfo.AssetAId,
				liquidityInfo.AssetA,
				liquidityInfo.AssetBId,
				liquidityInfo.AssetB,
				liquidityInfo.LpAmount,
				liquidityInfo.KLast,
				liquidityInfo.FeeRate,
				liquidityInfo.TreasuryAccountIndex,
				liquidityInfo.TreasuryRate,
			)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to construct liquidity info: %s", err.Error())
				return err
			}
			txDetail := &mempool.MempoolTxDetail{
				AssetId:      txInfo.PairIndex,
				AssetType:    commonAsset.LiquidityAssetType,
				AccountIndex: commonConstant.NilTxAccountIndex,
				AccountName:  commonConstant.NilAccountName,
				BalanceDelta: poolInfo.String(),
				Order:        0,
				AccountOrder: commonConstant.NilAccountOrder,
			}
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize oTx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFeeAssetId:  commonConstant.NilAssetId,
				GasFee:         commonConstant.NilAssetAmountStr,
				NftIndex:       commonConstant.NilTxNftIndex,
				PairIndex:      liquidityInfo.PairIndex,
				AssetId:        commonConstant.NilAssetId,
				TxAmount:       commonConstant.NilAssetAmountStr,
				NativeAddress:  commonConstant.NilL1Address,
				MempoolDetails: []*mempool.MempoolTxDetail{txDetail},
				TxInfo:         string(txInfoBytes),
				AccountIndex:   commonConstant.NilTxAccountIndex,
				Nonce:          commonConstant.NilNonce,
				ExpiredAt:      commonConstant.NilExpiredAt,
				L2BlockHeight:  commonConstant.NilBlockHeight,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			break
		case TxTypeDeposit:
			var accountInfo *account.Account
			// create mempool oTx
			txInfo, err := util.ParseDepositPubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit pub data: %s", err.Error())
				return err
			}
			accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
			if newAccountInfoMap[accountNameHash] != nil {
				accountInfo = newAccountInfoMap[accountNameHash]
			} else {
				accountInfo, err = GetAccountInfoByAccountNameHash(accountNameHash, ctx.AccountModel)
				if err != nil {
					logx.Errorf("[MonitorMempool] unable to get account info: %s", err.Error())
					return err
				}
			}
			txInfo.AccountIndex = accountInfo.AccountIndex
			var (
				mempoolTxDetails []*mempool.MempoolTxDetail
			)
			balanceDelta := &commonAsset.AccountAsset{
				AssetId:                  txInfo.AssetId,
				Balance:                  txInfo.AssetAmount,
				LpAmount:                 big.NewInt(0),
				OfferCanceledOrFinalized: big.NewInt(0),
			}
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      txInfo.AssetId,
				AssetType:    commonAsset.GeneralAssetType,
				AccountIndex: txInfo.AccountIndex,
				AccountName:  accountInfo.AccountName,
				BalanceDelta: balanceDelta.String(),
				Order:        0,
				AccountOrder: 0,
			})
			// serialize oTx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize oTx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFeeAssetId:  commonConstant.NilAssetId,
				GasFee:         commonConstant.NilAssetAmountStr,
				NftIndex:       commonConstant.NilTxNftIndex,
				PairIndex:      commonConstant.NilPairIndex,
				AssetId:        txInfo.AssetId,
				TxAmount:       txInfo.AssetAmount.String(),
				NativeAddress:  oTx.SenderAddress,
				MempoolDetails: mempoolTxDetails,
				TxInfo:         string(txInfoBytes),
				AccountIndex:   accountInfo.AccountIndex,
				Nonce:          commonConstant.NilNonce,
				ExpiredAt:      commonConstant.NilExpiredAt,
				L2BlockHeight:  commonConstant.NilBlockHeight,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			if !relatedAccountIndex[accountInfo.AccountIndex] {
				relatedAccountIndex[accountInfo.AccountIndex] = true
			}
			break
		case TxTypeDepositNft:
			// create mempool oTx
			var accountInfo *account.Account
			txInfo, err := util.ParseDepositNftPubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit nft pub data: %s", err.Error())
				return err
			}
			accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
			if newAccountInfoMap[accountNameHash] != nil {
				accountInfo = newAccountInfoMap[accountNameHash]
			} else {
				accountInfo, err = GetAccountInfoByAccountNameHash(accountNameHash, ctx.AccountModel)
				if err != nil {
					logx.Errorf("[MonitorMempool] unable to get account info: %s", err.Error())
					return err
				}
			}
			// complete oTx info
			txInfo.AccountIndex = accountInfo.AccountIndex
			redisLock, nftIndex, err := globalmapHandler.GetLatestNftIndexForWrite(ctx.NftModel, ctx.RedisConnection)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to get latest nft index: %s", err.Error())
				return err
			}
			// TODO possible resource leak
			defer redisLock.Release()
			var (
				nftInfo *commonAsset.NftInfo
			)
			if txInfo.NftIndex == 0 && txInfo.CreatorAccountIndex == 0 && txInfo.CreatorTreasuryRate == 0 {
				txInfo.NftIndex = nftIndex
			}
			nftInfo = commonAsset.ConstructNftInfo(
				txInfo.NftIndex,
				txInfo.CreatorAccountIndex,
				accountInfo.AccountIndex,
				common.Bytes2Hex(txInfo.NftContentHash),
				txInfo.NftL1TokenId.String(),
				txInfo.NftL1Address,
				txInfo.CreatorTreasuryRate,
				txInfo.CollectionId,
			)
			newNftInfoMap[nftInfo.NftIndex] = nftInfo
			var (
				mempoolTxDetails []*mempool.MempoolTxDetail
			)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to construct nft info: %s", err.Error())
				return err
			}
			// user info
			accountOrder := int64(0)
			emptyDeltaAsset := &commonAsset.AccountAsset{
				AssetId:                  0,
				Balance:                  big.NewInt(0),
				LpAmount:                 big.NewInt(0),
				OfferCanceledOrFinalized: big.NewInt(0),
			}
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      0,
				AssetType:    commonAsset.GeneralAssetType,
				AccountIndex: txInfo.AccountIndex,
				AccountName:  accountInfo.AccountName,
				BalanceDelta: emptyDeltaAsset.String(),
				AccountOrder: accountOrder,
			})
			// nft info
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      txInfo.NftIndex,
				AssetType:    commonAsset.NftAssetType,
				AccountIndex: txInfo.AccountIndex,
				AccountName:  accountInfo.AccountName,
				BalanceDelta: nftInfo.String(),
				AccountOrder: commonConstant.NilAccountOrder,
			})
			// serialize oTx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize oTx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFee:         commonConstant.NilAssetAmountStr,
				GasFeeAssetId:  commonConstant.NilAssetId,
				NftIndex:       nftIndex,
				PairIndex:      commonConstant.NilPairIndex,
				AssetId:        commonConstant.NilAssetId,
				TxAmount:       commonConstant.NilAssetAmountStr,
				NativeAddress:  oTx.SenderAddress,
				MempoolDetails: mempoolTxDetails,
				TxInfo:         string(txInfoBytes),
				AccountIndex:   accountInfo.AccountIndex,
				Nonce:          commonConstant.NilNonce,
				L2BlockHeight:  commonConstant.NilBlockHeight,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			if !relatedAccountIndex[accountInfo.AccountIndex] {
				relatedAccountIndex[accountInfo.AccountIndex] = true
			}
			// put into new nfts
			pendingNewNfts = append(pendingNewNfts, &nft.L2Nft{
				NftIndex:            nftInfo.NftIndex,
				CreatorAccountIndex: nftInfo.CreatorAccountIndex,
				OwnerAccountIndex:   nftInfo.OwnerAccountIndex,
				NftContentHash:      nftInfo.NftContentHash,
				NftL1Address:        nftInfo.NftL1Address,
				NftL1TokenId:        nftInfo.NftL1TokenId,
				CreatorTreasuryRate: nftInfo.CreatorTreasuryRate,
				CollectionId:        nftInfo.CollectionId,
			})
			break
		case TxTypeFullExit:
			// create mempool oTx
			var (
				accountInfo *commonAsset.AccountInfo
				redisLock   *redis.RedisLock
			)
			txInfo, err := util.ParseFullExitPubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit pub data: %s", err.Error())
				return err
			}
			accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
			if newAccountInfoMap[accountNameHash] != nil {
				accountInfo, err = commonAsset.ToFormatAccountInfo(newAccountInfoMap[accountNameHash])
				if err != nil {
					logx.Errorf("[MonitorMempool] unable convert to format account info: %s", err.Error())
					return err
				}
				for _, mempoolTx := range pendingNewMempoolTxs {
					if mempoolTx.AccountIndex != accountInfo.AccountIndex {
						continue
					}
					for _, txDetail := range mempoolTx.MempoolDetails {
						if txDetail.AccountIndex != accountInfo.AccountIndex || txDetail.AssetId != txInfo.AssetId {
							continue
						}
						if txDetail.AssetType == GeneralAssetType {
							if accountInfo.AssetInfo[txDetail.AssetId] == nil {
								accountInfo.AssetInfo[txDetail.AssetId] = &commonAsset.AccountAsset{
									AssetId:                  txDetail.AssetId,
									Balance:                  big.NewInt(0),
									LpAmount:                 big.NewInt(0),
									OfferCanceledOrFinalized: big.NewInt(0),
								}
							}
							nBalance, err := commonAsset.ComputeNewBalance(GeneralAssetType, accountInfo.AssetInfo[txDetail.AssetId].String(), txDetail.BalanceDelta)
							if err != nil {
								logx.Errorf("[MonitorMempool] unable to compute new balance: %s", err.Error())
								return err
							}
							accountInfo.AssetInfo[txDetail.AssetId], err = commonAsset.ParseAccountAsset(nBalance)
							if err != nil {
								logx.Errorf("[MonitorMempool] unable to parse account asset : %s", err.Error())
								return err
							}
						}
					}
				}
			} else {
				newAccountInfoMap[accountNameHash], err = GetAccountInfoByAccountNameHash(accountNameHash, ctx.AccountModel)
				if err != nil {
					logx.Errorf("[MonitorMempool] unable to get account info: %s", err.Error())
					return err
				}
				accountInfo, err = commonAsset.ToFormatAccountInfo(newAccountInfoMap[accountNameHash])
				if err != nil {
					logx.Errorf("[MonitorMempool] unable convert to format account info: %s", err.Error())
					return err
				}
				key := util.GetAccountKey(accountInfo.AccountIndex)
				lockKey := util.GetLockKey(key)
				redisLock = redis.NewRedisLock(ctx.RedisConnection, lockKey)
				redisLock.SetExpire(5)
				isAcquired, err := redisLock.Acquire()
				if err != nil {
					logx.Errorf("[MonitorMempool] unable to acquire the lock: %s", err.Error())
					return err
				}
				if !isAcquired {
					logx.Errorf("[MonitorMempool] unable to acquire the lock")
					return errors.New("[MonitorMempool] unable to acquire the lock")
				}
				mempoolTxs, err := ctx.MempoolModel.GetPendingMempoolTxsByAccountIndex(accountInfo.AccountIndex)
				if err != nil {
					if err != ErrNotFound {
						logx.Errorf("[MonitorMempool] unable to get pending mempool txs: %s", err.Error())
						return err
					}
				}
				for _, mempoolTx := range mempoolTxs {
					for _, txDetail := range mempoolTx.MempoolDetails {
						if txDetail.AccountIndex != accountInfo.AccountIndex || txDetail.AssetId != txInfo.AssetId {
							continue
						}
						if txDetail.AssetType == GeneralAssetType {
							nBalance, err := commonAsset.ComputeNewBalance(GeneralAssetType, accountInfo.AssetInfo[txDetail.AssetId].String(), txDetail.BalanceDelta)
							if err != nil {
								logx.Errorf("[MonitorMempool] unable to compute new balance: %s", err.Error())
								return err
							}
							accountInfo.AssetInfo[txDetail.AssetId], err = commonAsset.ParseAccountAsset(nBalance)
							if err != nil {
								logx.Errorf("[MonitorMempool] unable to parse account asset : %s", err.Error())
								return err
							}
						}
					}
				}
				redisLock.Release()
			}
			// complete oTx info
			txInfo.AccountIndex = accountInfo.AccountIndex
			if accountInfo.AssetInfo[txInfo.AssetId] == nil {
				txInfo.AssetAmount = big.NewInt(0)
			} else {
				txInfo.AssetAmount = accountInfo.AssetInfo[txInfo.AssetId].Balance
			}
			// do delta at committer
			var (
				mempoolTxDetails []*mempool.MempoolTxDetail
			)
			balanceDelta := &commonAsset.AccountAsset{
				AssetId:                  txInfo.AssetId,
				Balance:                  big.NewInt(0),
				LpAmount:                 big.NewInt(0),
				OfferCanceledOrFinalized: big.NewInt(0),
			}
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      txInfo.AssetId,
				AssetType:    commonAsset.GeneralAssetType,
				AccountIndex: txInfo.AccountIndex,
				AccountName:  accountInfo.AccountName,
				BalanceDelta: balanceDelta.String(),
				Order:        0,
				AccountOrder: 0,
			})
			// serialize oTx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize oTx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFee:         commonConstant.NilAssetAmountStr,
				GasFeeAssetId:  commonConstant.NilAssetId,
				NftIndex:       commonConstant.NilTxNftIndex,
				PairIndex:      commonConstant.NilPairIndex,
				AssetId:        txInfo.AssetId,
				TxAmount:       txInfo.AssetAmount.String(),
				NativeAddress:  oTx.SenderAddress,
				MempoolDetails: mempoolTxDetails,
				TxInfo:         string(txInfoBytes),
				AccountIndex:   accountInfo.AccountIndex,
				Nonce:          commonConstant.NilNonce,
				L2BlockHeight:  commonConstant.NilBlockHeight,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			if !relatedAccountIndex[accountInfo.AccountIndex] {
				relatedAccountIndex[accountInfo.AccountIndex] = true
			}
			break
		case TxTypeFullExitNft:
			// create mempool oTx
			var accountInfo *account.Account
			txInfo, err := util.ParseFullExitNftPubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit nft pub data: %s", err.Error())
				return err
			}
			accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
			if newAccountInfoMap[accountNameHash] == nil {
				accountInfo, err = GetAccountInfoByAccountNameHash(accountNameHash, ctx.AccountModel)
				if err != nil {
					logx.Errorf("[MonitorMempool] unable to get account info: %s", err.Error())
					return err
				}
			} else {
				accountInfo = newAccountInfoMap[accountNameHash]
			}
			var (
				nftAsset *nft.L2Nft
			)
			if newNftInfoMap[txInfo.NftIndex] == nil {
				nftAsset, err = ctx.NftModel.GetNftAsset(txInfo.NftIndex)
				if err != nil {
					if err == ErrNotFound {
						emptyNftInfo := commonAsset.EmptyNftInfo(txInfo.NftIndex)
						nftAsset = &nft.L2Nft{
							NftIndex:            emptyNftInfo.NftIndex,
							CreatorAccountIndex: emptyNftInfo.CreatorAccountIndex,
							OwnerAccountIndex:   emptyNftInfo.OwnerAccountIndex,
							NftContentHash:      emptyNftInfo.NftContentHash,
							NftL1Address:        emptyNftInfo.NftL1Address,
							NftL1TokenId:        emptyNftInfo.NftL1TokenId,
							CreatorTreasuryRate: emptyNftInfo.CreatorTreasuryRate,
							CollectionId:        emptyNftInfo.CollectionId,
						}
					} else {
						logx.Errorf("[MonitorMempool] unable to latest nft info: %s", err.Error())
						return err
					}
				} else {
					if nftAsset.OwnerAccountIndex != accountInfo.AccountIndex {
						emptyNftInfo := commonAsset.EmptyNftInfo(txInfo.NftIndex)
						nftAsset = &nft.L2Nft{
							NftIndex:            emptyNftInfo.NftIndex,
							CreatorAccountIndex: emptyNftInfo.CreatorAccountIndex,
							OwnerAccountIndex:   emptyNftInfo.OwnerAccountIndex,
							NftContentHash:      emptyNftInfo.NftContentHash,
							NftL1Address:        emptyNftInfo.NftL1Address,
							NftL1TokenId:        emptyNftInfo.NftL1TokenId,
							CreatorTreasuryRate: emptyNftInfo.CreatorTreasuryRate,
							CollectionId:        emptyNftInfo.CollectionId,
						}
					}
				}
			} else {
				nftAsset = &nft.L2Nft{
					NftIndex:            newNftInfoMap[txInfo.NftIndex].NftIndex,
					CreatorAccountIndex: newNftInfoMap[txInfo.NftIndex].CreatorAccountIndex,
					OwnerAccountIndex:   newNftInfoMap[txInfo.NftIndex].OwnerAccountIndex,
					NftContentHash:      newNftInfoMap[txInfo.NftIndex].NftContentHash,
					NftL1Address:        newNftInfoMap[txInfo.NftIndex].NftL1Address,
					NftL1TokenId:        newNftInfoMap[txInfo.NftIndex].NftL1TokenId,
					CreatorTreasuryRate: newNftInfoMap[txInfo.NftIndex].CreatorTreasuryRate,
					CollectionId:        newNftInfoMap[txInfo.NftIndex].CollectionId,
				}
			}

			var (
				creatorAccountNameHash []byte
			)
			if txInfo.CreatorAccountIndex == 0 && txInfo.CreatorTreasuryRate == 0 {
				creatorAccountNameHash = []byte{0}
			} else {
				creatorAccountInfo, err := ctx.AccountModel.GetAccountByAccountIndex(nftAsset.CreatorAccountIndex)
				if err != nil {
					logx.Errorf("[MonitorMempool] unable to get account info: %s", err.Error())
					return err
				}
				creatorAccountNameHash = common.FromHex(creatorAccountInfo.AccountNameHash)
			}
			// complete oTx info
			nftL1TokenId, isValid := new(big.Int).SetString(nftAsset.NftL1TokenId, Base)
			if !isValid {
				logx.Errorf("[MonitorMempool] unable to parse big int")
				return errors.New("[MonitorMempool] unable to parse big int")
			}
			txInfo = &commonTx.FullExitNftTxInfo{
				TxType:                 txInfo.TxType,
				AccountIndex:           accountInfo.AccountIndex,
				CreatorAccountIndex:    nftAsset.CreatorAccountIndex,
				CreatorTreasuryRate:    nftAsset.CreatorTreasuryRate,
				NftIndex:               txInfo.NftIndex,
				CollectionId:           nftAsset.CollectionId,
				NftL1Address:           nftAsset.NftL1Address,
				AccountNameHash:        txInfo.AccountNameHash,
				CreatorAccountNameHash: creatorAccountNameHash,
				NftContentHash:         common.FromHex(nftAsset.NftContentHash),
				NftL1TokenId:           nftL1TokenId,
			}
			var (
				mempoolTxDetails []*mempool.MempoolTxDetail
			)
			// empty account delta
			emptyAssetDelta := &commonAsset.AccountAsset{
				AssetId:                  0,
				Balance:                  big.NewInt(0),
				LpAmount:                 big.NewInt(0),
				OfferCanceledOrFinalized: big.NewInt(0),
			}
			accountOrder := int64(0)
			order := int64(0)
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      0,
				AssetType:    commonAsset.GeneralAssetType,
				AccountIndex: txInfo.AccountIndex,
				AccountName:  accountInfo.AccountName,
				BalanceDelta: emptyAssetDelta.String(),
				Order:        order,
				AccountOrder: accountOrder,
			})
			// nft info
			newNftInfo := commonAsset.EmptyNftInfo(
				txInfo.NftIndex,
			)
			order++
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      txInfo.NftIndex,
				AssetType:    commonAsset.NftAssetType,
				AccountIndex: txInfo.AccountIndex,
				AccountName:  accountInfo.AccountName,
				BalanceDelta: newNftInfo.String(),
				Order:        order,
				AccountOrder: commonConstant.NilAccountOrder,
			})
			// serialize oTx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize oTx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFee:         commonConstant.NilAssetAmountStr,
				GasFeeAssetId:  commonConstant.NilAssetId,
				NftIndex:       txInfo.NftIndex,
				PairIndex:      commonConstant.NilPairIndex,
				AssetId:        commonConstant.NilAssetId,
				TxAmount:       commonConstant.NilAssetAmountStr,
				NativeAddress:  oTx.SenderAddress,
				MempoolDetails: mempoolTxDetails,
				TxInfo:         string(txInfoBytes),
				AccountIndex:   accountInfo.AccountIndex,
				Nonce:          commonConstant.NilNonce,
				L2BlockHeight:  commonConstant.NilBlockHeight,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			if relatedAccountIndex[accountInfo.AccountIndex] == false {
				relatedAccountIndex[accountInfo.AccountIndex] = true
			}
			break
		default:
			logx.Errorf("[MonitorMempool] invalid oTx type")
			return errors.New("[MonitorMempool] invalid oTx type")
		}
	}
	// transaction: active accounts not in account table & update l2 oTx event & create mempool txs

	logx.Info("====================call CreateMempoolAndActiveAccount=======================")
	logx.Infof("accounts: %v, mempoolTxs: %v, finalL2TxEvents: %v",
		len(pendingNewAccounts),
		len(pendingNewMempoolTxs),
		len(txs))

	// clean cache
	var pendingDeletedKeys []string
	for index := range relatedAccountIndex {
		pendingDeletedKeys = append(pendingDeletedKeys, util.GetAccountKey(index))
	}
	_, _ = ctx.RedisConnection.Del(pendingDeletedKeys...)
	// update db
	err = ctx.L2TxEventMonitorModel.CreateMempoolAndActiveAccount(
		pendingNewAccounts,
		pendingNewMempoolTxs,
		pendingNewLiquidityInfos,
		pendingNewNfts,
		txs,
	)
	if err != nil {
		logx.Errorf("[MonitorMempool] unable to create mempool txs and update l2 oTx event monitors, error: %s",
			err.Error())
		return err
	}
	return nil
}
