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
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"math/big"
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
	"github.com/zeromicro/go-zero/core/stores/redis"

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

func (l *SendAddLiquidityTxLogic) HandleCreateFailAddLiquidityTx(txInfo *commonTx.AddLiquidityTxInfo, err error) error {
	errCreate := l.CreateFailAddLiquidityTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendaddliquiditytxlogic.HandleFailAddLiquidityTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendaddliquiditytxlogic.HandleFailAddLiquidityTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendAddLiquidityTxLogic) SendAddLiquidityTx(reqSendTx *globalRPCProto.ReqSendTxByRawInfo) (respSendTx *globalRPCProto.RespSendTx, err error) {
	respSendTx = &globalRPCProto.RespSendTx{}
	rawTxInfo := reqSendTx.TxInfo
	// parse addliquidity tx info
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendAddLiquidityTx] => [commonTx.ParseAddLiquidityTxInfo] : %s. invalid rawTxInfo %s",
			err.Error(), rawTxInfo)
		logx.Error(errInfo)
		return respSendTx, errors.New(errInfo)
	}

	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return respSendTx, err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetAAmount, err)
		return respSendTx, err
	}
	if err := util.CheckPackedAmount(txInfo.AssetBAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetBAmount, err)
		return respSendTx, err
	}
	if err := util.CheckPackedAmount(txInfo.LpAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.LpAmount, err)
		return respSendTx, err
	}
	if err := util.CheckPackedAmount(txInfo.TreasuryAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.TreasuryAmount, err)
		return respSendTx, err
	}
	if err := util.CheckPackedAmount(txInfo.KLast); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.KLast, err)
		return respSendTx, err
	}
	l.commglobalmap.DeleteLatestAccountInfoInCache(l.ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[DeleteLatestAccountInfoInCache] err:%v", err)
	}
	// check gas account index
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendAddLiquidityTx] unable to get sysconfig by name: %s", err.Error())
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, errors.New("[sendAddLiquidityTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendAddLiquidityTx] invalid gas account index")
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, errors.New("[sendAddLiquidityTx] invalid gas account index"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendWithdrawTx] invalid time stamp")
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, errors.New("[sendWithdrawTx] invalid time stamp"))
	}

	var (
		redisLock      *redis.RedisLock
		liquidityInfo  *commonAsset.LiquidityInfo
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)

	redisLock, liquidityInfo, err = globalmapHandler.GetLatestLiquidityInfoForWrite(
		l.svcCtx.LiquidityModel,
		l.svcCtx.MempoolModel,
		l.svcCtx.RedisConnection,
		txInfo.PairIndex,
	)
	if err != nil {
		logx.Errorf("[sendAddLiquidityTx] unable to get latest liquidity info for write: %s", err.Error())
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	defer redisLock.Release()

	// check params
	if liquidityInfo.AssetA == nil ||
		liquidityInfo.AssetB == nil {
		logx.Errorf("[sendAddLiquidityTx] invalid params")
		return respSendTx, errors.New("[sendAddLiquidityTx] invalid params")
	}

	var (
		lpAmount *big.Int
	)
	if liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 {
		lpAmount, err = util.ComputeEmptyLpAmount(txInfo.AssetAAmount, txInfo.AssetBAmount)
		if err != nil {
			logx.Errorf("[sendAddLiquidityTx] unable to compute lp amount: %s", err.Error())
			return respSendTx, err
		}
	} else {
		lpAmount, err = util.ComputeLpAmount(liquidityInfo, txInfo.AssetAAmount)
		if err != nil {
			return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, err)
		}
	}
	// add into tx info
	txInfo.LpAmount = lpAmount
	// get latest account info for from account index
	if accountInfoMap[txInfo.FromAccountIndex] == nil {
		accountInfoMap[txInfo.FromAccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.FromAccountIndex)
		if err != nil {
			logx.Errorf("[sendAddLiquidityTx] unable to get latest account info: %s", err.Error())
			return respSendTx, err
		}
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.GasAccountIndex,
		)
		if err != nil {
			logx.Errorf("[sendAddLiquidityTx] unable to get latest account info: %s", err.Error())
			return respSendTx, err
		}
	}
	if accountInfoMap[liquidityInfo.TreasuryAccountIndex] == nil {
		accountInfoMap[liquidityInfo.TreasuryAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			liquidityInfo.TreasuryAccountIndex,
		)
		if err != nil {
			logx.Errorf("[sendAddLiquidityTx] unable to get latest account info: %s", err.Error())
			return respSendTx, err
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify addLiquidity tx
	txDetails, err = txVerification.VerifyAddLiquidityTxInfo(
		accountInfoMap,
		liquidityInfo,
		txInfo)

	if err != nil {
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, err)
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
	// delete key
	key := util.GetLiquidityKeyForWrite(txInfo.PairIndex)
	key2 := util.GetLiquidityKeyForRead(txInfo.PairIndex)
	_, err = l.svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("[sendAddLiquidityTx] unable to delete key from redis: %s", err.Error())
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	_, err = l.svcCtx.RedisConnection.Del(key2)
	if err != nil {
		logx.Errorf("[sendAddLiquidityTx] unable to delete key from redis: %s", err.Error())
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	// insert into mempool
	err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return respSendTx, l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	respSendTx.TxId = txId
	// update redis
	// get latest liquidity info
	for _, txDetail := range txDetails {
		if txDetail.AssetType == commonAsset.LiquidityAssetType {
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[sendAddLiquidityTx] unable to compute new balance: %s", err.Error())
				return respSendTx, nil
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				logx.Errorf("[sendAddLiquidityTx] unable to parse liquidity info: %s", err.Error())
				return respSendTx, nil
			}
		}
	}
	liquidityInfoBytes, err := json.Marshal(liquidityInfo)
	if err != nil {
		logx.Errorf("[sendAddLiquidityTx] unable to marshal: %s", err.Error())
		return respSendTx, nil
	}
	_ = l.svcCtx.RedisConnection.Setex(key, string(liquidityInfoBytes), globalmapHandler.LiquidityExpiryTime)
	return respSendTx, nil
}

func (l *SendAddLiquidityTxLogic) CreateFailAddLiquidityTx(info *commonTx.AddLiquidityTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	txType := int64(commonTx.TxTypeAddLiquidity)
	txFeeAssetId := info.GasFeeAssetId

	assetAId := info.AssetAId
	assetBId := info.AssetBId
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[SendAddLiquidityTxLogic.CreateFailAddLiquidityTx] %s", err.Error())
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
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: txFeeAssetId,
		// tx status, 1 - success(default), 2 - failure
		TxStatus: TxFail,
		// AssetAId
		AssetAId: assetAId,
		// l1asset id
		AssetBId: assetBId,
		// tx amount
		TxAmount: info.AssetAAmount.String(),
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
	}

	err = l.svcCtx.FailTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[SendAddLiquidityTxLogic.CreateFailAddLiquidityTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
