package account

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountStatusByNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountStatusByNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountStatusByNameLogic {
	return &GetAccountStatusByNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountStatusByNameLogic) GetAccountStatusByName(req *types.ReqGetAccountStatusByName) (resp *types.RespGetAccountStatusByName, err error) {
	if !utils.ValidateAccountName(req.AccountName) {
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}

	account, err := l.svcCtx.AccountModel.GetAccountByAccountName(req.AccountName)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp = &types.RespGetAccountStatusByName{
		AccountStatus: uint32(account.Status),
		AccountPk:     account.PublicKey,
		AccountIndex:  uint32(account.AccountIndex),
	}
	return resp, nil
}
