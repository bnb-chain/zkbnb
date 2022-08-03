package sendrawtx

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

func SendAddLiquidityTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.RpcErrInvalidTx
	}

	if err := legendTxTypes.ValidateAddLiquidityTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.RpcErrInvalidTxField.RefineError(err)
	}

	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, svcCtx.SysConfigModel); err != nil {
		return txId, err
	}

	liquidityInfo, err := commglobalmap.GetLatestLiquidityInfoForWrite(ctx, txInfo.PairIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.RpcErrInvalidTxField.RefineError("invalid PairIndex")
		}
		logx.Errorf("fail to get liquidity info: %d, err: %s", txInfo.PairIndex, err.Error())
		return "", err
	}
	if liquidityInfo.AssetA == nil || liquidityInfo.AssetB == nil {
		logx.Errorf("invalid liquidity assets")
		return "", errorcode.RpcErrInternal
	}
	if liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 {
		txInfo.LpAmount, err = util.ComputeEmptyLpAmount(txInfo.AssetAAmount, txInfo.AssetBAmount)
		if err != nil {
			logx.Errorf("cannot computer lp amount, err: %s", err.Error())
			return "", errorcode.RpcErrInternal
		}
	} else {
		txInfo.LpAmount, err = util.ComputeLpAmount(liquidityInfo, txInfo.AssetAAmount)
		if err != nil {
			logx.Errorf("cannot computer lp amount, err: %s", err.Error())
			return "", errorcode.RpcErrInternal
		}
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.FromAccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.FromAccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.RpcErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.FromAccountIndex, err.Error())
		return "", errorcode.RpcErrInternal
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = commglobalmap.GetBasicAccountInfo(ctx, txInfo.GasAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return txId, errorcode.RpcErrInvalidTxField.RefineError("invalid GasAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.GasAccountIndex, err.Error())
			return "", errorcode.RpcErrInternal
		}
	}
	if accountInfoMap[liquidityInfo.TreasuryAccountIndex] == nil {
		accountInfoMap[liquidityInfo.TreasuryAccountIndex], err = commglobalmap.GetBasicAccountInfo(ctx, liquidityInfo.TreasuryAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return txId, errorcode.RpcErrInvalidTxField.RefineError("invalid liquidity")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", liquidityInfo.TreasuryAccountIndex, err.Error())
			return "", errorcode.RpcErrInternal
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify tx
	txDetails, err = txVerification.VerifyAddLiquidityTxInfo(
		accountInfoMap,
		liquidityInfo,
		txInfo)
	if err != nil {
		return "", errorcode.RpcErrVerification.RefineError(err)
	}

	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.RpcErrInternal
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
	if err := commglobalmap.DeleteLatestLiquidityInfoForWriteInCache(ctx, txInfo.PairIndex); err != nil {
		logx.Errorf("fail to delete liquidity info: %d, err: %s", txInfo.PairIndex, err.Error())
		return "", errorcode.RpcErrInternal
	}
	if err = CreateMempoolTx(mempoolTx, svcCtx.RedisConnection, svcCtx.MempoolModel); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = CreateFailTx(svcCtx.FailTxModel, commonTx.TxTypeAddLiquidity, txInfo, err)
		return "", err
	}
	// update cache, not key logic
	if err := commglobalmap.SetLatestLiquidityInfoForWrite(ctx, txInfo.PairIndex); err != nil {
		logx.Errorf("[SetLatestLiquidityInfoForWrite] param: %d, err: %s", txInfo.PairIndex, err.Error())
	}
	return txId, nil
}
