package logic

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/logic/sendrawtx"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type SendCreateCollectionTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewSendCreateCollectionTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCreateCollectionTxLogic {
	return &SendCreateCollectionTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *SendCreateCollectionTxLogic) SendCreateCollectionTx(in *globalRPCProto.ReqSendCreateCollectionTx) (*globalRPCProto.RespSendCreateCollectionTx, error) {
	txInfo, err := commonTx.ParseCreateCollectionTxInfo(in.TxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return nil, errorcode.RpcErrInvalidTx
	}

	if err := legendTxTypes.ValidateCreateCollectionTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return nil, errorcode.RpcErrInvalidTxField.RefineError(err)
	}

	if err := sendrawtx.CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		return nil, err
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.AccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.AccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.AccountIndex, err.Error())
		return nil, errorcode.RpcErrInternal
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

	txInfo.CollectionId = accountInfoMap[txInfo.AccountIndex].CollectionNonce

	var txDetails []*mempool.MempoolTxDetail
	txDetails, err = txVerification.VerifyCreateCollectionTxInfo(accountInfoMap, txInfo)
	if err != nil {
		return nil, errorcode.RpcErrVerification.RefineError(err)
	}

	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errorcode.RpcErrInternal
	}
	_, mempoolTx := sendrawtx.ConstructMempoolTx(
		commonTx.TxTypeCreateCollection,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		commonConstant.NilPairIndex,
		commonConstant.NilAssetId,
		accountInfoMap[txInfo.AccountIndex].AccountName,
		commonConstant.NilL1Address,
		string(txInfoBytes),
		"",
		txInfo.AccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	// construct nft Collection info
	nftCollectionInfo := &nft.L2NftCollection{
		CollectionId: txInfo.CollectionId,
		AccountIndex: txInfo.AccountIndex,
		Name:         txInfo.Name,
		Introduction: txInfo.Introduction,
		Status:       nft.CollectionPending,
	}
	if err = createMempoolTxForCreateCollection(nftCollectionInfo, mempoolTx, l.svcCtx); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = sendrawtx.CreateFailTx(l.svcCtx.FailTxModel, commonTx.TxTypeCreateCollection, txInfo, err)
		return nil, errorcode.RpcErrInternal
	}
	return &globalRPCProto.RespSendCreateCollectionTx{CollectionId: txInfo.CollectionId}, nil
}

func createMempoolTxForCreateCollection(
	nftCollectionInfo *nft.L2NftCollection,
	nMempoolTx *mempool.MempoolTx,
	svcCtx *svc.ServiceContext,
) (err error) {
	var keys []string
	for _, mempoolTxDetail := range nMempoolTx.MempoolDetails {
		keys = append(keys, util.GetAccountKey(mempoolTxDetail.AccountIndex))
	}
	if _, err := svcCtx.RedisConnection.Del(keys...); err != nil {
		logx.Errorf("fail to delete keys from redis: %s", err.Error())
		return err
	}
	// check collectionId exist
	exist, err := svcCtx.CollectionModel.IfCollectionExistsByCollectionId(nftCollectionInfo.CollectionId)
	if err != nil {
		return err
	}
	if exist {
		logx.Errorf("collectionId duplicate creation: %d", nftCollectionInfo.CollectionId)
		return errors.New("collectionId duplicate creation")
	}

	// write into mempool
	if err := svcCtx.MempoolModel.CreateMempoolTxAndL2CollectionAndNonce(nMempoolTx, nftCollectionInfo); err != nil {
		return err
	}
	return nil
}
