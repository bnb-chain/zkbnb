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
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"
	"reflect"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *SendTxLogic) sendTransferTx(rawTxInfo string) (txId string, err error) {
	// parse transfer tx info
	txInfo, err := commonTx.ParseTransferTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx.ParseTransferTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	/*
		Check Params
	*/
	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.AssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid assetId %v", txInfo.AssetId)
		return "", l.HandleCreateTransferFailTx(txInfo, errors.New(errInfo))
	}
	// check param: from account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.FromAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid accountIndex %v", txInfo.FromAccountIndex)
		return "", l.HandleCreateTransferFailTx(txInfo, errors.New(errInfo))
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.ToAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid accountIndex %v", txInfo.ToAccountIndex)
		return "", l.HandleCreateTransferFailTx(txInfo, errors.New(errInfo))
	}
	// check gas account index
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferTx] unable to get sysconfig by name: %s", err.Error())
		return "", l.HandleCreateTransferFailTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendTransferTx] invalid gas account index")
		return "", l.HandleCreateTransferFailTx(txInfo, errors.New("[sendTransferTx] invalid gas account index"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.FormatAccountInfo)
	)
	accountInfoMap[txInfo.FromAccountIndex], err = globalmapHandler.GetLatestAccountInfo(
		l.svcCtx.AccountModel,
		l.svcCtx.AccountHistoryModel,
		l.svcCtx.MempoolDetailModel,
		l.svcCtx.LiquidityPairModel,
		l.svcCtx.RedisConnection,
		txInfo.FromAccountIndex,
	)
	if err != nil {
		logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
		return "", l.HandleCreateTransferFailTx(txInfo, err)
	}
	// get account info by to index
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.ToAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateTransferFailTx(txInfo, err)
		}
	}
	// get account info by gas index
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		// get account info by gas index
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateTransferFailTx(txInfo, err)
		}
	}
	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify transfer tx
	txDetails, err = txVerification.VerifyTransferTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return "", l.HandleCreateTransferFailTx(txInfo, err)
	}

	/*
		Check tx details
	*/

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txId, err = l.CreateTxMempoolForTranferTx(commonTx.TxTypeTransfer, txDetails, txInfo)
	if err != nil {
		return "", l.HandleCreateTransferFailTx(txInfo, err)
	}
	return txId, nil
}

func (l *SendTxLogic) HandleCreateTransferFailTx(txInfo *commonTx.TransferTxInfo, err error) error {
	errCreate := l.CreateFailTransferTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendtransfertxlogic.HandleCreateTransferFailTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateTransferFailTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendTxLogic) CreateFailTransferTx(info *commonTx.TransferTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	txFeeAssetId := info.AssetId
	assetId := info.AssetId
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailTransferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeTransfer,
		// tx fee
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: txFeeAssetId,
		// tx status, 1 - success(default), 2 - failure
		TxStatus: tx.StatusFail,
		// l1asset id
		AssetAId: assetId,
		// AssetBId
		AssetBId: commonConstant.NilAssetId,
		// tx amount
		TxAmount: info.AssetAmount.String(),
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
		// native memo info
		Memo: info.Memo,
	}

	err = l.svcCtx.FailTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailTransferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}

func (l *SendTxLogic) CreateTxMempoolForTranferTx(
	txType uint8,
	nMempoolTxDetails []*mempool.MempoolTxDetail,
	txInfo *commonTx.TransferTxInfo,
) (resTxId string, err error) {
	var (
		nMempoolTx *mempool.MempoolTx
		bTxInfo    []byte
	)
	// generate tx id by random UUID
	resTxId = util.RandomUUID()
	// Marshal txInfo
	bTxInfo, err = json.Marshal(txInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateTxMempoolForTranferTx] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	nMempoolTx = &mempool.MempoolTx{
		TxHash:         resTxId,
		TxType:         int64(txType),
		GasFee:         txInfo.GasFeeAssetAmount.String(),
		GasFeeAssetId:  txInfo.GasFeeAssetId,
		AssetAId:       txInfo.AssetId,
		AssetBId:       commonConstant.NilAssetId,
		TxAmount:       txInfo.AssetAmount.String(),
		MempoolDetails: nMempoolTxDetails,
		TxInfo:         string(bTxInfo),
		Memo:           txInfo.Memo,
		L2BlockHeight:  commonConstant.NilBlockHeight,
		Status:         0,
	}

	// delete cache
	var keys []string
	for _, mempoolTxDetail := range nMempoolTxDetails {
		keys = append(keys, util.GetAccountKey(mempoolTxDetail.AccountIndex))
	}
	_, err = l.svcCtx.RedisConnection.Del(keys...)
	if err != nil {
		logx.Errorf("[CreateTxMempoolForTranferTx] error with redis: %s", err.Error())
		return "", err
	}
	// write into mempool
	err = l.svcCtx.MempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{nMempoolTx})
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateTxMempoolForTranferTx] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	return resTxId, nil
}
