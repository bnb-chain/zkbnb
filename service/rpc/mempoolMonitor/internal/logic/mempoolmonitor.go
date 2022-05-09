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
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/mempoolMonitor/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
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

	nextAccountIndex, err := ctx.AccountHistoryModel.GetLatestAccountIndex()
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
		pendingNewAccount        []*account.Account
		pendingNewAccountHistory []*account.AccountHistory
		pendingNewMempoolTxs     []*mempool.MempoolTx
		newAccountInfoMap        = make(map[string]*account.Account)
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
			nextAccountIndex++
			// create new account and account history
			accountInfo := &account.Account{
				AccountIndex:    nextAccountIndex,
				AccountName:     txInfo.AccountName,
				PublicKey:       txInfo.PubKey,
				AccountNameHash: txInfo.AccountNameHash,
				L1Address:       oTx.SenderAddress,
				Nonce:           commonConstant.NilNonce,
				AssetInfo:       commonConstant.EmptyAsset,
				AssetRoot:       commonConstant.NilHashStr,
				LiquidityInfo:   commonConstant.EmptyLiquidity,
				LiquidityRoot:   commonConstant.NilHashStr,
				Status:          account.AccountStatusPending,
			}
			accountHistory := &account.AccountHistory{
				AccountIndex:  nextAccountIndex,
				Nonce:         commonConstant.NilNonce,
				AssetInfo:     commonConstant.EmptyAsset,
				AssetRoot:     commonConstant.NilHashStr,
				LiquidityInfo: commonConstant.EmptyLiquidity,
				LiquidityRoot: commonConstant.NilHashStr,
				Status:        account.AccountHistoryStatusPending,
				L2BlockHeight: commonConstant.NilBlockHeight,
			}
			pendingNewAccount = append(pendingNewAccount, accountInfo)
			pendingNewAccountHistory = append(pendingNewAccountHistory, accountHistory)
			newAccountInfoMap[txInfo.AccountNameHash] = accountInfo
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
				GasFee:        commonConstant.NilAssetAmountStr,
				GasFeeAssetId: commonConstant.NilAssetId,
				AssetAId:      commonConstant.NilAssetId,
				AssetBId:      commonConstant.NilAssetId,
				TxAmount:      commonConstant.NilAssetAmountStr,
				NativeAddress: oTx.SenderAddress,
				TxInfo:        string(txInfoBytes),
				AccountIndex:  accountHistory.AccountIndex,
				Nonce:         commonConstant.NilNonce,
				L2BlockHeight: commonConstant.NilBlockHeight,
				Status:        mempool.PendingTxStatus,
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
			if newAccountInfoMap[txInfo.AccountNameHash] != nil {
				accountInfo = &account.Account{
					AccountIndex:    newAccountInfoMap[txInfo.AccountNameHash].AccountIndex,
					AccountName:     newAccountInfoMap[txInfo.AccountNameHash].AccountName,
					PublicKey:       newAccountInfoMap[txInfo.AccountNameHash].PublicKey,
					AccountNameHash: newAccountInfoMap[txInfo.AccountNameHash].AccountNameHash,
					L1Address:       newAccountInfoMap[txInfo.AccountNameHash].L1Address,
					Nonce:           newAccountInfoMap[txInfo.AccountNameHash].Nonce,
				}
			} else {
				accountInfo, err = GetAccountInfoByAccountNameHash(txInfo.AccountNameHash, ctx.AccountModel)
				if err != nil {
					logx.Errorf("[MonitorMempool] unable to get account info: %s", err.Error())
					return err
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
				AssetAId:       int64(txInfo.AssetId),
				AssetBId:       commonConstant.NilAssetId,
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
			if relatedAccountIndex[accountInfo.AccountIndex] == false {
				relatedAccountIndex[accountInfo.AccountIndex] = true
			}
			break
		case TxTypeDepositNft:
			// create mempool oTx
			txInfo, err := util.ParseDepositNftPubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit nft pub data: %s", err.Error())
				return err
			}
			accountInfo, err := GetAccountInfoByAccountNameHash(txInfo.AccountNameHash, ctx.AccountModel)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to get account info: %s", err.Error())
				return err
			}
			// complete oTx info
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
				commonConstant.NilAssetId,
				commonConstant.NilAssetAmountStr,
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
				BalanceDelta: delta.String(),
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
				AssetAId:       commonConstant.NilAssetId,
				AssetBId:       commonConstant.NilAssetId,
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
		case TxTypeFullExit:
			// create mempool oTx
			txInfo, err := util.ParseFullExitPubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit pub data: %s", err.Error())
				return err
			}
			accountInfo, err := GetAccountInfoByAccountNameHash(txInfo.AccountNameHash, ctx.AccountModel)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to get account info: %s", err.Error())
				return err
			}
			// complete oTx info
			txInfo.AccountIndex = uint32(accountInfo.AccountIndex)
			// TODO get remaining asset amount
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
				AssetAId:       int64(txInfo.AssetId),
				AssetBId:       commonConstant.NilAssetId,
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
			if relatedAccountIndex[accountInfo.AccountIndex] == false {
				relatedAccountIndex[accountInfo.AccountIndex] = true
			}
			break
		case TxTypeFullExitNft:
			// create mempool oTx
			txInfo, err := util.ParseFullExitNftPubData(common.FromHex(oTx.Pubdata))
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to parse deposit nft pub data: %s", err.Error())
				return err
			}
			accountInfo, err := GetAccountInfoByAccountNameHash(txInfo.AccountNameHash, ctx.AccountModel)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to get account info: %s", err.Error())
				return err
			}
			// complete oTx info
			txInfo.AccountIndex = uint32(accountInfo.AccountIndex)
			// TODO get nft info
			txInfo.NftContentHash = []byte("")
			txInfo.NftL1TokenId = big.NewInt(0)
			txInfo.Amount = 1
			txInfo.NftL1Address = ""

			var (
				mempoolTxDetails []*mempool.MempoolTxDetail
			)
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
				BalanceDelta: newNftInfo.String(),
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
				AssetAId:       commonConstant.NilAssetId,
				AssetBId:       commonConstant.NilAssetId,
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
	logx.Infof("accounts: %v, accountHistories: %v, mempoolTxs: %v, finalL2TxEvents: %v, nextAccountIndex: %v",
		len(pendingNewAccount), len(pendingNewAccountHistory),
		len(pendingNewMempoolTxs),
		len(txs), nextAccountIndex)

	// clean cache
	var pendingDeletedKeys []string
	for index, _ := range relatedAccountIndex {
		pendingDeletedKeys = append(pendingDeletedKeys, util.GetAccountKey(index))
	}
	_, _ = ctx.RedisConnection.Del(pendingDeletedKeys...)
	// update db
	err = ctx.L2TxEventMonitorModel.CreateMempoolAndActiveAccount(pendingNewAccount, pendingNewAccountHistory, pendingNewMempoolTxs, txs)
	if err != nil {
		logx.Errorf("[MonitorMempool] unable to create mempool txs and update l2 oTx event monitors, error: %s",
			err.Error())
		return err
	}
	return nil
}
