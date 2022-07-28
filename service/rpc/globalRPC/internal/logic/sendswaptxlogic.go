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
	"errors"
	"fmt"
	"math/big"
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
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type SendSwapTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewSendSwapTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSwapTxLogic {
	return &SendSwapTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *SendSwapTxLogic) SendSwapTx(in *globalRPCProto.ReqSendTxByRawInfo) (respSendTx *globalRPCProto.RespSendTx, err error) {
	respSendTx = &globalRPCProto.RespSendTx{}
	txInfo, err := commonTx.ParseSwapTxInfo(in.TxInfo)
	if err != nil {
		logx.Errorf("[ParseSwapTxInfo] err:%v", err)
		return nil, err
	}
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return nil, err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetAAmount, err)
		return nil, err
	}
	if err := util.CheckPackedAmount(txInfo.AssetBMinAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetBMinAmount, err)
		return nil, err
	}
	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.AssetAId))
	if err != nil {
		logx.Errorf("[CheckRequestParam] param:%v,err:%v", txInfo.AssetAId, err)
		return nil, err
	}
	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		logx.Errorf("[checkGasAccountIndex] err: %v", err)
		return nil, err
	}
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendSwapTx] invalid time stamp")
		return respSendTx, l.createFailSwapTx(txInfo, errors.New("[sendSwapTx] invalid time stamp"))
	}
	liquidityInfo, err := l.commglobalmap.GetLatestLiquidityInfoForWrite(l.ctx, txInfo.PairIndex)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to get latest liquidity info for write: %s", err.Error())
		return respSendTx, l.createFailSwapTx(txInfo, err)
	}
	if liquidityInfo.AssetA == nil || liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidityInfo.AssetB == nil || liquidityInfo.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("[sendSwapTx] invalid params")
		return respSendTx, l.createFailSwapTx(txInfo, errors.New("[sendSwapTx] invalid params"))
	}
	var toDelta *big.Int
	if liquidityInfo.AssetAId == txInfo.AssetAId &&
		liquidityInfo.AssetBId == txInfo.AssetBId {
		toDelta, _, err = util.ComputeDelta(
			liquidityInfo.AssetA,
			liquidityInfo.AssetB,
			liquidityInfo.AssetAId,
			liquidityInfo.AssetBId,
			txInfo.AssetAId,
			true,
			txInfo.AssetAAmount,
			liquidityInfo.FeeRate,
		)
	} else if liquidityInfo.AssetAId == txInfo.AssetBId &&
		liquidityInfo.AssetBId == txInfo.AssetAId {
		toDelta, _, err = util.ComputeDelta(
			liquidityInfo.AssetA,
			liquidityInfo.AssetB,
			liquidityInfo.AssetAId,
			liquidityInfo.AssetBId,
			txInfo.AssetBId,
			true,
			txInfo.AssetAAmount,
			liquidityInfo.FeeRate,
		)
	} else {
		err = errors.New("invalid pair assetIds")
	}
	if err != nil {
		errInfo := fmt.Sprintf("[logic.sendSwapTx] => [util.ComputeDelta]: %s. invalid AssetId: %v/%v/%v",
			err.Error(), txInfo.AssetAId, uint32(liquidityInfo.AssetAId), uint32(liquidityInfo.AssetBId))
		logx.Error(errInfo)
		return respSendTx, errors.New(errInfo)
	}
	// check if toDelta is less than minToAmount
	if toDelta.Cmp(txInfo.AssetBMinAmount) < 0 {
		errInfo := fmt.Sprintf("[logic.sendSwapTx] => minToAmount is bigger than toDelta: %s/%s",
			txInfo.AssetBMinAmount.String(), toDelta.String())
		logx.Error(errInfo)
		return respSendTx, errors.New(errInfo)
	}
	txInfo.AssetBAmountDelta = toDelta
	var accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	if accountInfoMap[txInfo.FromAccountIndex] == nil {
		accountInfoMap[txInfo.FromAccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.FromAccountIndex)
		if err != nil {
			logx.Errorf("[sendSwapTx] unable to get latest account info: %s", err.Error())
			return respSendTx, err
		}
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendSwapTx] unable to get latest account info: %s", err.Error())
			return respSendTx, err
		}
	}
	txDetails, err := txVerification.VerifySwapTxInfo(accountInfoMap, liquidityInfo, txInfo)
	if err != nil {
		logx.Errorf("[VerifySwapTxInfo] err: %v", err)
		return respSendTx, err
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return respSendTx, l.createFailSwapTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeSwap,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		txInfo.PairIndex,
		commonConstant.NilAssetId,
		txInfo.AssetAAmount.String(),
		"",
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	respSendTx.TxId = txId
	if err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel); err != nil {
		return respSendTx, l.createFailSwapTx(txInfo, err)
	}
	// TODO: update GetLatestLiquidityInfoForWrite cache
	return respSendTx, nil
}

func (l *SendSwapTxLogic) createFailSwapTx(info *commonTx.SwapTxInfo, inputErr error) error {
	txHash := util.RandomUUID()
	txFeeAssetId := info.GasFeeAssetId
	assetAId := info.AssetAId
	assetBId := info.AssetBId
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[SendSwapTxLogic.CreateFailSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	failTx := &tx.FailTx{
		TxHash:        txHash,
		TxType:        commonTx.TxTypeSwap,
		GasFee:        info.GasFeeAssetAmount.String(),
		GasFeeAssetId: int64(txFeeAssetId),
		TxStatus:      TxFail,
		AssetAId:      int64(assetAId),
		AssetBId:      int64(assetBId),
		TxAmount:      util.ZeroBigInt.String(),
		NativeAddress: nativeAddress,
		TxInfo:        string(txInfo),
		ExtraInfo:     inputErr.Error(),
	}
	if err = l.svcCtx.FailTxModel.CreateFailTx(failTx); err != nil {
		errInfo := fmt.Sprintf("[SendSwapTxLogic.CreateFailSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return inputErr
}
