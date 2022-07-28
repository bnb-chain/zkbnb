package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
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
		logx.Errorf("[ParseMintNftTxInfo] err:%v", err)
		return nil, err
	}
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return nil, err
	}
	if txInfo.NftCollectionId == commonConstant.NilCollectionId {
		return nil, l.createFailMintNftTx(txInfo, "nft collection id is nil")
	}
	accountInfo, err := l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.CreatorAccountIndex)
	if err != nil {
		return nil, l.createFailMintNftTx(txInfo, err.Error())
	}
	if accountInfo.CollectionNonce < txInfo.NftCollectionId {
		return nil, l.createFailMintNftTx(txInfo, "nft collection id is less than account nonce")
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.CreatorAccountIndex))
	if err != nil {
		return nil, l.createFailMintNftTx(txInfo, err.Error())
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.ToAccountIndex))
	if err != nil {
		return nil, l.createFailMintNftTx(txInfo, err.Error())
	}
	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		logx.Errorf("[checkGasAccountIndex] err: %v", err)
		return nil, err
	}
	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendMintNftTx] invalid time stamp")
		return nil, l.createFailMintNftTx(txInfo, errors.New("[sendMintNftTx] invalid time stamp").Error())
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
		nftIndex       int64
		redisLock      *redis.RedisLock
	)
	redisLock, nftIndex, err = globalmapHandler.GetLatestNftIndexForWrite(l.svcCtx.NftModel, l.svcCtx.RedisConnection)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to get latest nft index: %s", err.Error())
		return nil, err
	}
	defer redisLock.Release()
	accountInfoMap[txInfo.CreatorAccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.CreatorAccountIndex)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to get account info: %s", err.Error())
		return nil, l.createFailMintNftTx(txInfo, err.Error())
	}
	// get account info by to index
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.ToAccountIndex)
		if err != nil {
			logx.Errorf("[sendMintNftTx] unable to get account info: %s", err.Error())
			return nil, l.createFailMintNftTx(txInfo, err.Error())
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("[sendMintNftTx] invalid account name")
		return nil, l.createFailMintNftTx(txInfo, errors.New("[sendMintNftTx] invalid account name").Error())
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendMintNftTx] unable to get account info: %s", err.Error())
			return nil, l.createFailMintNftTx(txInfo, err.Error())
		}
	}
	var (
		txDetails []*mempool.MempoolTxDetail
	)
	txInfo.NftIndex = nftIndex
	txDetails, err = txVerification.VerifyMintNftTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return nil, l.createFailMintNftTx(txInfo, err.Error())
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
		return nil, l.createFailMintNftTx(txInfo, err.Error())
	}
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return nil, l.createFailMintNftTx(txInfo, err.Error())
	}
	_, mempoolTx := ConstructMempoolTx(
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
	err = createMempoolTxForMintNft(nftInfo, mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return nil, l.createFailMintNftTx(txInfo, err.Error())
	}
	// update redis
	var formatNftInfo *commonAsset.NftInfo
	for _, txDetail := range mempoolTx.MempoolDetails {
		if txDetail.AssetType == commonAsset.NftAssetType {
			formatNftInfo, err = commonAsset.ParseNftInfo(txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[sendMintNftTx] unable to parse nft info: %s", err.Error())
				return nil, err
			}
		}
	}
	nftInfoBytes, err := json.Marshal(formatNftInfo)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to marshal: %s", err.Error())
		return nil, err
	}
	_ = l.svcCtx.RedisConnection.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)

	return &globalRPCProto.RespSendMintNftTx{NftIndex: txInfo.NftIndex}, nil
}

func (l *SendMintNftTxLogic) createFailMintNftTx(info *commonTx.MintNftTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailMintNftTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeMintNft,
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

	err = l.svcCtx.FailTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailMintNftTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}

func createMempoolTxForMintNft(
	nftInfo *nft.L2Nft,
	nMempoolTx *mempool.MempoolTx,
	redisConnection *redis.Redis,
	mempoolModel mempool.MempoolModel,
) (err error) {
	var keys []string
	for _, mempoolTxDetail := range nMempoolTx.MempoolDetails {
		keys = append(keys, util.GetAccountKey(mempoolTxDetail.AccountIndex))
	}
	_, err = redisConnection.Del(keys...)
	if err != nil {
		logx.Errorf("[CreateMempoolTx] error with redis: %s", err.Error())
		return err
	}
	// write into mempool
	err = mempoolModel.CreateMempoolTxAndL2Nft(nMempoolTx, nftInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[CreateMempoolTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
