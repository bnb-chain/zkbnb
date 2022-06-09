package info

import (
	"context"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountsLogic {
	return &GetAccountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountsLogic) GetAccounts(req *types.ReqGetAccounts) (resp *types.RespGetAccounts, err error) {
	accounts, err := l.svcCtx.AccountHistoryModel.GetAccountsList(int(req.Limit), int64(req.Offset))
	if err != nil {
		errInfo := fmt.Sprintf("[explorer.info.GetAccountsList]<=>[AccountModel.GetAccountsList] %s", err.Error())
		logx.Error(errInfo)
		return packGetAccountsListResp(logic.FailStatus, "fail", errInfo, respResult), nil
	}

	total, err := l.svcCtx.AccountHistoryModel.GetAccountsTotalCount()
	if err != nil {
		errInfo := fmt.Sprintf("[explorer.info.GetAccountsList]<=>[AccountModel.GetAccountsTotalCount] %s", err.Error())
		logx.Error(errInfo)
		return packGetAccountsListResp(logic.FailStatus, "fail", errInfo, respResult), nil
	}

	dataAccountsList := make([]*types.DataAccountsList, 0)
	for _, account := range accounts {
		dataAccountsList = append(dataAccountsList, &types.DataAccountsList{
			AccountIndex: uint32(account.AccountIndex),
			AccountName:  account.AccountName,
			PublicKey:    account.PublicKey,
		})
	}
	resp := &types.RespGetAccountsList{
		Status: logic.SuccessStatus,
		Msg:    "success",
		Err:    "",
		Result: types.ResultGetAccountsList{
			Limit:  req.Limit,
			Offset: req.Offset,
			Total:  uint32(total),
			Data:   dataAccountsList,
		},
	}
	return resp, nil
}
