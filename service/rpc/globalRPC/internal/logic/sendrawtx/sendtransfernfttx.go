package sendrawtx

import (
	"context"
	"encoding/json"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

func SendTransferNftTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseTransferNftTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.RpcErrInvalidTx
	}

	if err := legendTxTypes.ValidateTransferNftTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.RpcErrInvalidTxField.RefineError(err)
	}

	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, svcCtx.SysConfigModel); err != nil {
		return "", err
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
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = commglobalmap.GetBasicAccountInfo(ctx, txInfo.ToAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.RpcErrInvalidTxField.RefineError("invalid ToAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.ToAccountIndex, err.Error())
			return "", errorcode.RpcErrInternal
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("invalid account name hash, expected: %s, actual: %s", accountInfoMap[txInfo.ToAccountIndex].AccountNameHash, txInfo.ToAccountNameHash)
		return "", errorcode.RpcErrInvalidTxField.RefineError("invalid ToAccountNameHash")
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = commglobalmap.GetBasicAccountInfo(ctx, txInfo.GasAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.RpcErrInvalidTxField.RefineError("invalid GasAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.GasAccountIndex, err.Error())
			return "", errorcode.RpcErrInternal
		}
	}

	nftInfo, err := commglobalmap.GetLatestNftInfoForRead(ctx, txInfo.NftIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.RpcErrInvalidTxField.RefineError("invalid NftIndex")
		}
		logx.Errorf("fail to get nft info: %d, err: %s", txInfo.NftIndex, err.Error())
		return "", err
	}
	if nftInfo.OwnerAccountIndex != txInfo.FromAccountIndex {
		logx.Errorf("not nft owner")
		return "", errorcode.RpcErrInvalidTxField.RefineError("not nft owner")
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify tx
	txDetails, err = txVerification.VerifyTransferNftTxInfo(
		accountInfoMap,
		nftInfo,
		txInfo,
	)
	if err != nil {
		return "", errorcode.RpcErrVerification.RefineError(err)
	}

	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.RpcErrInternal
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

	if err := commglobalmap.DeleteLatestNftInfoForReadInCache(ctx, txInfo.NftIndex); err != nil {
		logx.Errorf("[DeleteLatestNftInfoForReadInCache] param: %d, err: %s", txInfo.NftIndex, err.Error())
		return "", err
	}
	if err := svcCtx.MempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{mempoolTx}); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = CreateFailTx(svcCtx.FailTxModel, commonTx.TxTypeTransferNft, txInfo, err)
		return "", errorcode.RpcErrInternal
	}
	// update cache, not key logic
	if err := commglobalmap.SetLatestNftInfoForReadInCache(ctx, txInfo.NftIndex); err != nil {
		logx.Errorf("[SetLatestNftInfoForReadInCache] param: %d, err: %s", txInfo.NftIndex, err.Error())
	}
	return txId, nil
}
