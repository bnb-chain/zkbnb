package sendrawtx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

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

func SendRemoveLiquidityTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {

	// parse removeliquidity tx info
	txInfo, err := commonTx.ParseRemoveLiquidityTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendRemoveLiquidityTx] => [commonTx.ParseRemoveLiquidityTxInfo] : %s. invalid rawTxInfo %s",
			err.Error(), rawTxInfo)
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return "", err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAMinAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetAMinAmount, err)
		return "", err
	}
	if err := util.CheckPackedAmount(txInfo.AssetBMinAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetBMinAmount, err)
		return "", err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAAmountDelta); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetAAmountDelta, err)
		return "", err
	}
	if err := util.CheckPackedAmount(txInfo.AssetBAmountDelta); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetBAmountDelta, err)
		return "", err
	}
	err = commglobalmap.DeleteLatestAccountInfoInCache(ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[DeleteLatestAccountInfoInCache] err:%v", err)
	}
	// check gas account index
	gasAccountIndexConfig, err := svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendRemoveLiquidityTx] unable to get sysconfig by name: %s", err.Error())
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("[ParseInt] err: %s", err)
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, errors.New("[sendRemoveLiquidityTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendRemoveLiquidityTx] invalid gas account index")
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, errors.New("[sendRemoveLiquidityTx] invalid gas account index"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendRemoveLiquidityTx] invalid time stamp")
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, errors.New("[sendRemoveLiquidityTx] invalid time stamp"))
	}

	var (
		redisLock      *redis.RedisLock
		liquidityInfo  *commonAsset.LiquidityInfo
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)

	redisLock, liquidityInfo, err = globalmapHandler.GetLatestLiquidityInfoForWrite(
		svcCtx.LiquidityModel,
		svcCtx.MempoolModel,
		svcCtx.RedisConnection,
		txInfo.PairIndex,
	)
	if err != nil {
		logx.Errorf("[sendRemoveLiquidityTx] unable to get latest liquidity info for write: %s", err.Error())
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, err)
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
	assetAAmount, assetBAmount, err = util.ComputeRemoveLiquidityAmount(liquidityInfo, txInfo.LpAmount)
	if err != nil {
		logx.Errorf("[sendRemoveLiquidityTx] unable to compute lp portion: %s", err.Error())
		return "", err
	}
	if assetAAmount.Cmp(txInfo.AssetAMinAmount) < 0 || assetBAmount.Cmp(txInfo.AssetBMinAmount) < 0 {
		errInfo := fmt.Sprintf("[sendRemoveLiquidityTx] less than MinDelta: %s:%s/%s:%s",
			txInfo.AssetAMinAmount.String(), txInfo.AssetBMinAmount.String(), assetAAmount.String(), assetBAmount.String())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	// add into tx info
	txInfo.AssetAAmountDelta = assetAAmount
	txInfo.AssetBAmountDelta = assetBAmount

	// get latest account info for from account index
	if accountInfoMap[txInfo.FromAccountIndex] == nil {
		accountInfoMap[txInfo.FromAccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.FromAccountIndex)
		if err != nil {
			logx.Errorf("[sendRemoveLiquidityTx] unable to get latest account info: %s", err.Error())
			return "", err
		}
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			txInfo.GasAccountIndex,
		)
		if err != nil {
			logx.Errorf("[sendRemoveLiquidityTx] unable to get latest account info: %s", err.Error())
			return "", err
		}
	}
	if accountInfoMap[liquidityInfo.TreasuryAccountIndex] == nil {
		accountInfoMap[liquidityInfo.TreasuryAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
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
		logx.Errorf("[VerifyRemoveLiquidityTxInfo] err: %v", err)
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, err)
	}

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("[Marshal] err: %v", err)
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, err)
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
	_, err = svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("[sendRemoveLiquidityTx] unable to delete key from redis: %s", err.Error())
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, err)
	}
	_, err = svcCtx.RedisConnection.Del(key2)
	if err != nil {
		logx.Errorf("[sendRemoveLiquidityTx] unable to delete key from redis: %s", err.Error())
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, err)
	}
	// insert into mempool
	err = CreateMempoolTx(mempoolTx, svcCtx.RedisConnection, svcCtx.MempoolModel)
	if err != nil {
		logx.Errorf("[CreateMempoolTx] err: %v", err)
		return "", handleCreateFailRemoveLiquidityTx(svcCtx.FailTxModel, txInfo, err)
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
	_ = svcCtx.RedisConnection.Setex(key, string(liquidityInfoBytes), globalmapHandler.LiquidityExpiryTime)

	return txId, nil
}

func handleCreateFailRemoveLiquidityTx(failTxModel tx.FailTxModel, txInfo *commonTx.RemoveLiquidityTxInfo, err error) error {
	errCreate := createFailRemoveLiquidityTx(failTxModel, txInfo, err.Error())
	if errCreate != nil {
		logx.Errorf("[sendremoveliquiditytxlogic.HandleCreateFailRemoveLiquidityTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendremoveliquiditytxlogic.HandleCreateFailRemoveLiquidityTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func createFailRemoveLiquidityTx(failTxModel tx.FailTxModel, info *commonTx.RemoveLiquidityTxInfo, extraInfo string) error {
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

	err = failTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailRemoveLiquidityTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
