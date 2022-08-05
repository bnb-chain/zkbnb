package info

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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

func (l *GetAccountsLogic) GetAccounts(req *types.ReqGetAccounts) (*types.RespGetAccounts, error) {
	accounts, err := l.svcCtx.AccountModel.GetAccountsList(int(req.Limit), int64(req.Offset))
	if err != nil {
		logx.Errorf("[GetAccountsList] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	total, err := l.svcCtx.AccountModel.GetAccountsTotalCount()
	if err != nil {
		logx.Errorf("[GetAccountsTotalCount] err: %s", err.Error())
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetAccounts{
		Total:    uint32(total),
		Accounts: make([]*types.Accounts, 0),
	}
	for _, a := range accounts {
		resp.Accounts = append(resp.Accounts, &types.Accounts{
			AccountIndex: uint32(a.AccountIndex),
			AccountName:  a.AccountName,
			PublicKey:    a.PublicKey,
		})
	}
	return resp, nil
}
