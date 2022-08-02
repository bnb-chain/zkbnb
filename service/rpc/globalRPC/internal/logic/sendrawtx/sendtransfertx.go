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

func SendTransferTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseTransferTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx.ParseTransferTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param: %v, err: %s", txInfo.GasFeeAssetAmount, err.Error())
		return "", err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param: %v, err: %s", txInfo.AssetAmount, err.Error())
		return "", err
	}
	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.AssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid assetId %v", txInfo.AssetId)
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.FromAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid accountIndex %v", txInfo.FromAccountIndex)
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.ToAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferTx] err: invalid accountIndex %v", txInfo.ToAccountIndex)
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	gasAccountIndexConfig, err := svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferTx] unable to get sysconfig by name: %s", err.Error())
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, errors.New("[sendTransferTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendTransferTx] invalid gas account index")
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, errors.New("[sendTransferTx] invalid gas account index"))
	}
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendTransferTx] invalid time stamp")
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, errors.New("[sendTransferTx] invalid time stamp"))
	}
	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.FromAccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, err)
	}
	// get account info by to index
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			txInfo.ToAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
			return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, err)
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("[sendTransferTx] invalid account name")
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, errors.New("[sendTransferTx] invalid account name"))
	}
	// get account info by gas index
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		// get account info by gas index
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferTx] unable to get account info: %s", err.Error())
			return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, err)
		}
	}
	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify transfer tx
	txDetails, err = txVerification.VerifyTransferTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, err)
	}

	/*
		Check tx details
	*/

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeTransfer,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		commonConstant.NilPairIndex,
		txInfo.AssetId,
		txInfo.AssetAmount.String(),
		"",
		string(txInfoBytes),
		txInfo.Memo,
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	err = CreateMempoolTx(mempoolTx, svcCtx.RedisConnection, svcCtx.MempoolModel)
	if err != nil {
		return "", handleCreateFailTransferTx(svcCtx.FailTxModel, txInfo, err)
	}
	return txId, nil
}

func handleCreateFailTransferTx(failTxModel tx.FailTxModel, txInfo *commonTx.TransferTxInfo, err error) error {
	errCreate := createFailTransferTx(failTxModel, txInfo, err.Error())
	if errCreate != nil {
		logx.Errorf("[sendtransfertxlogic.HandleCreateFailTransferTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateFailTransferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func createFailTransferTx(failTxModel tx.FailTxModel, info *commonTx.TransferTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	txFeeAssetId := info.AssetId
	assetId := info.AssetId
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailTransferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeTransfer,
		// tx fee
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: txFeeAssetId,
		// tx status, 1 - success(default), 2 - failure
		TxStatus: tx.StatusFail,
		// l1asset id
		AssetAId: assetId,
		// AssetBId
		AssetBId: commonConstant.NilAssetId,
		// tx amount
		TxAmount: info.AssetAmount.String(),
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
		// native memo info
		Memo: info.Memo,
	}

	err = failTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailTransferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
