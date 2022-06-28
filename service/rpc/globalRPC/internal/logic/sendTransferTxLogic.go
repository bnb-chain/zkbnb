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
	"reflect"
	"strconv"
	"time"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"

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
		return "", l.HandleCreateFailTransferTx(txInfo, errors.New(errInfo))
	}
	// check param: from account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.FromAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid accountIndex %v", txInfo.FromAccountIndex)
		return "", l.HandleCreateFailTransferTx(txInfo, errors.New(errInfo))
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.ToAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid accountIndex %v", txInfo.ToAccountIndex)
		return "", l.HandleCreateFailTransferTx(txInfo, errors.New(errInfo))
	}
	// check gas account index
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferTx] unable to get sysconfig by name: %s", err.Error())
		return "", l.HandleCreateFailTransferTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", l.HandleCreateFailTransferTx(txInfo, errors.New("[sendTransferTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendTransferTx] invalid gas account index")
		return "", l.HandleCreateFailTransferTx(txInfo, errors.New("[sendTransferTx] invalid gas account index"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendTransferTx] invalid time stamp")
		return "", l.HandleCreateFailTransferTx(txInfo, errors.New("[sendTransferTx] invalid time stamp"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.FromAccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
		return "", l.HandleCreateFailTransferTx(txInfo, err)
	}
	// get account info by to index
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.ToAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateFailTransferTx(txInfo, err)
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("[sendTransferTx] invalid account name")
		return "", l.HandleCreateFailTransferTx(txInfo, errors.New("[sendTransferTx] invalid account name"))
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
			return "", l.HandleCreateFailTransferTx(txInfo, err)
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
		return "", l.HandleCreateFailTransferTx(txInfo, err)
	}

	/*
		Check tx details
	*/

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", l.HandleCreateFailTransferTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeTransfer,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		commonConstant.NilPairIndex,
		txInfo.AssetId,
		txInfo.AssetAmount.String(),
		"",
		string(txInfoBytes),
		txInfo.Memo,
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return "", l.HandleCreateFailTransferTx(txInfo, err)
	}
	return txId, nil
}

func (l *SendTxLogic) HandleCreateFailTransferTx(txInfo *commonTx.TransferTxInfo, err error) error {
	errCreate := l.CreateFailTransferTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendtransfertxlogic.HandleCreateFailTransferTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateFailTransferTx] %s", err.Error())
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
