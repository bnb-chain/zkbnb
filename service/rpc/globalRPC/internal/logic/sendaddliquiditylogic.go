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
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/txHandler"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"math/big"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *SendTxLogic) HandleCreateFailAddLiquidityTx(txInfo *commonTx.AddLiquidityTxInfo, err error) error {
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

func (l *SendTxLogic) sendAddLiquidityTx(rawTxInfo string) (txId string, err error) {
	// parse addliquidity tx info
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendAddLiquidityTx] => [commonTx.ParseAddLiquidityTxInfo] : %s. invalid rawTxInfo %s",
			err.Error(), rawTxInfo)
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	// check gas account index
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferTx] unable to get sysconfig by name: %s", err.Error())
		return "", l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", l.HandleCreateFailAddLiquidityTx(txInfo, errors.New("[sendTransferTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendTransferTx] invalid gas account index")
		return "", l.HandleCreateFailAddLiquidityTx(txInfo, errors.New("[sendTransferTx] invalid gas account index"))
	}

	var (
		redisLock      *redis.RedisLock
		liquidityInfo  *liquidity.Liquidity
		accountInfoMap = make(map[int64]*commonAsset.FormatAccountInfo)
	)

	redisLock, liquidityInfo, err = globalmapHandler.GetLatestLiquidityInfoForWrite(
		l.svcCtx.LiquidityModel,
		l.svcCtx.LiquidityHistoryModel,
		l.svcCtx.RedisConnection,
		txInfo.PairIndex,
	)
	if err != nil {
		logx.Errorf("[sendAddLiquidityTx] unable to get latest liquidity info for write: %s", err.Error())
		return "", l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	defer redisLock.Release()

	// check params
	if liquidityInfo.AssetA == "" ||
		liquidityInfo.AssetA == util.ZeroBigInt.String() ||
		liquidityInfo.AssetB == "" ||
		liquidityInfo.AssetB == util.ZeroBigInt.String() {
		logx.Errorf("[sendAddLiquidityTx] invalid params")
		return "", errors.New("[sendAddLiquidityTx] invalid params")
	}

	var (
		lpAmount *big.Int
	)
	poolABalance, isValid := new(big.Int).SetString(liquidityInfo.AssetA, 10)
	if !isValid {
		logx.Errorf("[sendAddLiquidityTx] unable to parse amount")
		return "", errors.New("[sendAddLiquidityTx] unable to parse amount")
	}
	poolBBalance, isValid := new(big.Int).SetString(liquidityInfo.AssetA, 10)
	if !isValid {
		logx.Errorf("[sendAddLiquidityTx] unable to parse amount")
		return "", errors.New("[sendAddLiquidityTx] unable to parse amount")
	}
	lpAmount, err = util.ComputeLpAmount(txInfo.AssetAAmount, txInfo.AssetBAmount)
	if err != nil {
		logx.Errorf("[sendAddLiquidityTx] unable to compute lp amount: %s", err.Error())
		return "", err
	}
	// add into tx info
	txInfo.LpAmount = lpAmount

	if liquidityInfo.AssetAId == txInfo.AssetAId &&
		liquidityInfo.AssetBId == txInfo.AssetBId {
		txInfo.PoolAAmount = poolABalance
		txInfo.PoolBAmount = poolBBalance
	} else if liquidityInfo.AssetAId == txInfo.AssetBId &&
		liquidityInfo.AssetBId == txInfo.AssetAId {
		txInfo.PoolAAmount = poolBBalance
		txInfo.PoolBAmount = poolABalance
	} else {
		logx.Errorf("[sendAddLiquidityTx] invalid pair index")
		return "", errors.New("[sendAddLiquidityTx] invalid pair index")
	}

	// get latest account info for from account index
	if accountInfoMap[txInfo.FromAccountIndex] == nil {
		accountInfoMap[txInfo.FromAccountIndex], err = globalmapHandler.GetLatestAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.AccountHistoryModel,
			l.svcCtx.MempoolModel,
			l.svcCtx.MempoolDetailModel,
			l.svcCtx.RedisConnection,
			txInfo.FromAccountIndex,
		)
		if err != nil {
			logx.Errorf("[sendAddLiquidityTx] unable to get latest account info: %s", err.Error())
			return "", err
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
			return "", err
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
		return "", l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeAddLiquidity,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		txInfo.AssetAId,
		txInfo.AssetBId,
		txInfo.LpAmount.String(),
		"",
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txDetails,
	)
	// delete key
	key := util.GetLiquidityKeyForWrite(txInfo.PairIndex)
	_, err = l.svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("[sendAddLiquidityTx] unable to delete key from redis: %s", err.Error())
		return "", l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	// insert into mempool
	err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return "", l.HandleCreateFailAddLiquidityTx(txInfo, err)
	}
	// update redis
	// get latest liquidity info
	for _, txDetail := range txDetails {
		if txDetail.AssetType == commonAsset.LiquidityAssetType {
			poolDelta, err := util.ParsePoolInfo(txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[sendAddLiquidityTx] unable to parse pool info: %s", err.Error())
				return txId, nil
			}
			poolA, isValid := new(big.Int).SetString(liquidityInfo.AssetA, 10)
			if !isValid {
				logx.Errorf("[sendAddLiquidityTx] unable to parse big int: %s", err.Error())
				return txId, nil
			}
			poolB, isValid := new(big.Int).SetString(liquidityInfo.AssetB, 10)
			if !isValid {
				logx.Errorf("[sendAddLiquidityTx] unable to parse big int: %s", err.Error())
				return txId, nil
			}
			liquidityInfo.AssetA = ffmath.Add(poolA, poolDelta.AssetAAmount).String()
			liquidityInfo.AssetB = ffmath.Add(poolB, poolDelta.AssetBAmount).String()
		}
	}
	liquidityInfoBytes, err := json.Marshal(liquidityInfo)
	if err != nil {
		logx.Errorf("[sendAddLiquidityTx] unable to marshal: %s", err.Error())
		return txId, nil
	}
	_ = l.svcCtx.RedisConnection.Setex(key, string(liquidityInfoBytes), globalmapHandler.LiquidityExpiryTime)
	return txId, nil
}

func (l *SendTxLogic) CreateFailAddLiquidityTx(info *commonTx.AddLiquidityTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	txType := int64(commonTx.TxTypeAddLiquidity)
	txFeeAssetId := info.GasFeeAssetId

	assetAId := info.AssetAId
	assetBId := info.AssetBId
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailAddLiquidityTx] %s", err.Error())
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
		TxStatus: txHandler.TxFail,
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
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailAddLiquidityTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
