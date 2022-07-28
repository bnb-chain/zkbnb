/*
 * Copyright © 2021 Zkbas Protocol
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
	"reflect"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/failtx"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/sysconf"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type SendTransferNftTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
	failtx        failtx.Model
	sysconf       sysconf.Model
}

func NewSendTransferNftTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTransferNftTxLogic {
	return &SendTransferNftTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
		failtx:        failtx.New(svcCtx),
		sysconf:       sysconf.New(svcCtx),
	}
}

func (l *SendTransferNftTxLogic) SendTransferNftTx(in *globalRPCProto.ReqSendTxByRawInfo) (respSendTx *globalRPCProto.RespSendTx, err error) {
	txInfo, err := commonTx.ParseTransferNftTxInfo(in.TxInfo)
	if err != nil {
		logx.Errorf("[ParseTransferNftTxInfo] err:%v", err)
		return nil, err
	}
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return nil, err
	}
	if err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.FromAccountIndex)); err != nil {
		logx.Errorf("[CheckRequestParam] param:%v,err:%v", txInfo.FromAccountIndex, err)
		return nil, err
	}
	if err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.ToAccountIndex)); err != nil {
		logx.Errorf("[CheckRequestParam] param:%v,err:%v", txInfo.ToAccountIndex, err)
		return nil, err
	}
	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		logx.Errorf("[checkGasAccountIndex] err: %v", err)
		return nil, err
	}
	var accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	nftInfo, err := l.commglobalmap.GetLatestNftInfoForRead(l.ctx, txInfo.NftIndex)
	if err != nil {
		logx.Errorf("[GetLatestNftInfoForRead] err:%v", err)
		return nil, err
	}
	accountInfoMap[txInfo.FromAccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to get account info: %s", err.Error())
		return nil, err
	}
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.ToAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferNftTx] unable to get account info: %s", err.Error())
			return nil, err
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("[sendTransferNftTx] invalid account name")
		return nil, err
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferNftTx] unable to get account info: %s", err.Error())
			return nil, err
		}
	}
	if nftInfo.OwnerAccountIndex != txInfo.FromAccountIndex {
		logx.Errorf("[sendTransferNftTx] you're not owner")
		return nil, err
	}
	// check expired at
	if txInfo.ExpiredAt < time.Now().UnixMilli() {
		logx.Errorf("[sendTransferNftTx] invalid time stamp")
		return nil, err
	}
	txDetails, err := txVerification.VerifyTransferNftTxInfo(accountInfoMap, nftInfo, txInfo)
	if err != nil {
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeTransferNft,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		txInfo.NftIndex,
		commonConstant.NilPairIndex,
		commonConstant.NilAssetId,
		commonConstant.NilAssetAmountStr,
		"",
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	respSendTx = &globalRPCProto.RespSendTx{
		TxId: txId,
	}
	if err := l.commglobalmap.DeleteLatestNftInfoForReadInCache(l.ctx, txInfo.NftIndex); err != nil {
		logx.Errorf("[DeleteLatestNftInfoForReadInCache] param:%v, err:%v", txInfo.NftIndex, err)
		return nil, err
	}
	if err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel); err != nil {
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	// update cacke, not key logic
	if err := l.commglobalmap.SetLatestNftInfoForReadInCache(l.ctx, txInfo.NftIndex); err != nil {
		logx.Errorf("[SetLatestNftInfoForReadInCache] param:%v, err:%v", txInfo.NftIndex, err)
	}
	return respSendTx, nil
}

func (l *SendTransferNftTxLogic) createFailTransferNftTx(info *commonTx.TransferNftTxInfo, inputErr error) error {
	txInfo, err := json.Marshal(info)
	if err != nil {
		logx.Errorf("[Marshal] err:%v", err)
		return err
	}
	failTx := &tx.FailTx{
		TxHash:        util.RandomUUID(),
		TxType:        commonTx.TxTypeTransferNft,
		GasFee:        info.GasFeeAssetAmount.String(),
		GasFeeAssetId: info.GasFeeAssetId,
		// tx status, 1 - success(default), 2 - failure
		TxStatus:      tx.StatusFail,
		AssetAId:      commonConstant.NilAssetId,
		AssetBId:      commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: "0x00",
		TxInfo:        string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: inputErr.Error(),
		Memo:      "",
	}
	if err = l.failtx.CreateFailTx(failTx); err != nil {
		logx.Errorf("[CreateFailTx] err:%v", err)
		return err
	}
	return inputErr
}
