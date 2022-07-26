package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
		logx.Errorf("[ParseCreateCollectionTxInfo] err:%v", err)
		return nil, err
	}
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return nil, err
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.AccountIndex))
	if err != nil {
		logx.Errorf("[CheckRequestParam] err:%v", err)
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.GasAccountIndex))
	if err != nil {
		logx.Errorf("[CheckRequestParam] err:%v", err)
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		logx.Errorf("[checkGasAccountIndex] err: %v", err)
		return nil, err
	}
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendCreateCollectionTx] invalid time stamp")
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.AccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.AccountIndex) //get  CollectionNonce + 1
	if err != nil {
		logx.Errorf("[sendCreateCollectionTx] unable to get account info: %s", err.Error())
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendCreateCollectionTx] unable to get account info: %s", err.Error())
			return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
		}
	}
	txInfo.CollectionId = accountInfoMap[txInfo.AccountIndex].CollectionNonce
	var txDetails []*mempool.MempoolTxDetail
	txDetails, err = txVerification.VerifyCreateCollectionTxInfo(accountInfoMap, txInfo)
	if err != nil {
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	_, mempoolTx := ConstructMempoolTx(
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
		l.createFailCreateCollectionTx(txInfo, err.Error())
		return nil, err
	}
	return &globalRPCProto.RespSendCreateCollectionTx{CollectionId: txInfo.CollectionId}, nil
}

func (l *SendCreateCollectionTxLogic) createFailCreateCollectionTx(info *commonTx.CreateCollectionTxInfo, extraInfo string) error {
	txInfo, err := json.Marshal(info)
	if err != nil {
		logx.Errorf("[Marshal] err:%v", err)
		return err
	}
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: util.RandomUUID(),
		// transaction type
		TxType: commonTx.TxTypeCreateCollection,
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
		NativeAddress: "0x00",
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
		// native memo info
		Memo: "",
	}
	return l.svcCtx.FailTxModel.CreateFailTx(failTx)
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
	_, err = svcCtx.RedisConnection.Del(keys...)
	if err != nil {
		logx.Errorf("[CreateMempoolTx] error with redis: %s", err.Error())
		return err
	}
	// check collectionId exist
	exist, err := svcCtx.CollectionModel.IfCollectionExistsByCollectionId(nftCollectionInfo.CollectionId)
	if err != nil {
		errInfo := fmt.Sprintf("[createMempoolTxForCreateCollection] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	if exist {
		return errors.New("[createMempoolTxForCreateCollection] collectionId duplicate creation")
	}

	// write into mempool
	err = svcCtx.MempoolModel.CreateMempoolTxAndL2CollectionAndNonce(nMempoolTx, nftCollectionInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[createMempoolTxForCreateCollection] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
