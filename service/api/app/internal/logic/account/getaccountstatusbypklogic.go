package account

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountStatusByPkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountStatusByPkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountStatusByPkLogic {
	return &GetAccountStatusByPkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountStatusByPkLogic) GetAccountStatusByPk(req *types.ReqGetAccountStatusByPk) (resp *types.RespGetAccountStatusByPk, err error) {
	if !utils.ValidateAccountPk(req.AccountPk) {
		logx.Errorf("invalid AccountPk: %s", req.AccountPk)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountPk")
	}

	accountIndex, err := l.svcCtx.MemCache.GetAccountIndexByName(req.AccountPk)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	account, err := l.svcCtx.MemCache.GetAccountWithFallback(accountIndex, func() (interface{}, error) {
		return l.svcCtx.AccountModel.GetAccountByAccountIndex(accountIndex)
	})
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	return &types.RespGetAccountStatusByPk{
		AccountStatus: int64(account.Status),
		AccountIndex:  account.AccountIndex,
		AccountName:   account.AccountName,
	}, nil
}
