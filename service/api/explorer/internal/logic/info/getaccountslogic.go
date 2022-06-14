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
	accounts, e := l.svcCtx.Account.GetAccountsList(int(req.Limit), int64(req.Offset))
	if e != nil {
		err = fmt.Errorf("[explorer.info.GetAccountsList]%s", e.Error())
		l.Error(err)
		return
	}

	total, e := l.svcCtx.Account.GetAccountsTotalCount()
	if e != nil {
		err = fmt.Errorf("[explorer.info.GetAccountsList]%s", e.Error())
		l.Error(err)
		return
	}
	resp.Total = uint32(total)

	for _, a := range accounts {
		resp.Accounts = append(resp.Accounts, &types.Accounts{
			AccountIndex: uint32(a.AccountIndex),
			AccountName:  a.AccountName,
			PublicKey:    a.PublicKey,
		})
	}

	return
}
