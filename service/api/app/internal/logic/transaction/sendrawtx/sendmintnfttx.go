package sendrawtx

import (
	"context"
	"encoding/json"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

func SendMintNftTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseMintNftTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.RpcErrInvalidTx
	}

	if err := legendTxTypes.ValidateMintNftTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.RpcErrInvalidTxField.RefineError(err)
	}

	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, svcCtx.SysConfigModel); err != nil {
		return "", err
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
		nftIndex       int64
		redisLock      *redis.RedisLock
	)

	accountInfoMap[txInfo.CreatorAccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.CreatorAccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.RpcErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.CreatorAccountIndex, err.Error())
		return "", errorcode.RpcErrInternal
	}
	if accountInfoMap[txInfo.CreatorAccountIndex].CollectionNonce < txInfo.NftCollectionId {
		logx.Errorf("collection id %d is greater than collection nonce %d",
			txInfo.NftCollectionId, accountInfoMap[txInfo.CreatorAccountIndex].CollectionNonce)
		return "", errorcode.RpcErrInvalidTxField.RefineError("invalid NftCollectionId")
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

	redisLock, nftIndex, err = globalmapHandler.GetLatestNftIndexForWrite(svcCtx.NftModel, svcCtx.RedisConn)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to get latest nft index: %s", err.Error())
		return "", err
	}
	defer redisLock.Release()

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	txInfo.NftIndex = nftIndex
	txDetails, err = txVerification.VerifyMintNftTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return "", errorcode.RpcErrVerification.RefineError(err)
	}

	// construct nft info
	nftInfo := &nft.L2Nft{
		NftIndex:            nftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		NftL1Address:        commonConstant.NilL1Address,
		NftL1TokenId:        commonConstant.NilL1TokenId,
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.NftCollectionId,
	}
	// delete key
	key := util.GetNftKeyForRead(nftIndex)
	_, err = svcCtx.RedisConn.Del(key)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to delete key from redis: %s", err.Error())
		return "", errorcode.RpcErrInternal
	}
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.RpcErrInternal
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeMintNft,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		nftIndex,
		commonConstant.NilPairIndex,
		commonConstant.NilAssetId,
		commonConstant.NilAssetAmountStr,
		"",
		string(txInfoBytes),
		"",
		txInfo.CreatorAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)

	if err := svcCtx.MempoolModel.CreateMempoolTxAndL2Nft(mempoolTx, nftInfo); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = CreateFailTx(svcCtx.FailTxModel, commonTx.TxTypeMintNft, txInfo, err)
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
	_ = svcCtx.RedisConn.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)

	return txId, nil
}
