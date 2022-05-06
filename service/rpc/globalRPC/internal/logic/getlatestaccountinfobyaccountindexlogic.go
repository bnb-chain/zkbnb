package logic

import (
	"context"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"reflect"

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

	err := util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(in.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestAccountInfoByAccountIndex] %s", err)
		logx.Error(errInfo)
		return packGetLatestAccountInfoByAccountIndexResp(FailStatus, FailMsg, errInfo, respResult), nil
	}

	// in.AccountIndex
	accountInfo, err := globalmapHandler.GetLatestAccountInfo(l.svcCtx, int64(in.AccountIndex))
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
		Nonce:        accountInfo.Nonce,
	}

	for _, v := range l2AssetsList {

		accountSingleAssetA, err := globalmapHandler.GetLatestAsset(l.svcCtx, accountInfo.AccountIndex, v.AssetId)
		if err != nil {
			errInfo := fmt.Sprintf("[logic.GetLatestAccountInfoByAccountIndex] => [GetLatestSingleAccountAsset] :%s. Invalid AccountIndex/AssetId: %v/%v ",
				err.Error(), uint32(accountInfo.AccountIndex), uint32(v.AssetId))
			return packGetLatestAccountInfoByAccountIndexResp(FailStatus, FailMsg, errInfo, respResult), nil
		}

		respResult.AssetResultAssets = append(respResult.AssetResultAssets,
			&globalRPCProto.AssetResult{
				AssetId: uint32(v.AssetId),
				Balance: accountSingleAssetA.Balance,
			})

	}

	return packGetLatestAccountInfoByAccountIndexResp(SuccessStatus, SuccessMsg, "", respResult), nil
}
