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
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *SendTxLogic) sendAtomicMatchTx(rawTxInfo string) (txId string, err error) {
	// parse transfer tx info
	txInfo, err := commonTx.ParseAtomicMatchTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx.ParseAtomicMatchTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	/*
		Check Params
	*/
	// check param: from account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx] err: invalid accountIndex %v", txInfo.AccountIndex)
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, errors.New(errInfo))
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.BuyOffer.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx] err: invalid accountIndex %v", txInfo.BuyOffer.AccountIndex)
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, errors.New(errInfo))
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.SellOffer.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx] err: invalid accountIndex %v", txInfo.SellOffer.AccountIndex)
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, errors.New(errInfo))
	}
	// check gas account index
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to get sysconfig by name: %s", err.Error())
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, errors.New("[sendAtomicMatchTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendAtomicMatchTx] invalid gas account index")
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, errors.New("[sendAtomicMatchTx] invalid gas account index"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now || txInfo.BuyOffer.ExpiredAt < now || txInfo.SellOffer.ExpiredAt < now {
		logx.Errorf("[sendAtomicMatchTx] invalid time stamp")
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, errors.New("[sendAtomicMatchTx] invalid time stamp"))
	}

	if txInfo.BuyOffer.NftIndex != txInfo.SellOffer.NftIndex ||
		txInfo.BuyOffer.AssetId != txInfo.SellOffer.AssetId ||
		txInfo.BuyOffer.AssetAmount.String() != txInfo.SellOffer.AssetAmount.String() ||
		txInfo.BuyOffer.TreasuryRate != txInfo.SellOffer.TreasuryRate {
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, errors.New("[sendAtomicMatchTx] invalid params"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	nftInfo, err := globalmapHandler.GetLatestNftInfoForRead(
		l.svcCtx.NftModel,
		l.svcCtx.MempoolModel,
		l.svcCtx.RedisConnection,
		txInfo.BuyOffer.NftIndex,
	)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to get latest nft index: %s", err.Error())
		return "", err
	}
	accountInfoMap[txInfo.AccountIndex], err = globalmapHandler.GetLatestAccountInfo(
		l.svcCtx.AccountModel,
		l.svcCtx.MempoolModel,
		l.svcCtx.RedisConnection,
		txInfo.AccountIndex,
	)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, err)
	}
	// get account info by to index
	if accountInfoMap[txInfo.BuyOffer.AccountIndex] == nil {
		accountInfoMap[txInfo.BuyOffer.AccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.BuyOffer.AccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateFailAtomicMatchTx(txInfo, err)
		}
	}
	if accountInfoMap[nftInfo.CreatorAccountIndex] == nil {
		accountInfoMap[nftInfo.CreatorAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			nftInfo.CreatorAccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateFailAtomicMatchTx(txInfo, err)
		}
	}
	if accountInfoMap[txInfo.SellOffer.AccountIndex] == nil {
		accountInfoMap[txInfo.SellOffer.AccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.SellOffer.AccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateFailAtomicMatchTx(txInfo, err)
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
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateFailAtomicMatchTx(txInfo, err)
		}
	}
	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify transfer tx
	txDetails, err = txVerification.VerifyAtomicMatchTxInfo(
		accountInfoMap,
		nftInfo,
		txInfo,
	)
	if err != nil {
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, err)
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
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeAtomicMatch,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		txInfo.BuyOffer.NftIndex,
		commonConstant.NilPairIndex,
		txInfo.BuyOffer.AssetId,
		txInfo.BuyOffer.AssetAmount.String(),
		"",
		string(txInfoBytes),
		"",
		txInfo.AccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return "", l.HandleCreateFailAtomicMatchTx(txInfo, err)
	}
	return txId, nil
}

func (l *SendTxLogic) HandleCreateFailAtomicMatchTx(txInfo *commonTx.AtomicMatchTxInfo, err error) error {
	errCreate := l.CreateFailAtomicMatchTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendtransfertxlogic.HandleCreateFailAtomicMatchTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateFailAtomicMatchTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendTxLogic) CreateFailAtomicMatchTx(info *commonTx.AtomicMatchTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailAtomicMatchTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeAtomicMatch,
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
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailAtomicMatchTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
