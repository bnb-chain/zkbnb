package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountStatusByAccountPkLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	account account.AccountModel
}

func NewGetAccountStatusByAccountPkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountStatusByAccountPkLogic {
	return &GetAccountStatusByAccountPkLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		account: account.New(svcCtx),
	}
}

func (l *GetAccountStatusByAccountPkLogic) GetAccountStatusByAccountPk(req *types.ReqGetAccountStatusByAccountPk) (*types.RespGetAccountStatusByAccountPk, error) {
	account, err := l.account.GetAccountByPk(req.AccountPk)
	if err != nil {
		logx.Errorf("[GetAccountByPk] err:%v", err)
		return nil, err
	}
	return &types.RespGetAccountStatusByAccountPk{
		AccountStatus: int64(account.Status),
		AccountIndex:  account.AccountIndex,
		AccountName:   account.AccountName,
	}, nil
}
