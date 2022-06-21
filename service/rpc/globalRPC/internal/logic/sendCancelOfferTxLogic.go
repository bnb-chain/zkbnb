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
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"math/big"
	"reflect"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *SendTxLogic) sendCancelOfferTx(rawTxInfo string) (txId string, err error) {
	// parse transfer tx info
	txInfo, err := commonTx.ParseCancelOfferTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendCancelOfferTx.ParseCancelOfferTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	/*
		Check Params
	*/
	// check param: from account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendCancelOfferTx] err: invalid accountIndex %v", txInfo.AccountIndex)
		return "", l.HandleCreateFailCancelOfferTx(txInfo, errors.New(errInfo))
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.GasAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendCancelOfferTx] err: invalid accountIndex %v", txInfo.GasAccountIndex)
		return "", l.HandleCreateFailCancelOfferTx(txInfo, errors.New(errInfo))
	}
	// check gas account index
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendCancelOfferTx] unable to get sysconfig by name: %s", err.Error())
		return "", l.HandleCreateFailCancelOfferTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", l.HandleCreateFailCancelOfferTx(txInfo, errors.New("[sendCancelOfferTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendCancelOfferTx] invalid gas account index")
		return "", l.HandleCreateFailCancelOfferTx(txInfo, errors.New("[sendCancelOfferTx] invalid gas account index"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendCancelOfferTx] invalid time stamp")
		return "", l.HandleCreateFailCancelOfferTx(txInfo, errors.New("[sendCancelOfferTx] invalid time stamp"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.AccountIndex], err = globalmapHandler.GetLatestAccountInfo(
		l.svcCtx.AccountModel,
		l.svcCtx.MempoolModel,
		l.svcCtx.RedisConnection,
		txInfo.AccountIndex,
	)
	if err != nil {
		logx.Errorf("[sendCancelOfferTx] unable to get account info: %s", err.Error())
		return "", l.HandleCreateFailCancelOfferTx(txInfo, err)
	}
	offerAssetId := txInfo.OfferId / 128
	offerIndex := txInfo.OfferId % 128
	if accountInfoMap[txInfo.AccountIndex].AssetInfo[offerAssetId] == nil {
		accountInfoMap[txInfo.AccountIndex].AssetInfo[offerAssetId] = &commonAsset.AccountAsset{
			AssetId:                  0,
			Balance:                  big.NewInt(0),
			LpAmount:                 big.NewInt(0),
			OfferCanceledOrFinalized: big.NewInt(0),
		}
	} else {
		offerInfo := accountInfoMap[txInfo.AccountIndex].AssetInfo[offerAssetId].OfferCanceledOrFinalized
		xBit := offerInfo.Bit(int(offerIndex))
		if xBit == 1 {
			logx.Errorf("[sendCancelOfferTx] the offer is already confirmed or canceled")
			return "", l.HandleCreateFailCancelOfferTx(txInfo, errors.New("[sendCancelOfferTx] the offer is already confirmed or canceled"))
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
			logx.Errorf("[sendCancelOfferTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateFailCancelOfferTx(txInfo, err)
		}
	}
	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify transfer tx
	txDetails, err = txVerification.VerifyCancelOfferTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return "", l.HandleCreateFailCancelOfferTx(txInfo, err)
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
		return "", l.HandleCreateFailCancelOfferTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeCancelOffer,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		commonConstant.NilPairIndex,
		commonConstant.NilAssetId,
		accountInfoMap[txInfo.AccountIndex].AccountName,
		commonConstant.NilL1Address,
		string(txInfoBytes),
		"",
		txInfo.AccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return "", l.HandleCreateFailCancelOfferTx(txInfo, err)
	}
	return txId, nil
}

func (l *SendTxLogic) HandleCreateFailCancelOfferTx(txInfo *commonTx.CancelOfferTxInfo, err error) error {
	errCreate := l.CreateFailCancelOfferTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendtransfertxlogic.HandleCreateFailCancelOfferTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateFailCancelOfferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendTxLogic) CreateFailCancelOfferTx(info *commonTx.CancelOfferTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailCancelOfferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeCancelOffer,
		// tx fee
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: info.GasFeeAssetId,
		// tx status, 1 - success(default), 2 - failure
		TxStatus: tx.StatusFail,
		// l1asset id
		AssetAId: commonConstant.NilAssetId,
		// AssetBId
		AssetBId: commonConstant.NilAssetId,
		// tx amount
		TxAmount: commonConstant.NilAssetAmountStr,
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
		// native memo info
		Memo: "",
	}

	err = l.svcCtx.FailTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailCancelOfferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
