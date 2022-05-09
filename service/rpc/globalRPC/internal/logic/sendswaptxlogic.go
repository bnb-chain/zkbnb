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
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/txHandler"
	"math/big"
	"reflect"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *SendTxLogic) sendSwapTx(rawTxInfo string) (txId string, err error) {
	// parse swap tx info
	txInfo, err := commonTx.ParseSwapTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendSwapTx.ParseSwapTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	/*
		Check Params
	*/
	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.AssetAId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendSwapTx] err: invalid assetAId %v", txInfo.AssetAId)
		return "", l.HandleCreateFailSwapTx(txInfo, errors.New(errInfo))
	}

	// check gas account index
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferTx] unable to get sysconfig by name: %s", err.Error())
		return "", l.HandleCreateFailSwapTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", l.HandleCreateFailSwapTx(txInfo, errors.New("[sendTransferTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendTransferTx] invalid gas account index")
		return "", l.HandleCreateFailSwapTx(txInfo, errors.New("[sendTransferTx] invalid gas account index"))
	}

	// get pool index
	poolSysconfigInfo, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.PoolAccountIndex)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to get sys config by name: %s", err.Error())
		return "", err
	}
	poolAccountIndex, err := strconv.ParseInt(poolSysconfigInfo.Value, 10, 64)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to parse pool account index: %s", err.Error())
		return "", err
	}

	if txInfo.ToAccountIndex != poolAccountIndex {
		return "", errors.New("[sendSwapTx] invalid pool index")
	}

	// init account info map
	var (
		accountInfoMap = make(map[int64]*commonAsset.FormatAccountInfo)
	)

	accountInfoMap[poolAccountIndex], err = globalmapHandler.GetLatestAccountInfo(
		l.svcCtx.AccountModel,
		l.svcCtx.AccountHistoryModel,
		l.svcCtx.MempoolModel,
		l.svcCtx.LiquidityPairModel,
		l.svcCtx.RedisConnection,
		poolAccountIndex,
	)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to get latest account info: %s", err.Error())
		return "", err
	}

	// check params
	if accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex] == nil ||
		accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetA == "" ||
		accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetA == util.ZeroBigInt.String() ||
		accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetB == "" ||
		accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetB == util.ZeroBigInt.String() {
		logx.Errorf("[sendSwapTx] invalid params")
		return "", errors.New("[sendSwapTx] invalid params")
	}

	// compute delta
	var (
		toDelta *big.Int
	)
	poolABalance, isValid := new(big.Int).SetString(accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetA, 10)
	if !isValid {
		logx.Errorf("[sendSwapTx] unable to parse amount")
		return "", errors.New("[sendSwapTx] unable to parse amount")
	}
	poolBBalance, isValid := new(big.Int).SetString(accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetA, 10)
	if !isValid {
		logx.Errorf("[sendSwapTx] unable to parse amount")
		return "", errors.New("[sendSwapTx] unable to parse amount")
	}
	if accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId == txInfo.AssetAId &&
		accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId == txInfo.AssetBId {
		toDelta, _, err = util.ComputeDelta(
			poolABalance,
			poolBBalance,
			accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId,
			accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId,
			txInfo.AssetAId,
			true,
			txInfo.AssetAAmount,
			txInfo.FeeRate)
		txInfo.PoolAAmount = poolABalance
		txInfo.PoolBAmount = poolBBalance
	} else if accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId == txInfo.AssetBId &&
		accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId == txInfo.AssetAId {
		toDelta, _, err = util.ComputeDelta(
			poolABalance,
			poolBBalance,
			accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId,
			accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId,
			txInfo.AssetAId,
			true,
			txInfo.AssetAAmount,
			txInfo.FeeRate)

		txInfo.PoolAAmount = poolBBalance
		txInfo.PoolBAmount = poolABalance
	} else {
		err = errors.New("invalid pair assetIds")
	}

	if err != nil {
		errInfo := fmt.Sprintf("[logic.sendSwapTx] => [util.ComputeDelta]: %s. invalid AssetId: %v/%v/%v",
			err.Error(), txInfo.AssetAId,
			uint32(accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetAId),
			uint32(accountInfoMap[poolAccountIndex].LiquidityInfo[txInfo.PairIndex].AssetBId))
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	// check if toDelta is over minToAmount
	if toDelta.Cmp(txInfo.AssetBMinAmount) < 0 {
		errInfo := fmt.Sprintf("[logic.sendSwapTx] => minToAmount is bigger than toDelta: %s/%s",
			txInfo.AssetBMinAmount.String(), toDelta.String())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	// complete tx info
	txInfo.AssetBAmountDelta = toDelta

	// get latest account info for from account index
	if accountInfoMap[txInfo.FromAccountIndex] == nil {
		accountInfoMap[txInfo.FromAccountIndex], err = globalmapHandler.GetLatestAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.AccountHistoryModel,
			l.svcCtx.MempoolModel,
			l.svcCtx.LiquidityPairModel,
			l.svcCtx.RedisConnection,
			txInfo.FromAccountIndex,
		)
		if err != nil {
			logx.Errorf("[sendSwapTx] unable to get latest account info: %s", err.Error())
			return "", err
		}
	}
	if accountInfoMap[txInfo.TreasuryAccountIndex] == nil {
		accountInfoMap[txInfo.TreasuryAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.TreasuryAccountIndex,
		)
		if err != nil {
			logx.Errorf("[sendSwapTx] unable to get latest account info: %s", err.Error())
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
			logx.Errorf("[sendSwapTx] unable to get latest account info: %s", err.Error())
			return "", err
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	/*
		Get txDetails
	*/

	// verify swap tx
	txDetails, err = txVerification.VerifySwapTxInfo(
		accountInfoMap,
		txInfo)

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", l.HandleCreateFailSwapTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeTransfer,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		txInfo.AssetAId,
		txInfo.AssetBId,
		txInfo.AssetAAmount.String(),
		"",
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txDetails,
	)
	err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return "", l.HandleCreateFailSwapTx(txInfo, err)
	}

	return txId, nil
}

func (l *SendTxLogic) HandleCreateFailSwapTx(txInfo *commonTx.SwapTxInfo, err error) error {
	errCreate := l.CreateFailSwapTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendswaptxlogic.HandleCreateFailSwapTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendswaptxlogic.HandleCreateFailSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendTxLogic) CreateFailSwapTx(info *commonTx.SwapTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	txFeeAssetId := info.GasFeeAssetId
	assetAId := info.AssetAId
	assetBId := info.AssetBId
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeSwap,
		// tx fee
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: int64(txFeeAssetId),
		// tx status, 1 - success(default), 2 - failure
		TxStatus: txHandler.TxFail,
		// AssetAId
		AssetAId: int64(assetAId),
		// l1asset id
		AssetBId: int64(assetBId),
		// tx amount
		TxAmount: util.ZeroBigInt.String(),
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
	}

	err = l.svcCtx.FailTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
