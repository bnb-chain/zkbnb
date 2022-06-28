package logic

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"time"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
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
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[CheckRequestParam] err:%v", err)
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendCreateCollectionTx] invalid gas account index")
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendCreateCollectionTx] invalid time stamp")
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.AccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.AccountIndex)
	if err != nil {
		logx.Errorf("[sendCreateCollectionTx] unable to get account info: %s", err.Error())
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	// get account info by gas index
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		// get account info by gas index
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendCreateCollectionTx] unable to get account info: %s", err.Error())
			return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
		}
	}
	txInfo.CollectionId = accountInfoMap[txInfo.AccountIndex].CollectionNonce
	var (
		txDetails []*mempool.MempoolTxDetail
	)
	txDetails, err = txVerification.VerifyCreateCollectionTxInfo(accountInfoMap, txInfo)
	if err != nil {
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	txId, mempoolTx := ConstructMempoolTx(
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
	err = CreateMempoolTx(mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	collectionId, err := strconv.ParseInt(txId, 10, 64)
	if err != nil {
		return nil, l.createFailCreateCollectionTx(txInfo, err.Error())
	}
	return &globalRPCProto.RespSendCreateCollectionTx{CollectionId: collectionId}, nil
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
