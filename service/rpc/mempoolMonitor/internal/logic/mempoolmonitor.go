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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/commonAsset"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/account"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/asset"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/l2asset"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/mempool"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/util"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/mempoolMonitor/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

func MonitorMempool(
	l2TxEventMonitorModel l2TxEventMonitor.L2TxEventMonitorModel,
	l2assetInfoModel l2asset.L2AssetInfoModel,
	accountModel account.AccountModel,
	accountHistoryModel account.AccountHistoryModel,
	accountAssetModel asset.AccountAssetModel,
	ctx *svc.ServiceContext,
	nContext context.Context,
) error {
	// get tx from l2txEventMonitor
	txs, err := l2TxEventMonitorModel.GetL2TxEventMonitorsByStatus(PendingStatus)
	if err != nil {
		if err == l2TxEventMonitor.ErrNotFound {
			logx.Info("[MonitorMempool] no l2 tx event monitors")
			return err
		} else {
			logx.Error("[MonitorMempool] unable to get l2 tx event monitors")
			return err
		}
	}
	// initialize mempool txs

	nextAccountIndex, err := accountHistoryModel.GetLatestAccountIndex()
	if err != nil {
		if err == account.ErrNotFound {
			nextAccountIndex = -1
		} else {
			errInfo := fmt.Sprintf("[mempoolMoniter.MonitorMempool]<=>[accountModel.GetMaxAccountIndex] %s", err.Error())
			logx.Error(errInfo)
			return err
		}
	}

	var (
		pendingNewAccountHistory []*account.AccountHistory
		pendingNewMempoolTxs     []*mempool.MempoolTx
	)
	// get last handled request id
	currentRequestId, err := l2TxEventMonitorModel.GetLastHandledRequestId()
	if err != nil {
		logx.Errorf("[MonitorMempool] unable to get last handled request id: %s", err.Error())
		return err
	}
	for _, tx := range txs {
		// set tx as handled
		tx.Status = l2TxEventMonitor.HandledStatus
		// request id must be in order
		if tx.RequestId != currentRequestId+1 {
			logx.Errorf("[MonitorMempool] invalid request id")
			return errors.New("[MonitorMempool] invalid request id")
		}
		currentRequestId++

		// handle tx based on tx type
		switch tx.TxType {
		case TxTypeRegisterZns:
			// parse tx info
			txInfo, err := util.ParseRegisterZnsPubData(common.FromHex(tx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse registerZNS pub data: %s", err.Error())
				return err
			}
			// check if the account name has been registered
			_, err = accountHistoryModel.GetAccountByAccountName(txInfo.AccountName)
			if err != ErrNotFound {
				logx.Errorf("[MonitorMempool] account name has been registered")
				return errors.New("[MonitorMempool] account name has been registered")
			}
			// set correct account index
			nextAccountIndex++
			nameHash, err := util.AccountNameHash(txInfo.AccountName)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to compute account name hash: %s", err.Error())
				return err
			}
			// create new account history
			accountHistory := &account.AccountHistory{
				AccountIndex:    nextAccountIndex,
				AccountName:     txInfo.AccountName,
				AccountNameHash: nameHash,
				PublicKey:       txInfo.PubKey,
				L1Address:       tx.SenderAddress,
				Nonce:           -1,
				Status:          account.AccountHistoryStatusPending,
				L2BlockHeight:   -1,
			}
			pendingNewAccountHistory = append(pendingNewAccountHistory, accountHistory)
			// create mempool tx
			// serialize tx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize tx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:        RandomTxHash(),
				TxType:        int64(txInfo.TxType),
				GasFee:        0,
				GasFeeAssetId: 0,
				AssetAId:      -1,
				AssetBId:      -1,
				TxAmount:      "0",
				NativeAddress: tx.SenderAddress,
				TxInfo:        string(txInfoBytes),
				AccountIndex:  accountHistory.AccountIndex,
				Nonce:         -1,
				L2BlockHeight: -1,
				Status:        mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			break
		case TxTypeDeposit:
			// create mempool tx
			txInfo, err := util.ParseDepositPubData(common.FromHex(tx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit pub data: %s", err.Error())
				return err
			}
			var (
				accountInfo *account.Account
			)
			accountHistoryInfo, err := accountHistoryModel.GetAccountByAccountNameHash(txInfo.AccountNameHash)
			if err != nil {
				if err == ErrNotFound {
					accountInfo, err = accountModel.GetAccountByAccountNameHash(txInfo.AccountNameHash)
					if err != nil {
						logx.Errorf("[MonitorMempool] unable to get account by account name hash: %s", err.Error())
						return err
					}
				} else {
					logx.Errorf("[MonitorMempool] unable to get account history by account name hash: %s", err.Error())
					return err
				}
			} else {
				accountInfo = &account.Account{
					AccountIndex:    accountHistoryInfo.AccountIndex,
					AccountName:     accountHistoryInfo.AccountName,
					PublicKey:       accountHistoryInfo.PublicKey,
					AccountNameHash: accountHistoryInfo.AccountNameHash,
					L1Address:       accountHistoryInfo.L1Address,
					Nonce:           accountHistoryInfo.Nonce,
				}
			}
			txInfo.AccountIndex = uint32(accountInfo.AccountIndex)
			var (
				mempoolTxDetails []*mempool.MempoolTxDetail
			)
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      int64(txInfo.AssetId),
				AssetType:    commonAsset.GeneralAssetType,
				AccountIndex: int64(txInfo.AccountIndex),
				AccountName:  accountInfo.AccountName,
				BalanceDelta: txInfo.AssetAmount.String(),
			})
			// serialize tx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize tx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFee:         0,
				GasFeeAssetId:  0,
				AssetAId:       int64(txInfo.AssetId),
				AssetBId:       -1,
				TxAmount:       txInfo.AssetAmount.String(),
				NativeAddress:  tx.SenderAddress,
				MempoolDetails: mempoolTxDetails,
				TxInfo:         string(txInfoBytes),
				AccountIndex:   accountInfo.AccountIndex,
				Nonce:          -1,
				L2BlockHeight:  -1,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			break
		case TxTypeDepositNft:
			// create mempool tx
			txInfo, err := util.ParseDepositNftPubData(common.FromHex(tx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit nft pub data: %s", err.Error())
				return err
			}
			var (
				accountInfo *account.Account
			)
			accountHistoryInfo, err := accountHistoryModel.GetAccountByAccountNameHash(txInfo.AccountNameHash)
			if err != nil {
				if err == ErrNotFound {
					accountInfo, err = accountModel.GetAccountByAccountNameHash(txInfo.AccountNameHash)
					if err != nil {
						logx.Errorf("[MonitorMempool] unable to get account by account name hash: %s", err.Error())
						return err
					}
				} else {
					logx.Errorf("[MonitorMempool] unable to get account history by account name hash: %s", err.Error())
					return err
				}
			} else {
				accountInfo = &account.Account{
					AccountIndex:    accountHistoryInfo.AccountIndex,
					AccountName:     accountHistoryInfo.AccountName,
					PublicKey:       accountHistoryInfo.PublicKey,
					AccountNameHash: accountHistoryInfo.AccountNameHash,
					L1Address:       accountHistoryInfo.L1Address,
					Nonce:           accountHistoryInfo.Nonce,
				}
			}
			// complete tx info
			txInfo.AccountIndex = uint32(accountInfo.AccountIndex)
			// TODO nft index
			txInfo.NftIndex = uint64(0)
			// TODO get nft content hash
			txInfo.NftContentHash = []byte("")
			var (
				mempoolTxDetails []*mempool.MempoolTxDetail
			)
			delta, err := util.ConstructNftInfo(
				int64(txInfo.NftIndex),
				accountInfo.AccountIndex,
				accountInfo.AccountIndex,
				-1,
				"0",
				common.Bytes2Hex(txInfo.NftContentHash),
				txInfo.NftL1TokenId.String(),
				txInfo.NftL1Address,
			)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to construct nft info: %s", err.Error())
				return err
			}
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      int64(txInfo.NftIndex),
				AssetType:    commonAsset.GeneralAssetType,
				AccountIndex: int64(txInfo.AccountIndex),
				AccountName:  accountInfo.AccountName,
				Balance:      util.EmptyNftInfo(int64(txInfo.NftIndex)).String(),
				BalanceDelta: delta.String(),
			})
			// serialize tx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize tx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFee:         0,
				GasFeeAssetId:  0,
				AssetAId:       -1,
				AssetBId:       -1,
				TxAmount:       "0",
				NativeAddress:  tx.SenderAddress,
				MempoolDetails: mempoolTxDetails,
				TxInfo:         string(txInfoBytes),
				AccountIndex:   accountInfo.AccountIndex,
				Nonce:          -1,
				L2BlockHeight:  -1,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			break
		case TxTypeFullExit:
			// create mempool tx
			txInfo, err := util.ParseFullExitPubData(common.FromHex(tx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit pub data: %s", err.Error())
				return err
			}
			var (
				accountInfo *account.Account
			)
			accountHistoryInfo, err := accountHistoryModel.GetAccountByAccountNameHash(txInfo.AccountNameHash)
			if err != nil {
				if err == ErrNotFound {
					accountInfo, err = accountModel.GetAccountByAccountNameHash(txInfo.AccountNameHash)
					if err != nil {
						logx.Errorf("[MonitorMempool] unable to get account by account name hash: %s", err.Error())
						return err
					}
				} else {
					logx.Errorf("[MonitorMempool] unable to get account history by account name hash: %s", err.Error())
					return err
				}
			} else {
				accountInfo = &account.Account{
					AccountIndex:    accountHistoryInfo.AccountIndex,
					AccountName:     accountHistoryInfo.AccountName,
					PublicKey:       accountHistoryInfo.PublicKey,
					AccountNameHash: accountHistoryInfo.AccountNameHash,
					L1Address:       accountHistoryInfo.L1Address,
					Nonce:           accountHistoryInfo.Nonce,
				}
			}
			// complete tx info
			txInfo.AccountIndex = uint32(accountInfo.AccountIndex)
			// TODO get remaining asset amount
			txInfo.AssetAmount = big.NewInt(0)
			var (
				mempoolTxDetails []*mempool.MempoolTxDetail
			)
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      int64(txInfo.AssetId),
				AssetType:    commonAsset.GeneralAssetType,
				AccountIndex: int64(txInfo.AccountIndex),
				AccountName:  accountInfo.AccountName,
				BalanceDelta: ffmath.Neg(txInfo.AssetAmount).String(),
			})
			// serialize tx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize tx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFee:         0,
				GasFeeAssetId:  0,
				AssetAId:       int64(txInfo.AssetId),
				AssetBId:       -1,
				TxAmount:       txInfo.AssetAmount.String(),
				NativeAddress:  tx.SenderAddress,
				MempoolDetails: mempoolTxDetails,
				TxInfo:         string(txInfoBytes),
				AccountIndex:   accountInfo.AccountIndex,
				Nonce:          -1,
				L2BlockHeight:  -1,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			break
		case TxTypeFullExitNft:
			// create mempool tx
			txInfo, err := util.ParseFullExitNftPubData(common.FromHex(tx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit nft pub data: %s", err.Error())
				return err
			}
			var (
				accountInfo *account.Account
			)
			accountHistoryInfo, err := accountHistoryModel.GetAccountByAccountNameHash(txInfo.AccountNameHash)
			if err != nil {
				if err == ErrNotFound {
					accountInfo, err = accountModel.GetAccountByAccountNameHash(txInfo.AccountNameHash)
					if err != nil {
						logx.Errorf("[MonitorMempool] unable to get account by account name hash: %s", err.Error())
						return err
					}
				} else {
					logx.Errorf("[MonitorMempool] unable to get account history by account name hash: %s", err.Error())
					return err
				}
			} else {
				accountInfo = &account.Account{
					AccountIndex:    accountHistoryInfo.AccountIndex,
					AccountName:     accountHistoryInfo.AccountName,
					PublicKey:       accountHistoryInfo.PublicKey,
					AccountNameHash: accountHistoryInfo.AccountNameHash,
					L1Address:       accountHistoryInfo.L1Address,
					Nonce:           accountHistoryInfo.Nonce,
				}
			}
			// complete tx info
			txInfo.AccountIndex = uint32(accountInfo.AccountIndex)
			// TODO get nft info
			txInfo.NftContentHash = []byte("")
			txInfo.NftL1TokenId = big.NewInt(0)
			txInfo.Amount = 1
			txInfo.NftL1Address = ""

			var (
				mempoolTxDetails []*mempool.MempoolTxDetail
			)
			// TODO get old nft info
			oldNftInfo := util.EmptyNftInfo(int64(txInfo.NftIndex))
			newNftInfo := util.EmptyNftInfo(
				int64(txInfo.NftIndex),
			)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to construct nft info: %s", err.Error())
				return err
			}
			mempoolTxDetails = append(mempoolTxDetails, &mempool.MempoolTxDetail{
				AssetId:      int64(txInfo.NftIndex),
				AssetType:    commonAsset.GeneralAssetType,
				AccountIndex: int64(txInfo.AccountIndex),
				AccountName:  accountInfo.AccountName,
				Balance:      oldNftInfo.String(),
				BalanceDelta: newNftInfo.String(),
			})
			// serialize tx info
			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to serialize tx info : %s", err.Error())
				return err
			}
			mempoolTx := &mempool.MempoolTx{
				TxHash:         RandomTxHash(),
				TxType:         int64(txInfo.TxType),
				GasFee:         0,
				GasFeeAssetId:  0,
				AssetAId:       -1,
				AssetBId:       -1,
				TxAmount:       "0",
				NativeAddress:  tx.SenderAddress,
				MempoolDetails: mempoolTxDetails,
				TxInfo:         string(txInfoBytes),
				AccountIndex:   accountInfo.AccountIndex,
				Nonce:          -1,
				L2BlockHeight:  -1,
				Status:         mempool.PendingTxStatus,
			}
			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
			break
		default:
			logx.Errorf("[MonitorMempool] invalid tx type")
			return errors.New("[MonitorMempool] invalid tx type")
		}
	}
	// transaction: active accounts not in account table & update l2 tx event & create mempool txs

	logx.Info("====================call CreateMempoolAndActiveAccount=======================")
	logx.Infof("mempoolTxs: %v, finalL2TxEvents: %v, nextAccountIndex: %v", len(pendingNewMempoolTxs),
		len(txs), nextAccountIndex)

	err = ctx.L2TxEventMonitorModel.CreateMempoolAndActiveAccount(pendingNewAccountHistory, pendingNewMempoolTxs, txs)
	if err != nil {
		logx.Errorf("[MonitorMempool] unable to create mempool txs and update l2 tx event monitors, error: %s",
			err.Error())
		return err
	} else {
		// TODO globalRpc
		//resRpc, err := ctx.GlobalRPC.ResetGlobalMap(nContext, &globalrpc.ReqResetGlobalMap{})
		//if err != nil {
		//	errInfo := fmt.Sprintf("[MonitorMempool]<=>[GlobalRPC.ResetGlobalMap] err is not nil: %s",
		//		err.Error())
		//	logx.Errorf(errInfo)
		//	return errors.New(errInfo)
		//} else if resRpc == nil {
		//	errInfo := fmt.Sprintf("[MonitorMempool]<=>[GlobalRPC.ResetGlobalMap] resRpc is nil")
		//	logx.Errorf(errInfo)
		//	return errors.New(errInfo)
		//} else if resRpc.Status != 0 {
		//	errInfo := fmt.Sprintf("[MonitorMempool]<=>[GlobalRPC.ResetGlobalMap] resRpc failed, %s", resRpc.Err)
		//	logx.Errorf(errInfo)
		//	return errors.New(errInfo)
		//} else {
		//	logx.Info("[GlobalRPC.ResetGlobalMap] Successful")
		//}
	}
	return nil
}
