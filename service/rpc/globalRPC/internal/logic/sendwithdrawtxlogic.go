/*
 *
 *  * Copyright © 2021 Zecrey Protocol
 *  *
 *  * Licensed under the Apache License, Version 2.0 (the "License");
 *  * you may not use this file except in compliance with the License.
 *  * You may obtain a copy of the License at
 *  *
 *  *     http://www.apache.org/licenses/LICENSE-2.0
 *  *
 *  * Unless required by applicable law or agreed to in writing, software
 *  * distributed under the License is distributed on an "AS IS" BASIS,
 *  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  * See the License for the specific language governing permissions and
 *  * limitations under the License.
 *
 */

package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendWithdrawTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewSendWithdrawTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWithdrawTxLogic {
	return &SendWithdrawTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}
func (l *SendWithdrawTxLogic) SendWithdrawTx(in *globalRPCProto.ReqSendTxByRawInfo) (respSendTx *globalRPCProto.RespSendTx, err error) {
	rawTxInfo := in.TxInfo
	respSendTx = &globalRPCProto.RespSendTx{}
	// parse withdraw tx info
	txInfo, err := commonTx.ParseWithdrawTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendWithdrawTx.ParseWithdrawTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return respSendTx, errors.New(errInfo)
	}
	/*
		Check Params
	*/
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return respSendTx, err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetAmount, err)
		return respSendTx, err
	}
	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.AssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendWithdrawTx] err: invalid assetId %v", txInfo.AssetId)
		return respSendTx, l.HandleCreateFailWithdrawTx(txInfo, errors.New(errInfo))
	}

	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.GasFeeAssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendWithdrawTx] err: invalid gasFeeAssetId %v", txInfo.GasFeeAssetId)
		return respSendTx, l.HandleCreateFailWithdrawTx(txInfo, errors.New(errInfo))
	}
	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		logx.Errorf("[checkGasAccountIndex] err: %v", err)
		return nil, err
	}
	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendWithdrawTx] invalid time stamp")
		return respSendTx, l.HandleCreateFailWithdrawTx(txInfo, errors.New("[sendWithdrawTx] invalid time stamp"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.FromAccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[sendWithdrawTx] unable to get account info: %s", err.Error())
		return respSendTx, l.HandleCreateFailWithdrawTx(txInfo, err)
	}
	// get account info by gas index
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		// get account info by gas index
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendWithdrawTx] unable to get account info: %s", err.Error())
			return respSendTx, l.HandleCreateFailWithdrawTx(txInfo, err)
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	/*
		Get txDetails
	*/
	// verify withdraw tx
	txDetails, err = txVerification.VerifyWithdrawTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return respSendTx, l.HandleCreateFailWithdrawTx(txInfo, err)
	}

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return respSendTx, l.HandleCreateFailWithdrawTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeWithdraw,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		commonConstant.NilPairIndex,
		txInfo.AssetId,
		txInfo.AssetAmount.String(),
		txInfo.ToAddress,
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return respSendTx, l.HandleCreateFailWithdrawTx(txInfo, err)
	}
	respSendTx.TxId = txId
	return respSendTx, nil
}

func (l *SendWithdrawTxLogic) HandleCreateFailWithdrawTx(txInfo *commonTx.WithdrawTxInfo, err error) error {
	errCreate := l.CreateFailWithdrawTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Errorf("[sendwithdrawtxlogic.HandleCreateFailWithdrawTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendwithdrawtxlogic.HandleCreateFailWithdrawTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendWithdrawTxLogic) CreateFailWithdrawTx(info *commonTx.WithdrawTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	txFeeAssetId := info.AssetId
	assetId := info.AssetId
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailWithdrawTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeWithdraw,
		// tx fee
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: int64(txFeeAssetId),
		// tx status, 1 - success(default), 2 - failure
		TxStatus: TxFail,
		// l1asset id
		AssetAId: int64(assetId),
		// tx amount
		TxAmount: info.AssetAmount.String(),
		// layer1 address
		NativeAddress: info.ToAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
	}

	err = l.svcCtx.FailTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailWithdrawTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
