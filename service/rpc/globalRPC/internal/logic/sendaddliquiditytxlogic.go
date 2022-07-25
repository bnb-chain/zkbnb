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
	"math/big"
	"strconv"
	"time"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendAddLiquidityTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewSendAddLiquidityTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAddLiquidityTxLogic {
	return &SendAddLiquidityTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *SendAddLiquidityTxLogic) SendAddLiquidityTx(reqSendTx *globalRPCProto.ReqSendTxByRawInfo) (respSendTx *globalRPCProto.RespSendTx, err error) {
	respSendTx = &globalRPCProto.RespSendTx{}
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(reqSendTx.TxInfo)
	if err != nil {
		logx.Errorf("[ParseAddLiquidityTxInfo] param:%v,err:%v", reqSendTx.TxInfo, err)
		return nil, err
	}
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return nil, err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAAmount); err != nil {
		logx.Errorf("[CheckPackedAmount] param:%v,err:%v", txInfo.AssetAAmount, err)
		return nil, err
	}
	if err := util.CheckPackedAmount(txInfo.AssetBAmount); err != nil {
		logx.Errorf("[CheckPackedAmount] param:%v,err:%v", txInfo.AssetBAmount, err)
		return nil, err
	}
	// check gas account index
	if err := l.checkGasAccountIndex(txInfo.GasAccountIndex); err != nil {
		logx.Errorf("[checkGasAccountIndex] err: %v", err)
		return nil, err
	}
	// check expired at
	if err := l.checkExpiredAt(txInfo.ExpiredAt); err != nil {
		logx.Errorf("[sendWithdrawTx] invalid time stamp")
		return nil, err
	}
	liquidityInfo, err := l.commglobalmap.GetLatestLiquidityInfoForWrite(l.ctx, txInfo.PairIndex)
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForWrite] err: %v", err)
		return nil, err
	}
	if liquidityInfo.AssetA == nil || liquidityInfo.AssetB == nil {
		logx.Errorf("[ErrInvalidLiquidityAsset]")
		return nil, errcode.ErrInvalidLiquidityAsset
	}
	var accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	if liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 {
		txInfo.LpAmount, err = util.ComputeEmptyLpAmount(txInfo.AssetAAmount, txInfo.AssetBAmount)
		if err != nil {
			logx.Errorf("[ComputeEmptyLpAmount] : %v", err)
			return nil, err
		}
	} else {
		txInfo.LpAmount, err = util.ComputeLpAmount(liquidityInfo, txInfo.AssetAAmount)
		if err != nil {
			return nil, err
		}
	}
	if accountInfoMap[txInfo.FromAccountIndex] == nil {
		accountInfoMap[txInfo.FromAccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.FromAccountIndex)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] param:%v,err:%v", txInfo.FromAccountIndex, err)
			return nil, err
		}
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] param:%v,err:%v", txInfo.GasAccountIndex, err)
			return nil, err
		}
	}
	if accountInfoMap[liquidityInfo.TreasuryAccountIndex] == nil {
		accountInfoMap[liquidityInfo.TreasuryAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, liquidityInfo.TreasuryAccountIndex)
		if err != nil {
			logx.Errorf("[GetLatestAccountInfo] param:%v,err:%v", liquidityInfo.TreasuryAccountIndex, err)
			return nil, err
		}
	}
	txDetails, err := txVerification.VerifyAddLiquidityTxInfo(accountInfoMap, liquidityInfo, txInfo)
	if err != nil {
		logx.Errorf("[VerifyAddLiquidityTxInfo] param:%v, err:%v", txInfo, err)
		return nil, l.createFailAddLiquidityTx(txInfo, err)
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return nil, err
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeAddLiquidity,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		txInfo.PairIndex,
		commonConstant.NilAssetId,
		txInfo.LpAmount.String(),
		"",
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	respSendTx.TxId = txId
	if err := l.commglobalmap.DeleteLatestLiquidityInfoForWriteInCache(l.ctx, txInfo.PairIndex); err != nil {
		logx.Errorf("[DeleteLatestLiquidityInfoForWriteInCache] param:%v, err:%v", txInfo.PairIndex, err)
		return nil, err
	}
	if err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel); err != nil {
		logx.Errorf("[CreateMempoolTx] param:%v, err:%v", mempoolTx, err)
		return nil, err
	}
	// update cacke, not key logic
	if err := l.commglobalmap.SetLatestLiquidityInfoForWrite(l.ctx, txInfo.PairIndex); err != nil {
		logx.Errorf("[SetLatestLiquidityInfoForWrite] param:%v, err:%v", txInfo.PairIndex, err)
	}
	return respSendTx, nil
}

func (l *SendAddLiquidityTxLogic) createFailAddLiquidityTx(info *commonTx.AddLiquidityTxInfo, errInput error) error {
	txInfo, err := json.Marshal(info)
	if err != nil {
		logx.Errorf("[createFailAddLiquidityTx] Marshal param:%v, err:%v", txInfo, err)
		return errcode.ErrMarshal.RefineError(err)
	}
	failTx := &tx.FailTx{
		TxHash:        util.RandomUUID(),
		TxType:        int64(commonTx.TxTypeAddLiquidity),
		GasFee:        info.GasFeeAssetAmount.String(),
		GasFeeAssetId: info.GasFeeAssetId,
		TxStatus:      TxFail,
		AssetAId:      info.AssetAId,
		AssetBId:      info.AssetBId,
		TxAmount:      info.AssetAAmount.String(),
		NativeAddress: "0x00",
		TxInfo:        string(txInfo),
		ExtraInfo:     errInput.Error(),
	}
	if err = l.svcCtx.FailTxModel.CreateFailTx(failTx); err != nil {
		return errcode.ErrCreateFailTx.RefineError(err)
	}
	return errInput
}

func (l *SendAddLiquidityTxLogic) checkGasAccountIndex(txGasAccountIndex int64) error {
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[GetSysconfigByName] err: %v", err)
		return err
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("[ParseInt] param:%v,err:%v", gasAccountIndexConfig.Value, err)
		return err
	}
	if gasAccountIndex != txGasAccountIndex {
		logx.Errorf("[ParseInt] param:%v, txGasAccountIndex:%v, err:%v", gasAccountIndex, txGasAccountIndex, err)
		return errcode.ErrInvalidGasAccountIndex
	}
	return nil
}

func (l *SendAddLiquidityTxLogic) checkExpiredAt(expiredAt int64) error {
	now := time.Now().UnixMilli()
	if expiredAt < now {
		logx.Errorf("[sendWithdrawTx] invalid time stamp,expiredAt:%v,now:%v", expiredAt, now)
		return errcode.ErrInvalidExpiredAt
	}
	return nil
}
