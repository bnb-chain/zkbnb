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

func SendTransferNftTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	// parse transfer tx info
	txInfo, err := commonTx.ParseTransferNftTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferNftTx.ParseTransferNftTxInfo] %s", err.Error())
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
	// check param: from account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.FromAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferNftTx] err: invalid accountIndex %v", txInfo.FromAccountIndex)
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.ToAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendTransferNftTx] err: invalid accountIndex %v", txInfo.ToAccountIndex)
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	commglobalmap.DeleteLatestAccountInfoInCache(ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[DeleteLatestAccountInfoInCache] err:%v", err)
	}
	commglobalmap.DeleteLatestAccountInfoInCache(ctx, txInfo.ToAccountIndex)
	if err != nil {
		logx.Errorf("[DeleteLatestAccountInfoInCache] err:%v", err)
	}
	// check gas account index
	gasAccountIndexConfig, err := svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to get sysconfig by name: %s", err.Error())
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, errors.New("[sendTransferNftTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendTransferNftTx] invalid gas account index")
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, errors.New("[sendTransferNftTx] invalid gas account index"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	nftInfo, err := globalmapHandler.GetLatestNftInfoForRead(
		svcCtx.NftModel,
		svcCtx.MempoolModel,
		svcCtx.RedisConnection,
		txInfo.NftIndex,
	)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to get nft info")
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, err)
	}
	accountInfoMap[txInfo.FromAccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to get account info: %s", err.Error())
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, err)
	}
	// get account info by to index
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			txInfo.ToAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferNftTx] unable to get account info: %s", err.Error())
			return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, err)
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("[sendTransferNftTx] invalid account name")
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, errors.New("[sendTransferNftTx] invalid account name"))
	}
	// get account info by gas index
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		// get account info by gas index
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendTransferNftTx] unable to get account info: %s", err.Error())
			return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, err)
		}
	}

	if nftInfo.OwnerAccountIndex != txInfo.FromAccountIndex {
		logx.Errorf("[sendTransferNftTx] you're not owner")
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, errors.New("[sendTransferNftTx] you're not owner"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendTransferNftTx] invalid time stamp")
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, errors.New("[sendTransferNftTx] invalid time stamp"))
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify transfer tx
	txDetails, err = txVerification.VerifyTransferNftTxInfo(
		accountInfoMap,
		nftInfo,
		txInfo,
	)
	if err != nil {
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, err)
	}

	/*
		Check tx details
	*/

	/*
		Create Mempool Transaction
	*/
	// delete key
	key := util.GetNftKeyForRead(txInfo.NftIndex)
	_, err = svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to delete key from redis: %s", err.Error())
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, err)
	}
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, err)
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
	err = CreateMempoolTx(mempoolTx, svcCtx.RedisConnection, svcCtx.MempoolModel)
	if err != nil {
		return "", handleCreateFailTransferNftTx(svcCtx.FailTxModel, txInfo, err)
	}

	// update redis
	var formatNftInfo *commonAsset.NftInfo
	for _, txDetail := range mempoolTx.MempoolDetails {
		if txDetail.AssetType == commonAsset.NftAssetType {
			formatNftInfo, err = commonAsset.ParseNftInfo(txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[sendTransferNftTx] unable to parse nft info: %s", err.Error())
				return txId, nil
			}
		}
	}
	nftInfoBytes, err := json.Marshal(formatNftInfo)
	if err != nil {
		logx.Errorf("[sendTransferNftTx] unable to marshal: %s", err.Error())
		return txId, nil
	}
	_ = svcCtx.RedisConnection.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)
	return txId, nil
}

func handleCreateFailTransferNftTx(failTxModel tx.FailTxModel, txInfo *commonTx.TransferNftTxInfo, err error) error {
	errCreate := createFailTransferNftTx(failTxModel, txInfo, err.Error())
	if errCreate != nil {
		logx.Errorf("[sendtransfertxlogic.HandleCreateFailTransferNftTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateFailTransferNftTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func createFailTransferNftTx(failTxModel tx.FailTxModel, info *commonTx.TransferNftTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailTransferNftTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeTransferNft,
		// tx fee
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: info.GasFeeAssetId,
		// tx status, 1 - success(default), 2 - failure
		TxStatus: tx.StatusFail,
		// l1asset id
		AssetAId: commonConstant.NilAssetId,
		// AssetBId
		AssetBId: commonConstant.NilAssetId,
		// tx amount
		TxAmount: commonConstant.NilAssetAmountStr,
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
		// native memo info
		Memo: "",
	}

	err = failTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailTransferNftTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
