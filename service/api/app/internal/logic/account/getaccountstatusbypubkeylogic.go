package account

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountStatusByPubKeyLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	account account.AccountModel
}

func NewGetAccountStatusByPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountStatusByPubKeyLogic {
	return &GetAccountStatusByPubKeyLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		account: account.New(svcCtx),
	}
}

func (l *GetAccountStatusByPubKeyLogic) GetAccountStatusByPubKey(req *types.ReqGetAccountStatusByPubKey) (resp *types.RespGetAccountStatusByPubKey, err error) {
	account, err := l.account.GetAccountByPk(req.AccountPk)
	if err != nil {
		logx.Errorf("[GetAccountByPk] err:%v", err)
		return nil, err
	}
	resp = &types.RespGetAccountStatusByPubKey{
		AccountStatus: uint32(account.Status),
		AccountName:   account.AccountName,
		AccountIndex:  uint32(account.AccountIndex),
	}
	return resp, nil
}
