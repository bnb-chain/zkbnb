package info

import (
	"context"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext

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
	resp := &types.RespGetAccounts{}
	accounts, e := l.account.GetAccountsList(int(req.Limit), int64(req.Offset))
	if e != nil {
		err := fmt.Errorf("[explorer.info.GetAccountsList]%s", e.Error())
		l.Error(err)
		return nil, err
	}

	total, e := l.account.GetAccountsTotalCount()
	if e != nil {
		err := fmt.Errorf("[explorer.info.GetAccountsList]%s", e.Error())
		l.Error(err)
		return nil, err
	}
	resp.Total = uint32(total)

	for _, a := range accounts {
		resp.Accounts = append(resp.Accounts, &types.Accounts{
			AccountIndex: uint32(a.AccountIndex),
			AccountName:  a.AccountName,
			PublicKey:    a.PublicKey,
		})
	}

	return resp, nil
}
