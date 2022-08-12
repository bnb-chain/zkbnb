package account

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

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
	accounts, err := l.svcCtx.AccountModel.GetAccountsList(int(req.Limit), int64(req.Offset))
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	total, err := l.svcCtx.AccountModel.GetAccountsTotalCount()
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	resp = &types.RespGetAccounts{
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
