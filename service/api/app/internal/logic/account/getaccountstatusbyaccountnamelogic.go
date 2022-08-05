package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/checker"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountStatusByAccountNameLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	account account.Model
}

func NewGetAccountStatusByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountStatusByAccountNameLogic {
	return &GetAccountStatusByAccountNameLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		account: account.New(svcCtx),
	}
}

func (l *GetAccountStatusByAccountNameLogic) GetAccountStatusByAccountName(req *types.ReqGetAccountStatusByAccountName) (resp *types.RespGetAccountStatusByAccountName, err error) {
	if checker.CheckAccountName(req.AccountName) {
		logx.Errorf("[CheckAccountIndex] param: %s", req.AccountName)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}
	account, err := l.account.GetBasicAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[GetBasicAccountByAccountName] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp = &types.RespGetAccountStatusByAccountName{
		AccountStatus: uint32(account.Status),
		AccountPk:     account.PublicKey,
		AccountIndex:  uint32(account.AccountIndex),
	}
	return resp, nil
}
