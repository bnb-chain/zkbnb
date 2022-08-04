package logic

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
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/logic/sendrawtx"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type SendMintNftTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewSendMintNftTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMintNftTxLogic {
	return &SendMintNftTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *SendMintNftTxLogic) SendMintNftTx(in *globalRPCProto.ReqSendMintNftTx) (*globalRPCProto.RespSendMintNftTx, error) {
	txInfo, err := commonTx.ParseMintNftTxInfo(in.TxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return nil, errorcode.RpcErrInvalidTx
	}

	if err := legendTxTypes.ValidateMintNftTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return nil, errorcode.RpcErrInvalidTxField.RefineError(err)
	}

	if err := sendrawtx.CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		return nil, err
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
		nftIndex       int64
		redisLock      *redis.RedisLock
	)

	accountInfoMap[txInfo.CreatorAccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.CreatorAccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.CreatorAccountIndex, err.Error())
		return nil, errorcode.RpcErrInternal
	}
	if accountInfoMap[txInfo.CreatorAccountIndex].CollectionNonce < txInfo.NftCollectionId {
		logx.Errorf("collection id %d is greater than collection nonce %d",
			txInfo.NftCollectionId, accountInfoMap[txInfo.CreatorAccountIndex].CollectionNonce)
		return nil, errorcode.RpcErrInvalidTxField.RefineError("invalid NftCollectionId")
	}

	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.ToAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return nil, errorcode.RpcErrInvalidTxField.RefineError("invalid ToAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.ToAccountIndex, err.Error())
			return nil, errorcode.RpcErrInternal
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("invalid account name hash, expected: %s, actual: %s", accountInfoMap[txInfo.ToAccountIndex].AccountNameHash, txInfo.ToAccountNameHash)
		return nil, errorcode.RpcErrInvalidTxField.RefineError("invalid ToAccountNameHash")
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.GasAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return nil, errorcode.RpcErrInvalidTxField.RefineError("invalid GasAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.GasAccountIndex, err.Error())
			return nil, errorcode.RpcErrInternal
		}
	}

	redisLock, nftIndex, err = globalmapHandler.GetLatestNftIndexForWrite(l.svcCtx.NftModel, l.svcCtx.RedisConnection)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to get latest nft index: %s", err.Error())
		return nil, err
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
		return nil, errorcode.RpcErrVerification.RefineError(err)
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
	_, err = l.svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to delete key from redis: %s", err.Error())
		return nil, errorcode.RpcErrInternal
	}
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errorcode.RpcErrInternal
	}
	_, mempoolTx := sendrawtx.ConstructMempoolTx(
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

	if err := l.svcCtx.MempoolModel.CreateMempoolTxAndL2Nft(mempoolTx, nftInfo); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = sendrawtx.CreateFailTx(l.svcCtx.FailTxModel, commonTx.TxTypeMintNft, txInfo, err)
		return nil, errorcode.RpcErrInternal
	}

	resp := &globalRPCProto.RespSendMintNftTx{NftIndex: txInfo.NftIndex}

	// update redis
	var formatNftInfo *commonAsset.NftInfo
	for _, txDetail := range mempoolTx.MempoolDetails {
		if txDetail.AssetType == commonAsset.NftAssetType {
			formatNftInfo, err = commonAsset.ParseNftInfo(txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("unable to parse nft info: %s", err.Error())
				return resp, nil
			}
		}
	}
	nftInfoBytes, err := json.Marshal(formatNftInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return resp, nil
	}
	_ = l.svcCtx.RedisConnection.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)

	return resp, nil
}
