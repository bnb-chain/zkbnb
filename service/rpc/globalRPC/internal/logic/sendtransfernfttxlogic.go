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
	"strconv"
	"time"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/failtx"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/sysconf"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

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
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	if err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.ToAccountIndex)); err != nil {
		logx.Errorf("[CheckRequestParam] param:%v,err:%v", txInfo.ToAccountIndex, err)
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	gasAccountIndexConfig, err := l.sysconf.GetSysconfigByName(l.ctx, sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to get sysconfig by name: %s", err.Error())
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return nil, l.createFailTransferNftTx(txInfo, errors.New("[sendTransferNftTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendTransferNftTx] invalid gas account index")
		return nil, l.createFailTransferNftTx(txInfo, errors.New("[sendTransferNftTx] invalid gas account index"))
	}
	var accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	nftInfo, err := globalmapHandler.GetLatestNftInfoForRead(l.svcCtx.NftModel,
		l.svcCtx.MempoolModel, l.svcCtx.RedisConnection, txInfo.NftIndex)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to get nft info")
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	accountInfoMap[txInfo.FromAccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to get account info: %s", err.Error())
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.ToAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferNftTx] unable to get account info: %s", err.Error())
			return nil, l.createFailTransferNftTx(txInfo, err)
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("[sendTransferNftTx] invalid account name")
		return nil, l.createFailTransferNftTx(txInfo, errors.New("[sendTransferNftTx] invalid account name"))
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferNftTx] unable to get account info: %s", err.Error())
			return nil, l.createFailTransferNftTx(txInfo, err)
		}
	}
	if nftInfo.OwnerAccountIndex != txInfo.FromAccountIndex {
		logx.Errorf("[sendTransferNftTx] you're not owner")
		return nil, l.createFailTransferNftTx(txInfo, errors.New("[sendTransferNftTx] you're not owner"))
	}
	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendTransferNftTx] invalid time stamp")
		return nil, l.createFailTransferNftTx(txInfo, errors.New("[sendTransferNftTx] invalid time stamp"))
	}
	var txDetails []*mempool.MempoolTxDetail
	txDetails, err = txVerification.VerifyTransferNftTxInfo(accountInfoMap, nftInfo, txInfo)
	if err != nil {
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	// delete key
	key := util.GetNftKeyForRead(txInfo.NftIndex)
	if _, err = l.svcCtx.RedisConnection.Del(key); err != nil {
		logx.Errorf("[sendTransferNftTx] unable to delete key from redis: %s", err.Error())
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	// write into mempool
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
	if err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel); err != nil {
		return nil, l.createFailTransferNftTx(txInfo, err)
	}
	respSendTx = &globalRPCProto.RespSendTx{
		TxId: txId,
	}
	// update redis
	var formatNftInfo *commonAsset.NftInfo
	for _, txDetail := range mempoolTx.MempoolDetails {
		if txDetail.AssetType == commonAsset.NftAssetType {
			formatNftInfo, err = commonAsset.ParseNftInfo(txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[sendTransferNftTx] unable to parse nft info: %s", err.Error())
				return respSendTx, nil
			}
		}
	}
	nftInfoBytes, err := json.Marshal(formatNftInfo)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to marshal: %s", err.Error())
		return respSendTx, nil
	}
	_ = l.svcCtx.RedisConnection.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)
	return respSendTx, nil
}

func (l *SendTransferNftTxLogic) createFailTransferNftTx(info *commonTx.TransferNftTxInfo, inputErr error) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		logx.Errorf("[Marshal] err:%v", err)
		return err
	}
	failTx := &tx.FailTx{
		TxHash:        txHash,
		TxType:        commonTx.TxTypeTransferNft,
		GasFee:        info.GasFeeAssetAmount.String(),
		GasFeeAssetId: info.GasFeeAssetId,
		// tx status, 1 - success(default), 2 - failure
		TxStatus:      tx.StatusFail,
		AssetAId:      commonConstant.NilAssetId,
		AssetBId:      commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: nativeAddress,
		TxInfo:        string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: err.Error(),
		Memo:      "",
	}
	if err = l.failtx.CreateFailTx(failTx); err != nil {
		logx.Errorf("[CreateFailTx] err:%v", err)
		return err
	}
	return inputErr
}
