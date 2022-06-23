package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountsLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	account account.AccountModel
}

func NewGetAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountsLogic {
	return &GetAccountsLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		account: account.New(svcCtx),
	}
}

func (l *GetAccountsLogic) GetAccounts(req *types.ReqGetAccounts) (*types.RespGetAccounts, error) {
	accounts, err := l.account.GetAccountsList(int(req.Limit), int64(req.Offset))
	if err != nil {
		logx.Errorf("[GetAccountsList] err:%v", err)
		return nil, err
	}
	total, err := l.account.GetAccountsTotalCount()
	if err != nil {
		logx.Errorf("[GetAccountsTotalCount] err:%v", err)
		return nil, err
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
