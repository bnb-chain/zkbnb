package sendrawtx

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

func SendWithdrawNftTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseWithdrawNftTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.RpcErrInvalidTx
	}

	if err := legendTxTypes.ValidateWithdrawNftTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.RpcErrInvalidTxField.RefineError(err)
	}

	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, svcCtx.SysConfigModel); err != nil {
		return "", err
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.AccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.AccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.RpcErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.AccountIndex, err.Error())
		return "", errorcode.RpcErrInternal
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
	if accountInfoMap[txInfo.CreatorAccountIndex] == nil {
		accountInfoMap[txInfo.CreatorAccountIndex], err = commglobalmap.GetBasicAccountInfo(ctx, txInfo.CreatorAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.RpcErrInvalidTxField.RefineError("invalid CreatorAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.CreatorAccountIndex, err.Error())
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
	if nftInfo.OwnerAccountIndex != txInfo.AccountIndex {
		logx.Errorf("not nft owner")
		return "", errorcode.RpcErrInvalidTxField.RefineError("not nft owner")
	}

	txInfo.CreatorAccountIndex = nftInfo.CreatorAccountIndex
	txInfo.CreatorAccountNameHash = common.FromHex(accountInfoMap[nftInfo.CreatorAccountIndex].AccountNameHash)
	txInfo.CreatorTreasuryRate = nftInfo.CreatorTreasuryRate
	txInfo.NftContentHash = common.FromHex(nftInfo.NftContentHash)
	txInfo.NftL1Address = nftInfo.NftL1Address
	txInfo.NftL1TokenId, _ = new(big.Int).SetString(nftInfo.NftL1TokenId, 10)
	txInfo.CollectionId = nftInfo.CollectionId

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify tx
	txDetails, err = txVerification.VerifyWithdrawNftTxInfo(
		accountInfoMap,
		nftInfo,
		txInfo,
	)
	if err != nil {
		return "", errorcode.RpcErrVerification.RefineError(err)
	}

	// delete key
	key := util.GetNftKeyForRead(txInfo.NftIndex)
	_, err = svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("unable to delete key from redis, key: %s, err: %s", key, err.Error())
		return "", errorcode.RpcErrInternal
	}
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.RpcErrInternal
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeWithdrawNft,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		txInfo.NftIndex,
		commonConstant.NilPairIndex,
		commonConstant.NilAssetId,
		commonConstant.NilAssetAmountStr,
		"",
		string(txInfoBytes),
		"",
		txInfo.AccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	err = CreateMempoolTx(mempoolTx, svcCtx.RedisConnection, svcCtx.MempoolModel)
	if err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = CreateFailTx(svcCtx.FailTxModel, commonTx.TxTypeCancelOffer, txInfo, err)
		return "", errorcode.RpcErrInternal
	}
	// update redis
	var formatNftInfo *commonAsset.NftInfo
	for _, txDetail := range mempoolTx.MempoolDetails {
		if txDetail.AssetType == commonAsset.NftAssetType {
			formatNftInfo, err = commonAsset.ParseNftInfo(txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("unable to parse nft info: %s", err.Error())
				return txId, nil
			}
		}
	}
	nftInfoBytes, err := json.Marshal(formatNftInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return txId, nil
	}
	_ = svcCtx.RedisConnection.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)

	return txId, nil
}
