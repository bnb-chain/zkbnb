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
	"github.com/zecrey-labs/zecrey/common/commonAsset"
	"reflect"

	"github.com/zecrey-labs/zecrey/common/commonAccount"
	"github.com/zecrey-labs/zecrey/common/utils"

	"github.com/zecrey-labs/zecrey/common/commonTx"
	"github.com/zecrey-labs/zecrey/common/model/mempool"
	"github.com/zecrey-labs/zecrey/common/model/tx"
	"github.com/zecrey-labs/zecrey/common/zcrypto/cryptoUtils"
	"github.com/zecrey-labs/zecrey/common/zcrypto/zecreyProofs"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/logic/globalmapHandler"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/logic/txHandler"
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
	err = utils.CheckRequestParam(utils.TypeAssetId, reflect.ValueOf(txInfo.AssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid assetId %v", txInfo.AssetId)
		return "", l.HandleCreateTransferFailTx(txInfo, errors.New(errInfo))
	}

	var (
		accountMap = make(map[uint32]bool)
	)
	for _, accountIndex := range txInfo.AccountsIndex {
		err = utils.CheckRequestParam(utils.TypeAccountIndex, reflect.ValueOf(accountIndex))
		if err != nil {
			errInfo := fmt.Sprintf("[sendTransferTx] err: invalid accountIndex %v", txInfo.AccountsIndex)
			return "", l.HandleCreateTransferFailTx(txInfo, errors.New(errInfo))
		}
		if accountMap[accountIndex] == true {
			errInfo := fmt.Sprintf("[sendTransferTx] err: duplicated accountIndex %v", txInfo.AccountsIndex)
			return "", l.HandleCreateTransferFailTx(txInfo, errors.New(errInfo))
		} else {
			accountMap[accountIndex] = true
		}
	}

	err = utils.CheckRequestParam(utils.TypeGasFee, reflect.ValueOf(txInfo.GasFee))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid gas fee %v", txInfo.GasFee)
		return "", l.HandleCreateTransferFailTx(txInfo, errors.New(errInfo))
	}

	var (
		accountAssetInfoList   []*zecreyProofs.AccountAssetInfo
		gasAssetInfo           *zecreyProofs.AccountAssetInfo
		accountSingleAssetList []*AccountSingleAsset
		_                      *AccountSingleAsset
	)
	/*
		Construct accountAssetInfoList & accountSingleAssetList
	*/
	for _, accountIndex := range txInfo.AccountsIndex {
		// get accountInfo by accountIndex
		accountInfo, err := GetLatestSingleAccountAsset(l.svcCtx, accountIndex, txInfo.AssetId)
		if err != nil {
			return "", l.HandleCreateTransferFailTx(txInfo, err)
		}

		nAccountInfo, err := zecreyProofs.ConstructAccountAssetInfo(
			accountInfo.AccountId,
			int64(accountInfo.AccountIndex),
			accountInfo.AccountName,
			accountInfo.PublicKey,
			int64(accountInfo.AssetId),
			accountInfo.BalanceEnc,
		)
		if err != nil {
			return "", l.HandleCreateTransferFailTx(txInfo, err)
		}
		accountAssetInfoList = append(accountAssetInfoList, nAccountInfo)
		accountSingleAssetList = append(accountSingleAssetList, accountInfo)
	}
	gasAccountAsset, err := GetLatestSingleAccountAsset(l.svcCtx, commonAccount.GasAccountIndex, txInfo.AssetId)
	if err != nil {
		return "", l.HandleCreateTransferFailTx(txInfo, err)
	}
	gasAssetInfo, err = zecreyProofs.ConstructAccountAssetInfo(
		gasAccountAsset.AccountId,
		int64(gasAccountAsset.AccountIndex),
		gasAccountAsset.AccountName,
		gasAccountAsset.PublicKey,
		int64(gasAccountAsset.AssetId),
		gasAccountAsset.BalanceEnc,
	)
	if err != nil {
		return "", l.HandleCreateTransferFailTx(txInfo, err)
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	/*
		Get txDetails
	*/

	res, err := json.Marshal(accountAssetInfoList)
	logx.Info("accountAssetInfoList:", string(res))
	res, err = json.Marshal(gasAssetInfo)
	logx.Info("gasAssetInfo:", string(res))
	res, err = json.Marshal(txInfo)
	logx.Info("txInfo:", string(res))

	// verify transfer tx
	txDetails, err = zecreyProofs.VerifyTransferTx(accountAssetInfoList, gasAssetInfo, txInfo)
	if err != nil {
		return "", l.HandleCreateTransferFailTx(txInfo, err)
	}

	/*
		Check tx details
	*/

	// check txDetails length
	if len(txDetails) != zecreyProofs.TransferTxDetailsCount {
		errInfo := fmt.Sprintf("[sendtxlogic.sendTransferTx] txDetails count error, len(txDetails) = %v", len(txDetails))
		return "", l.HandleCreateTransferFailTx(txInfo, errors.New(errInfo))
	}
	// check accountIndex && assetId in txDetail
	for _, v := range txDetails {
		if !(uint32(v.AssetId) == txInfo.AssetId) {
			errInfo := fmt.Sprintf("[sendtxlogic.sendTransferTx] txDetail error, txDetail = %v", v)
			return "", l.HandleCreateTransferFailTx(txInfo, errors.New(errInfo))
		}
	}

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txId, err = l.CreateTxMempoolForTranferTx(txDetails, txInfo, txHandler.TxTypeTransfer)
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
	txHash := cryptoUtils.GetRandomUUID()
	txType := int64(txHandler.TxTypeTransfer)
	txFee := info.GasFee
	txFeeAssetId := info.AssetId
	assetId := info.AssetId
	txAmount := int64(0)
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
		TxType: txType,
		// tx fee
		GasFee: int64(txFee),
		// tx fee l1asset id
		GasFeeAssetId: int64(txFeeAssetId),
		// tx status, 1 - success(default), 2 - failure
		TxStatus: txHandler.TxFail,
		// l1asset id
		AssetAId: int64(assetId),
		// AssetBId
		AssetBId: commonAsset.NilAssetId,
		// ChainId
		ChainId: commonTx.L2TxChainId,
		// tx amount
		TxAmount: int64(txAmount),
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

func (l *SendTxLogic) CreateTxMempoolForTranferTx(nMempoolTxDetails []*mempool.MempoolTxDetail, txInfo *commonTx.TransferTxInfo, txType uint8) (resTxId string, err error) {

	var (
		nMempoolTx *mempool.MempoolTx
		bTxInfo    []byte
	)
	// generate tx id by random UUID
	resTxId = cryptoUtils.GetRandomUUID()
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
		GasFee:         int64(txInfo.GasFee),
		GasFeeAssetId:  int64(txInfo.AssetId),
		AssetAId:       int64(txInfo.AssetId),
		AssetBId:       commonAsset.NilAssetId,
		TxAmount:       0,
		NativeAddress:  "",
		MempoolDetails: nMempoolTxDetails,
		ChainId:        commonTx.L2TxChainId,
		TxInfo:         string(bTxInfo),
		ExtraInfo:      "",
		Memo:           txInfo.Memo,
		L2BlockHeight:  0,
		Status:         0,
	}

	// write into mempool
	err = l.svcCtx.MempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{nMempoolTx})
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateTxMempoolForTranferTx] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	// update mempool state
	go globalmapHandler.UpdateGlobalMap(nMempoolTx)

	return resTxId, nil
}
