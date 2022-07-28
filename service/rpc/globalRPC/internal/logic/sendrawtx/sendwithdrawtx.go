package sendrawtx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

func SendWithdrawTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	// parse withdraw tx info
	txInfo, err := commonTx.ParseWithdrawTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendWithdrawTx.ParseWithdrawTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	/*
		Check Params
	*/
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return "", err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetAmount, err)
		return "", err
	}
	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.AssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendWithdrawTx] err: invalid assetId %v", txInfo.AssetId)
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}

	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.GasFeeAssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendWithdrawTx] err: invalid gasFeeAssetId %v", txInfo.GasFeeAssetId)
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	commglobalmap.DeleteLatestAccountInfoInCache(ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[DeleteLatestAccountInfoInCache] err:%v", err)
	}
	// check gas account index
	gasAccountIndexConfig, err := svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendWithdrawTx] unable to get sysconfig by name: %s", err.Error())
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, errors.New("[sendWithdrawTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendWithdrawTx] invalid gas account index")
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, errors.New("[sendWithdrawTx] invalid gas account index"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendWithdrawTx] invalid time stamp")
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, errors.New("[sendWithdrawTx] invalid time stamp"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.FromAccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[sendWithdrawTx] unable to get account info: %s", err.Error())
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, err)
	}
	// get account info by gas index
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		// get account info by gas index
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendWithdrawTx] unable to get account info: %s", err.Error())
			return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, err)
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	/*
		Get txDetails
	*/
	// verify withdraw tx
	txDetails, err = txVerification.VerifyWithdrawTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, err)
	}

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeWithdraw,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		commonConstant.NilPairIndex,
		txInfo.AssetId,
		txInfo.AssetAmount.String(),
		txInfo.ToAddress,
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	err = CreateMempoolTx(mempoolTx, svcCtx.RedisConnection, svcCtx.MempoolModel)
	if err != nil {
		return "", handleCreateFailWithdrawTx(svcCtx.FailTxModel, txInfo, err)
	}

	return txId, nil
}

func handleCreateFailWithdrawTx(failTxModel tx.FailTxModel, txInfo *commonTx.WithdrawTxInfo, err error) error {
	errCreate := createFailWithdrawTx(failTxModel, txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendwithdrawtxlogic.HandleCreateFailWithdrawTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendwithdrawtxlogic.HandleCreateFailWithdrawTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func createFailWithdrawTx(failTxModel tx.FailTxModel, info *commonTx.WithdrawTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	txFeeAssetId := info.AssetId
	assetId := info.AssetId
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailWithdrawTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeWithdraw,
		// tx fee
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: int64(txFeeAssetId),
		// tx status, 1 - success(default), 2 - failure
		TxStatus: TxFail,
		// l1asset id
		AssetAId: int64(assetId),
		// tx amount
		TxAmount: info.AssetAmount.String(),
		// layer1 address
		NativeAddress: info.ToAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
	}

	err = failTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailWithdrawTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
