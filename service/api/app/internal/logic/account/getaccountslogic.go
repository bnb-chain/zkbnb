package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
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

func (l *GetAccountsLogic) GetAccounts(req *types.ReqGetRange) (resp *types.Accounts, err error) {
	total, err := l.svcCtx.MemCache.GetAccountTotalCountWiltFallback(func() (interface{}, error) {
		return l.svcCtx.AccountModel.GetAccountsTotalCount()
	})
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	resp = &types.Accounts{
		Accounts: make([]*types.SimpleAccount, 0),
		Total:    uint32(total),
	}

	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	accounts, err := l.svcCtx.AccountModel.GetAccountsList(int(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, errorcode.AppErrInternal
	}
	for _, a := range accounts {
		resp.Accounts = append(resp.Accounts, &types.SimpleAccount{
			Index: a.AccountIndex,
			Name:  a.AccountName,
			Pk:    a.PublicKey,
		})
	}
	return resp, nil
}
