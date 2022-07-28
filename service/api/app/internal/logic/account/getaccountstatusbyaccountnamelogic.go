package account

import (
	"context"
	"github.com/bnb-chain/zkbas/common/checker"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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
		logx.Errorf("[CheckAccountIndex] param:%v", req.AccountName)
		return nil, errorcode.AppErrInvalidParam
	}
	account, err := l.account.GetBasicAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[GetBasicAccountByAccountName] err:%v", err)
		return nil, err
	}
	resp = &types.RespGetAccountStatusByAccountName{
		AccountStatus: uint32(account.Status),
		AccountPk:     account.PublicKey,
		AccountIndex:  uint32(account.AccountIndex),
	}
	return resp, nil
}
