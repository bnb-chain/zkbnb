package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountStatusByAccountPkLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	account account.Model
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
	//TODO: check pk
	account, err := l.account.GetBasicAccountByAccountPk(l.ctx, req.AccountPk)
	if err != nil {
		logx.Errorf("[GetBasicAccountByAccountPk] err:%v", err)
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	return &types.RespGetAccountStatusByAccountPk{
		AccountStatus: int64(account.Status),
		AccountIndex:  account.AccountIndex,
		AccountName:   account.AccountName,
	}, nil
}
