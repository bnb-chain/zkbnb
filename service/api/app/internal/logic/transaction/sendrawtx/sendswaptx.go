package sendrawtx

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

func SendSwapTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseSwapTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.AppErrInvalidTx
	}

	if err := legendTxTypes.ValidateSwapTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError(err)
	}

	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, svcCtx.SysConfigModel); err != nil {
		return "", err
	}

	liquidityInfo, err := commglobalmap.GetLatestLiquidityInfoForWrite(ctx, txInfo.PairIndex)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to get latest liquidity info for write: %s", err.Error())
		return "", errorcode.AppErrInternal
	}

	// check params
	if liquidityInfo.AssetA == nil || liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidityInfo.AssetB == nil || liquidityInfo.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("invalid params")
		return "", errorcode.AppErrInternal
	}

	// compute delta
	var (
		toDelta *big.Int
	)
	if liquidityInfo.AssetAId == txInfo.AssetAId &&
		liquidityInfo.AssetBId == txInfo.AssetBId {
		toDelta, _, err = util.ComputeDelta(
			liquidityInfo.AssetA,
			liquidityInfo.AssetB,
			liquidityInfo.AssetAId,
			liquidityInfo.AssetBId,
			txInfo.AssetAId,
			true,
			txInfo.AssetAAmount,
			liquidityInfo.FeeRate,
		)
	} else if liquidityInfo.AssetAId == txInfo.AssetBId &&
		liquidityInfo.AssetBId == txInfo.AssetAId {
		toDelta, _, err = util.ComputeDelta(
			liquidityInfo.AssetA,
			liquidityInfo.AssetB,
			liquidityInfo.AssetAId,
			liquidityInfo.AssetBId,
			txInfo.AssetBId,
			true,
			txInfo.AssetAAmount,
			liquidityInfo.FeeRate,
		)
	} else {
		err = errors.New("invalid pair assetIds")
	}
	if err != nil {
		logx.Errorf("invalid AssetIds: %d %d %d, err: %s",
			txInfo.AssetAId,
			uint32(liquidityInfo.AssetAId),
			uint32(liquidityInfo.AssetBId),
			err.Error())
		return "", errorcode.AppErrInternal
	}
	// check if toDelta is less than minToAmount
	if toDelta.Cmp(txInfo.AssetBMinAmount) < 0 {
		logx.Errorf("minToAmount is bigger than toDelta: %s, %s",
			txInfo.AssetBMinAmount.String(), toDelta.String())
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid AssetBMinAmount")
	}
	// complete tx info
	txInfo.AssetBAmountDelta = toDelta

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.FromAccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.FromAccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.FromAccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = commglobalmap.GetBasicAccountInfo(ctx, txInfo.GasAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("invalid GasAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.GasAccountIndex, err.Error())
			return "", errorcode.AppErrInternal
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify  tx
	txDetails, err = txVerification.VerifySwapTxInfo(
		accountInfoMap,
		liquidityInfo,
		txInfo)
	if err != nil {
		return "", errorcode.AppErrVerification.RefineError(err)
	}

	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.AppErrInternal
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeSwap,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		txInfo.PairIndex,
		commonConstant.NilAssetId,
		txInfo.AssetAAmount.String(),
		"",
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	// delete key
	keyW := util.GetLiquidityKeyForWrite(txInfo.PairIndex)
	keyR := util.GetLiquidityKeyForRead(txInfo.PairIndex)
	_, err = svcCtx.RedisConn.Del(keyW)
	if err != nil {
		logx.Errorf("unable to delete key from redis, key: %s, err: %s", keyW, err.Error())
		return "", errorcode.AppErrInternal
	}
	_, err = svcCtx.RedisConn.Del(keyR)
	if err != nil {
		logx.Errorf("unable to delete key from redis, key: %s, err: %s", keyR, err.Error())
		return "", errorcode.AppErrInternal
	}
	// insert into mempool
	if err := svcCtx.MempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{mempoolTx}); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = CreateFailTx(svcCtx.FailTxModel, commonTx.TxTypeSwap, txInfo, err)
		return "", errorcode.AppErrInternal
	}
	// update redis
	// get latest liquidity info
	for _, txDetail := range txDetails {
		if txDetail.AssetType == commonAsset.LiquidityAssetType {
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("unable to compute new balance: %s", err.Error())
				return txId, errorcode.AppErrInternal
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				logx.Errorf("unable to parse liquidity info: %s", err.Error())
				return txId, errorcode.AppErrInternal
			}
		}
	}
	liquidityInfoBytes, err := json.Marshal(liquidityInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return txId, errorcode.AppErrInternal
	}
	_ = svcCtx.RedisConn.Setex(keyW, string(liquidityInfoBytes), globalmapHandler.LiquidityExpiryTime)
	return txId, nil
}
