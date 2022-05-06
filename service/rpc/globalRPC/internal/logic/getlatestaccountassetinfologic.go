package logic

import (
	"context"
	"fmt"
	"reflect"

	"github.com/zecrey-labs/zecrey/common/utils"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/svc"

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

	err := utils.CheckRequestParam(utils.TypeAccountIndex, reflect.ValueOf(in.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountAssetInf] %s", err)
		logx.Error(errInfo)
		return packGetLatestAccountAssetInfoResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	err = utils.CheckRequestParam(utils.TypeAssetId, reflect.ValueOf(in.AssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountAssetInf] %s", err)
		logx.Error(errInfo)
		return packGetLatestAccountAssetInfoResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	accountInfo, err := l.svcCtx.AccountHistoryModel.GetAccountByAccountIndex(int64(in.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountAssetInfo] => [AccountModel.GetAccountByAccountIndex] :%s. Invalid AccountIndex: %v ", err.Error(), in.AccountIndex)
		logx.Error(errInfo)
		return packGetLatestAccountAssetInfoResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	accountSingleAssetA, err := GetLatestSingleAccountAsset(l.svcCtx, uint32(in.AccountIndex), uint32(in.AssetId))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountAssetInfo] => [GetLatestSingleAccountAsset] :%s. Invalid AccountIndex/AssetId: %v/%v ",
			err.Error(), uint32(in.AccountIndex), uint32(in.AssetId))
		return packGetLatestAccountAssetInfoResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	respResult = &globalRPCProto.ResultGetLatestAccountAssetInfo{
		AccountIndex: uint32(in.AccountIndex),
		AccountName:  accountInfo.AccountName,
		AccountPk:    accountInfo.PublicKey,
		AssetResultAssets: &globalRPCProto.AssetResult{
			AssetId:    uint32(in.AssetId),
			BalanceEnc: accountSingleAssetA.BalanceEnc,
		},
	}

	return packGetLatestAccountAssetInfoResp(SuccessStatus, SuccessMsg, "", respResult), nil
}
