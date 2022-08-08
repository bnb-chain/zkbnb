package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
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
	if utils.CheckAccountPk(req.AccountPk) {
		logx.Errorf("invalid AccountPk: %s", req.AccountPk)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountPk")
	}

	account, err := l.svcCtx.AccountModel.GetAccountByPk(req.AccountPk)
	if err != nil {
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
