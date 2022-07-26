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
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendTransferTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewSendTransferTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTransferTxLogic {
	return &SendTransferTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *SendTransferTxLogic) SendTransferTx(in *globalRPCProto.ReqSendTxByRawInfo) (respSendTx *globalRPCProto.RespSendTx, err error) {
	respSendTx = &globalRPCProto.RespSendTx{}
	txInfo, err := commonTx.ParseTransferTxInfo(in.TxInfo)
	if err != nil {
		logx.Errorf("[ParseTransferTxInfo] err:%v", err)
		return nil, err
	}
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return nil, err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAmount); err != nil {
		logx.Errorf("[CheckRequestParam] param:%v,err:%v", txInfo.AssetAmount, err)
		return nil, err
	}
	if err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.AssetId)); err != nil {
		logx.Errorf("[CheckRequestParam] param:%v,err:%v", txInfo.AssetId, err)
		return nil, err
	}
	if err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.FromAccountIndex)); err != nil {
		logx.Errorf("[CheckRequestParam] param:%v,err:%v", txInfo.FromAccountIndex, err)
		return nil, err
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.ToAccountIndex))
	if err != nil {
		logx.Errorf("[CheckRequestParam] param:%v,err:%v", txInfo.ToAccountIndex, err)
		return nil, err
	}
	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		logx.Errorf("[checkGasAccountIndex] err: %v", err)
		return nil, err
	}
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendTransferTx] invalid time stamp")
		return respSendTx, l.createFailTransferTx(txInfo, errors.New("[sendTransferTx] invalid time stamp"))
	}
	var accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	accountInfoMap[txInfo.FromAccountIndex], err = l.commglobalmap.GetLatestAccountInfoWithCache(l.ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
		return respSendTx, l.createFailTransferTx(txInfo, err)
	}
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = l.commglobalmap.GetBasicAccountInfoWithCache(l.ctx, txInfo.ToAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
			return respSendTx, l.createFailTransferTx(txInfo, err)
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("[sendTransferTx] invalid account name")
		return respSendTx, l.createFailTransferTx(txInfo, errors.New("[sendTransferTx] invalid account name"))
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = l.commglobalmap.GetBasicAccountInfoWithCache(l.ctx, txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
			return respSendTx, l.createFailTransferTx(txInfo, err)
		}
	}
	var txDetails []*mempool.MempoolTxDetail
	txDetails, err = txVerification.VerifyTransferTxInfo(accountInfoMap, txInfo)
	if err != nil {
		return respSendTx, l.createFailTransferTx(txInfo, err)
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return respSendTx, l.createFailTransferTx(txInfo, err)
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
	respSendTx.TxId = txId
	if err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel); err != nil {
		return respSendTx, l.createFailTransferTx(txInfo, err)
	}
	if err := l.commglobalmap.SetLatestAccountInfoInToCache(l.ctx, txInfo.FromAccountIndex); err != nil {
		logx.Errorf("[SetLatestAccountInfoInToCache] unable to set account info in cache: %v", err)
	}
	if err := l.commglobalmap.SetLatestAccountInfoInToCache(l.ctx, txInfo.ToAccountIndex); err != nil {
		logx.Errorf("[SetLatestAccountInfoInToCache] unable to set account info in cache: %v", err)
	}
	return respSendTx, nil
}

func (l *SendTransferTxLogic) createFailTransferTx(info *commonTx.TransferTxInfo, inputErr error) error {
	txInfo, err := json.Marshal(info)
	if err != nil {
		logx.Errorf("[Marshal] err:%v", err)
		return err
	}
	failTx := &tx.FailTx{
		TxHash:        util.RandomUUID(),
		TxType:        commonTx.TxTypeTransfer,
		GasFee:        info.GasFeeAssetAmount.String(),
		GasFeeAssetId: info.AssetId,
		TxStatus:      tx.StatusFail,
		AssetAId:      info.AssetId,
		AssetBId:      commonConstant.NilAssetId,
		TxAmount:      info.AssetAmount.String(),
		NativeAddress: "0x00",
		TxInfo:        string(txInfo),
		ExtraInfo:     inputErr.Error(),
		Memo:          info.Memo,
	}
	if err = l.svcCtx.FailTxModel.CreateFailTx(failTx); err != nil {
		logx.Errorf("[CreateFailTx] err:%v", err)
		return err
	}
	return inputErr
}
