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
	"reflect"

	"github.com/zecrey-labs/zecrey/common/commonAccount"
	"github.com/zecrey-labs/zecrey/common/commonTx"
	"github.com/zecrey-labs/zecrey/common/model/account"
	"github.com/zecrey-labs/zecrey/common/model/asset"
	"github.com/zecrey-labs/zecrey/common/model/liquidityPair"
	"github.com/zecrey-labs/zecrey/common/model/mempool"
	"github.com/zecrey-labs/zecrey/common/model/tx"
	"github.com/zecrey-labs/zecrey/common/utils"
	"github.com/zecrey-labs/zecrey/common/zcrypto/cryptoUtils"
	"github.com/zecrey-labs/zecrey/common/zcrypto/zecreyProofs"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/logic/globalmapHandler"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/logic/txHandler"
	"github.com/zeromicro/go-zero/core/logx"
	"go.etcd.io/etcd/client/v3/concurrency"
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
	err = utils.CheckRequestParam(utils.TypeAssetId, reflect.ValueOf(txInfo.AssetAId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendSwapTx] err: invalid assetAId %v", txInfo.AssetAId)
		return "", l.HandleCreateSwapFailTx(txInfo, errors.New(errInfo))
	}

	err = utils.CheckRequestParam(utils.TypeAssetId, reflect.ValueOf(txInfo.AssetBId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendSwapTx] err: invalid assetBId %v", txInfo.AssetBId)
		return "", l.HandleCreateSwapFailTx(txInfo, errors.New(errInfo))
	}

	err = utils.CheckRequestParam(utils.TypeGasFee, reflect.ValueOf(txInfo.GasFee))
	if err != nil {
		errInfo := fmt.Sprintf("[sendSwapTx] err: invalid gas fee %v", txInfo.GasFee)
		return "", l.HandleCreateSwapFailTx(txInfo, errors.New(errInfo))
	}

	/*
		 PairIndex     uint32
			AccountIndex  uint32
			AssetAId      uint32
			AssetBId      uint32
			GasFeeAssetId uint32
			GasFee        uint64
			FeeRate       uint32
			TreasuryRate  uint32
			B_A_Delta     uint64
			MinB_B_Delta  uint64
			Proof         string
	*/

	var (
		fromAccountIndex = txInfo.AccountIndex
		fromAssetId      = txInfo.AssetAId
		toAssetId        = txInfo.AssetBId
		fromAmount       = txInfo.B_A_Delta
		minToAmount      = txInfo.MinB_B_Delta
		feeRate          = txInfo.FeeRate
		// gasFee           = txInfo.GasFee
		gasFeeAssetId = txInfo.GasFeeAssetId
		// treasuryRate     = txInfo.TreasuryRate

		toDelta uint64
	)
	/*
		accountAssetAInfo,
		accountAssetBInfo,
		poolAccountInfo,
		accountAssetGasInfo,
		gasAccountInfo,
		treasuryAccountInfo,
	*/

	// todo:p1 checker
	// delta

	var (
		pairInfo         *liquidityPair.LiquidityPair
		poolAccountInfo  *account.AccountHistory
		poolLiquidity    *asset.AccountLiquidity
		poolAssetAAmount uint64
		poolAssetBAmount uint64
	)

	pairInfo, poolAccountInfo, poolLiquidity,
		poolAssetAAmount, poolAssetBAmount, err = GetLatestPoolInfo(l.svcCtx, txInfo.PairIndex)

	if err != nil {
		errInfo := fmt.Sprintf("[sendSwapTx] %s", err.Error())
		return "", errors.New(errInfo)
	}

	// todo:p1 check if feeRate is valid
	/*
		if uint32(pairInfo.FeeRate) != feeRate{
			errInfo := fmt.Sprintf("[logic.sendSwapTx] => Invalid feeRate: %v", feeRate)
			logx.Error(errInfo)
			return "", errors.New(errInfo)
		}

	*/

	// compute delta
	if uint32(pairInfo.AssetAId) == fromAssetId && uint32(pairInfo.AssetBId) == toAssetId {
		toDelta, _, err = utils.ComputeDelta(
			poolAssetAAmount,
			poolAssetBAmount,
			uint32(pairInfo.AssetAId),
			uint32(pairInfo.AssetBId),
			fromAssetId,
			true,
			fromAmount,
			int64(feeRate))
	} else if uint32(pairInfo.AssetAId) == toAssetId && uint32(pairInfo.AssetBId) == fromAssetId {
		toDelta, _, err = utils.ComputeDelta(
			poolAssetAAmount,
			poolAssetBAmount,
			uint32(pairInfo.AssetAId),
			uint32(pairInfo.AssetBId),
			fromAssetId,
			true,
			fromAmount,
			int64(feeRate))
	} else {
		err = errors.New("invalid pair assetIds")
	}

	if err != nil {
		errInfo := fmt.Sprintf("[logic.sendSwapTx] => [utils.ComputeDelta]: %s. invalid AssetId: %v/%v/%v",
			err.Error(), txInfo.AssetAId, uint32(pairInfo.AssetAId), uint32(pairInfo.AssetBId))
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	// check if toDelta is over minToAmount
	if minToAmount > toDelta {
		errInfo := fmt.Sprintf("[logic.sendSwapTx] => minToAmount is bigger than toDelta: %v/%v",
			minToAmount, toDelta)
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	var (
		accountAssetAInfo   *zecreyProofs.AccountAssetInfo
		accountAssetBInfo   *zecreyProofs.AccountAssetInfo
		poolLiquidityInfo   *zecreyProofs.AccountLiquidityInfo
		accountAssetGasInfo *zecreyProofs.AccountAssetInfo
		gasAssetInfo        *zecreyProofs.AccountAssetInfo
		treasuryAccountInfo *zecreyProofs.AccountAssetInfo
	)

	/*
		accountAssetAInfo,
		accountAssetBInfo,
		poolAccountInfo,
		accountAssetGasInfo,
		gasAccountInfo,
		treasuryAccountInfo,
	*/

	// accountAssetAInfo

	ctx := context.Background()
	lockArray := make([]*concurrency.Mutex, 0)

	globalKey := globalmapHandler.GetAccountAssetGlobalKey(uint32(fromAccountIndex), fromAssetId)
	keyLock := PrefixLock + globalKey
	lockFromAsset := concurrency.NewMutex(globalmapHandler.GlobalEtcd.S, keyLock)
	if err := lockFromAsset.TryLock(ctx); err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", err
	}
	lockArray = append([]*concurrency.Mutex{lockFromAsset}, lockArray...)

	accountSingleAssetA, err := GetLatestSingleAccountAsset(l.svcCtx, fromAccountIndex, fromAssetId)
	if err != nil {
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}
	accountAssetAInfo, err = zecreyProofs.ConstructAccountAssetInfo(
		accountSingleAssetA.AccountId,
		int64(accountSingleAssetA.AccountIndex),
		accountSingleAssetA.AccountName,
		accountSingleAssetA.PublicKey,
		int64(accountSingleAssetA.AssetId),
		accountSingleAssetA.BalanceEnc,
	)
	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}

	// accountAssetBInfo

	globalKey = globalmapHandler.GetAccountAssetGlobalKey(uint32(fromAccountIndex), toAssetId)
	keyLock = PrefixLock + globalKey
	lockToAsset := concurrency.NewMutex(globalmapHandler.GlobalEtcd.S, keyLock)
	if err := lockToAsset.TryLock(ctx); err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", err
	}
	lockArray = append([]*concurrency.Mutex{lockToAsset}, lockArray...)

	accountSingleAssetB, err := GetLatestSingleAccountAsset(l.svcCtx, fromAccountIndex, toAssetId)
	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}
	accountAssetBInfo, err = zecreyProofs.ConstructAccountAssetInfo(
		accountSingleAssetB.AccountId,
		int64(accountSingleAssetB.AccountIndex),
		accountSingleAssetB.AccountName,
		accountSingleAssetB.PublicKey,
		int64(accountSingleAssetB.AssetId),
		accountSingleAssetB.BalanceEnc,
	)
	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}

	// poolAccountInfo
	poolLiquidityInfo, err = zecreyProofs.ConstructAccountLiquidityInfo(
		// account info
		poolAccountInfo.ID,
		poolAccountInfo.AccountIndex,
		poolAccountInfo.AccountName,
		poolAccountInfo.PublicKey,

		// liquidity info
		pairInfo.PairIndex,
		pairInfo.AssetAId,
		pairInfo.AssetBId,
		int64(poolAssetAAmount),
		poolLiquidity.AssetAR,
		int64(poolAssetBAmount),
		poolLiquidity.AssetBR,
		pairInfo.FeeRate,
		pairInfo.TreasuryRate,
		poolLiquidity.LpEnc,
	)
	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}

	globalKey = globalmapHandler.GetAccountAssetGlobalKey(uint32(fromAccountIndex), gasFeeAssetId)
	keyLock = PrefixLock + globalKey
	lockAssetGasFee := concurrency.NewMutex(globalmapHandler.GlobalEtcd.S, keyLock)
	if err := lockAssetGasFee.TryLock(ctx); err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", err
	}
	lockArray = append([]*concurrency.Mutex{lockAssetGasFee}, lockArray...)

	// accountAssetGasInfo
	accountSingleAssetGas, err := GetLatestSingleAccountAsset(l.svcCtx, fromAccountIndex, gasFeeAssetId)
	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)

	}

	accountAssetGasInfo, err = zecreyProofs.ConstructAccountAssetInfo(
		accountSingleAssetGas.AccountId,
		int64(accountSingleAssetGas.AccountIndex),
		accountSingleAssetGas.AccountName,
		accountSingleAssetGas.PublicKey,
		int64(accountSingleAssetGas.AssetId),
		accountSingleAssetGas.BalanceEnc,
	)

	// gasAccountInfo
	gasAccountSingleAssetGas, err := GetLatestSingleAccountAsset(l.svcCtx, commonAccount.GasAccountIndex, gasFeeAssetId)
	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}
	gasAssetInfo, err = zecreyProofs.ConstructAccountAssetInfo(
		gasAccountSingleAssetGas.AccountId,
		int64(gasAccountSingleAssetGas.AccountIndex),
		gasAccountSingleAssetGas.AccountName,
		gasAccountSingleAssetGas.PublicKey,
		int64(gasAccountSingleAssetGas.AssetId),
		gasAccountSingleAssetGas.BalanceEnc,
	)

	if err != nil {
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}

	// treasuryAccountInfo
	treasuryAccountSingleAssetA, err := GetLatestSingleAccountAsset(l.svcCtx, commonAccount.TreasuryAccountIndex, fromAssetId)
	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}
	treasuryAccountInfo, err = zecreyProofs.ConstructAccountAssetInfo(
		treasuryAccountSingleAssetA.AccountId,
		int64(treasuryAccountSingleAssetA.AccountIndex),
		treasuryAccountSingleAssetA.AccountName,
		treasuryAccountSingleAssetA.PublicKey,
		int64(treasuryAccountSingleAssetA.AssetId),
		treasuryAccountSingleAssetA.BalanceEnc,
	)

	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	/*
		Get txDetails
	*/

	// verify swap tx
	txDetails, err = zecreyProofs.VerifySwapTx(
		toDelta,
		accountAssetAInfo,
		accountAssetBInfo,
		poolLiquidityInfo,
		accountAssetGasInfo,
		gasAssetInfo,
		treasuryAccountInfo,
		txInfo)
	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}

	/*
		Check tx details
	*/

	// check txDetails length
	if len(txDetails) != zecreyProofs.SwapTxDetailsCount {
		errInfo := fmt.Sprintf("[sendtxlogic.sendSwapTx] txDetails count error, len(txDetails) = %v", len(txDetails))
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, errors.New(errInfo))
	}

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txId, err = l.CreateTxMempoolForSwapTx(txDetails, txInfo, txHandler.TxTypeSwap)
	if err != nil {
		for _, lock := range lockArray {
			lock.Unlock(ctx)
		}
		return "", l.HandleCreateSwapFailTx(txInfo, err)
	}

	for _, lock := range lockArray {
		lock.Unlock(ctx)
	}

	return txId, nil
}

func (l *SendTxLogic) HandleCreateSwapFailTx(txInfo *commonTx.SwapTxInfo, err error) error {
	errCreate := l.CreateFailSwapTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendswaptxlogic.HandleCreateSwapFailTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendswaptxlogic.HandleCreateSwapFailTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendTxLogic) CreateFailSwapTx(info *commonTx.SwapTxInfo, extraInfo string) error {
	txHash := cryptoUtils.GetRandomUUID()
	txType := int64(txHandler.TxTypeSwap)
	txFee := info.GasFee
	txFeeAssetId := info.GasFeeAssetId
	assetAId := info.AssetAId
	assetBId := info.AssetBId
	txAmount := int64(0)
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
		TxType: txType,
		// tx fee
		GasFee: int64(txFee),
		// tx fee l1asset id
		GasFeeAssetId: int64(txFeeAssetId),
		// tx status, 1 - success(default), 2 - failure
		TxStatus: txHandler.TxFail,
		// AssetAId
		AssetAId: int64(assetAId),
		// l1asset id
		AssetBId: int64(assetBId),
		// tx amount
		TxAmount: int64(txAmount),
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

func (l *SendTxLogic) CreateTxMempoolForSwapTx(nMempoolTxDetails []*mempool.MempoolTxDetail, txInfo *commonTx.SwapTxInfo, txType uint8) (resTxId string, err error) {

	var (
		nMempoolTx *mempool.MempoolTx
		bTxInfo    []byte
	)
	// generate tx id by random UUID
	resTxId = cryptoUtils.GetRandomUUID()
	// Marshal txInfo
	bTxInfo, err = json.Marshal(txInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateTxMempoolForSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	nMempoolTx = &mempool.MempoolTx{
		TxHash:         resTxId,
		TxType:         int64(txType),
		GasFee:         int64(txInfo.GasFee),
		GasFeeAssetId:  int64(txInfo.GasFeeAssetId),
		AssetAId:       int64(txInfo.AssetAId),
		AssetBId:       int64(txInfo.AssetBId),
		TxAmount:       0,
		MempoolDetails: nMempoolTxDetails,
		ChainId:        commonTx.L2TxChainId,
		TxInfo:         string(bTxInfo),
		ExtraInfo:      "",
		Memo:           "",
		L2BlockHeight:  0,
		Status:         0,
	}

	// write into mempool
	err = l.svcCtx.MempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{nMempoolTx})
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateTxMempoolForSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	// update mempool state
	err = globalmapHandler.UpdateGlobalMap(nMempoolTx)
	if err != nil {
		logx.Error(err)
		return "", err
	}

	return resTxId, nil
}
