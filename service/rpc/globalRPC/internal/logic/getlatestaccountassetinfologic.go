package logic

import (
	"context"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"reflect"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestAccountAssetInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestAccountAssetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAccountAssetInfoLogic {
	return &GetLatestAccountAssetInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func packGetLatestAccountAssetInfoResp(
	status int64,
	msg string,
	err string,
	result *globalRPCProto.ResultGetLatestAccountAssetInfo,
) (res *globalRPCProto.RespGetLatestAccountAssetInfo) {
	res = &globalRPCProto.RespGetLatestAccountAssetInfo{
		Status: status,
		Msg:    msg,
		Err:    err,
		Result: result,
	}
	return res
}

func (l *GetLatestAccountAssetInfoLogic) GetLatestAccountAssetInfo(in *globalRPCProto.ReqGetLatestAccountAssetInfo) (*globalRPCProto.RespGetLatestAccountAssetInfo, error) {
	var (
		respResult *globalRPCProto.ResultGetLatestAccountAssetInfo
	)

	err := util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(in.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountAssetInf] %s", err)
		logx.Error(errInfo)
		return packGetLatestAccountAssetInfoResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(in.AssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountAssetInf] %s", err)
		logx.Error(errInfo)
		return packGetLatestAccountAssetInfoResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	accountInfo, err := globalmapHandler.GetLatestAccountInfo(
		l.svcCtx.AccountModel,
		l.svcCtx.AccountHistoryModel,
		l.svcCtx.MempoolModel,
		l.svcCtx.LiquidityPairModel,
		l.svcCtx.RedisConnection,
		int64(in.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountAssetInfo] => [AccountModel.GetAccountByAccountIndex] :%s. Invalid AccountIndex: %v ", err.Error(), in.AccountIndex)
		logx.Error(errInfo)
		return packGetLatestAccountAssetInfoResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	respResult = &globalRPCProto.ResultGetLatestAccountAssetInfo{
		AccountIndex: uint32(in.AccountIndex),
		AccountName:  accountInfo.AccountName,
		AccountPk:    accountInfo.PublicKey,
		Nonce:        accountInfo.Nonce,
		AssetResultAssets: &globalRPCProto.AssetResult{
			AssetId: uint32(in.AssetId),
			Balance: accountInfo.AssetInfo[int64(in.AssetId)],
		},
	}

	return packGetLatestAccountAssetInfoResp(SuccessStatus, SuccessMsg, "", respResult), nil
}
