package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountStatusByAccountPkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountStatusByAccountPkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountStatusByAccountPkLogic {
	return &GetAccountStatusByAccountPkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountStatusByAccountPkLogic) GetAccountStatusByAccountPk(req *types.ReqGetAccountStatusByAccountPk) (*types.RespGetAccountStatusByAccountPk, error) {
	//TODO: check pk
	account, err := l.svcCtx.AccountModel.GetAccountByPk(req.AccountPk)
	if err != nil {
		logx.Errorf("[GetBasicAccountByAccountPk] err: %s", err.Error())
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
