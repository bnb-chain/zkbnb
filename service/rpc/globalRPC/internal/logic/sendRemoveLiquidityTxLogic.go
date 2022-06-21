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
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"math/big"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *SendTxLogic) sendRemoveLiquidityTx(rawTxInfo string) (txId string, err error) {

	// parse removeliquidity tx info
	txInfo, err := commonTx.ParseRemoveLiquidityTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendRemoveLiquidityTx] => [commonTx.ParseRemoveLiquidityTxInfo] : %s. invalid rawTxInfo %s",
			err.Error(), rawTxInfo)
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	// check gas account index
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendRemoveLiquidityTx] unable to get sysconfig by name: %s", err.Error())
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, errors.New("[sendRemoveLiquidityTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendRemoveLiquidityTx] invalid gas account index")
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, errors.New("[sendRemoveLiquidityTx] invalid gas account index"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendRemoveLiquidityTx] invalid time stamp")
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, errors.New("[sendRemoveLiquidityTx] invalid time stamp"))
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
		logx.Errorf("[sendRemoveLiquidityTx] unable to get latest liquidity info for write: %s", err.Error())
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, err)
	}
	defer redisLock.Release()

	// check params
	if liquidityInfo.AssetA == nil ||
		liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidityInfo.AssetB == nil ||
		liquidityInfo.AssetB.Cmp(big.NewInt(0)) == 0 ||
		liquidityInfo.LpAmount == nil ||
		liquidityInfo.LpAmount.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("[sendRemoveLiquidityTx] invalid params")
		return "", errors.New("[sendRemoveLiquidityTx] invalid params")
	}

	var (
		assetAAmount, assetBAmount *big.Int
	)
	assetAAmount, assetBAmount = util.ComputeRemoveLiquidityAmount(
		liquidityInfo,
		txInfo.LpAmount,
	)
	if err != nil {
		logx.Errorf("[sendRemoveLiquidityTx] unable to compute lp portion: %s", err.Error())
		return "", err
	}
	if assetAAmount.Cmp(txInfo.AssetAMinAmount) < 0 || assetBAmount.Cmp(txInfo.AssetBMinAmount) < 0 {
		errInfo := fmt.Sprintf("[logic.sendRemoveLiquidityTx] less than MinDelta: %s:%s/%s:%s",
			txInfo.AssetAMinAmount.String(), txInfo.AssetBMinAmount.String(), assetAAmount.String(), assetBAmount.String())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	// add into tx info
	txInfo.AssetAAmountDelta = assetAAmount
	txInfo.AssetBAmountDelta = assetBAmount

	// get latest account info for from account index
	if accountInfoMap[txInfo.FromAccountIndex] == nil {
		accountInfoMap[txInfo.FromAccountIndex], err = globalmapHandler.GetLatestAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.MempoolModel,
			l.svcCtx.RedisConnection,
			txInfo.FromAccountIndex,
		)
		if err != nil {
			logx.Errorf("[sendRemoveLiquidityTx] unable to get latest account info: %s", err.Error())
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
			logx.Errorf("[sendRemoveLiquidityTx] unable to get latest account info: %s", err.Error())
			return "", err
		}
	}
	if accountInfoMap[liquidityInfo.TreasuryAccountIndex] == nil {
		accountInfoMap[liquidityInfo.TreasuryAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			liquidityInfo.TreasuryAccountIndex,
		)
		if err != nil {
			logx.Errorf("[sendRemoveLiquidityTx] unable to get latest account info: %s", err.Error())
			return "", err
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify RemoveLiquidity tx
	txDetails, err = txVerification.VerifyRemoveLiquidityTxInfo(
		accountInfoMap,
		liquidityInfo,
		txInfo)
	if err != nil {
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, err)
	}

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeRemoveLiquidity,
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
		logx.Errorf("[sendRemoveLiquidityTx] unable to delete key from redis: %s", err.Error())
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, err)
	}
	_, err = l.svcCtx.RedisConnection.Del(key2)
	if err != nil {
		logx.Errorf("[sendRemoveLiquidityTx] unable to delete key from redis: %s", err.Error())
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, err)
	}
	// insert into mempool
	err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return "", l.HandleCreateFailRemoveLiquidityTx(txInfo, err)
	}
	// update redis
	// get latest liquidity info
	for _, txDetail := range txDetails {
		if txDetail.AssetType == commonAsset.LiquidityAssetType {
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[sendAddLiquidityTx] unable to compute new balance: %s", err.Error())
				return txId, nil
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				logx.Errorf("[sendAddLiquidityTx] unable to parse liquidity info: %s", err.Error())
				return txId, nil
			}
		}
	}
	liquidityInfoBytes, err := json.Marshal(liquidityInfo)
	if err != nil {
		logx.Errorf("[sendRemoveLiquidityTx] unable to marshal: %s", err.Error())
		return txId, nil
	}
	_ = l.svcCtx.RedisConnection.Setex(key, string(liquidityInfoBytes), globalmapHandler.LiquidityExpiryTime)

	return txId, nil
}

func (l *SendTxLogic) HandleCreateFailRemoveLiquidityTx(txInfo *commonTx.RemoveLiquidityTxInfo, err error) error {
	errCreate := l.CreateFailRemoveLiquidityTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendremoveliquiditytxlogic.HandleCreateFailRemoveLiquidityTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendremoveliquiditytxlogic.HandleCreateFailRemoveLiquidityTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendTxLogic) CreateFailRemoveLiquidityTx(info *commonTx.RemoveLiquidityTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	txFeeAssetId := info.GasFeeAssetId

	assetAId := info.AssetAId
	assetBId := info.AssetBId
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailRemoveLiquidityTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeRemoveLiquidity,
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
		TxAmount: info.LpAmount.String(),
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
	}

	err = l.svcCtx.FailTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailRemoveLiquidityTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
