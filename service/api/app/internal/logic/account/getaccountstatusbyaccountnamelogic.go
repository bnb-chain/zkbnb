package account

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/errcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
	"github.com/bnb-chain/zkbas/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountStatusByAccountNameLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	account account.AccountModel
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
	if utils.CheckAccountName(req.AccountName) {
		logx.Errorf("[CheckAccountIndex] param:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	accountInfo, err := l.account.GetAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}
	resp = &types.RespGetAccountStatusByAccountName{
		AccountStatus: uint32(accountInfo.Status),
		AccountPk:     accountInfo.PublicKey,
		AccountIndex:  uint32(accountInfo.AccountIndex),
	}
	return resp, nil
}
