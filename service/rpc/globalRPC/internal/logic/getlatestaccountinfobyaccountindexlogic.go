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

type GetLatestAccountInfoByAccountIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestAccountInfoByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAccountInfoByAccountIndexLogic {
	return &GetLatestAccountInfoByAccountIndexLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func packGetLatestAccountInfoByAccountIndexResp(
	status int64,
	msg string,
	err string,
	result *globalRPCProto.ResultGetLatestAccountInfoByAccountIndex,
) (res *globalRPCProto.RespGetLatestAccountInfoByAccountIndex) {
	res = &globalRPCProto.RespGetLatestAccountInfoByAccountIndex{
		Status: status,
		Msg:    msg,
		Err:    err,
		Result: result,
	}
	return res
}

func (l *GetLatestAccountInfoByAccountIndexLogic) GetLatestAccountInfoByAccountIndex(in *globalRPCProto.ReqGetLatestAccountInfoByAccountIndex) (*globalRPCProto.RespGetLatestAccountInfoByAccountIndex, error) {
	var (
		respResult *globalRPCProto.ResultGetLatestAccountInfoByAccountIndex
	)

	err := utils.CheckRequestParam(utils.TypeAccountIndex, reflect.ValueOf(in.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountInfoByAccountIndex] %s", err)
		logx.Error(errInfo)
		return packGetLatestAccountInfoByAccountIndexResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	// in.AccountIndex
	accountInfo, err := l.svcCtx.AccountHistoryModel.GetAccountByAccountIndex(int64(in.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountInfoByAccountIndex] => [AccountModel.GetAccountByAccountIndex] :%s. Invalid AccountIndex: %v ", err.Error(), in.AccountIndex)
		logx.Error(errInfo)
		return packGetLatestAccountInfoByAccountIndexResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	l2AssetsList, err := l.svcCtx.L2AssetModel.GetL2AssetsList()
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountInfoByAccountIndex] => [L2AssetModel.GetL2AssetsList] :%s. ", err.Error())
		logx.Error(errInfo)
		return packGetLatestAccountInfoByAccountIndexResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	respResult = &globalRPCProto.ResultGetLatestAccountInfoByAccountIndex{
		AccountIndex: uint32(accountInfo.AccountIndex),
		AccountName:  accountInfo.AccountName,
		AccountPk:    accountInfo.PublicKey,
	}

	for _, v := range l2AssetsList {

		accountSingleAssetA, err := GetLatestSingleAccountAsset(l.svcCtx, uint32(accountInfo.AccountIndex), uint32(v.L2AssetId))
		if err != nil {
			errInfo := fmt.Sprintf("[logic.GetLatestAccountInfoByAccountIndex] => [GetLatestSingleAccountAsset] :%s. Invalid AccountIndex/AssetId: %v/%v ",
				err.Error(), uint32(accountInfo.AccountIndex), uint32(v.L2AssetId))
			return packGetLatestAccountInfoByAccountIndexResp(FailStatus, FailMsg, errInfo, respResult), nil
		}

		respResult.AssetResultAssets = append(respResult.AssetResultAssets,
			&globalRPCProto.AssetResult{
				AssetId:    uint32(v.L2AssetId),
				BalanceEnc: accountSingleAssetA.BalanceEnc,
			})

	}

	return packGetLatestAccountInfoByAccountIndexResp(SuccessStatus, SuccessMsg, "", respResult), nil
}
