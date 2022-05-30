package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/accounthistory"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByAccountNameLogic struct {
	logx.Logger
	ctx            context.Context
	svcCtx         *svc.ServiceContext
	accountHistory accounthistory.AccountHistory
	l2asset        l2asset.L2asset
	globalRPC      globalrpc.GlobalRPC
	account        account.AccountModel
}

func NewGetAccountInfoByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAccountNameLogic {
	return &GetAccountInfoByAccountNameLogic{
		Logger:         logx.WithContext(ctx),
		ctx:            ctx,
		svcCtx:         svcCtx,
		accountHistory: accounthistory.New(svcCtx.Config),
		l2asset:        l2asset.New(svcCtx.Config),
		globalRPC:      globalrpc.New(svcCtx.Config, ctx),
		account:        account.New(svcCtx.Config),
	}
}

func (l *GetAccountInfoByAccountNameLogic) GetAccountInfoByAccountName(req *types.ReqGetAccountInfoByAccountName) (resp *types.RespGetAccountInfoByAccountName, err error) {
	resp.AssetsAccount = make([]*types.Asset, 0)
	if utils.CheckAccountName(req.AccountName) {
		logx.Error("[CheckAccountName] req.AccountName:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	accountName := utils.FormatSting(req.AccountName)
	if utils.CheckFormatAccountName(accountName) {
		logx.Error("[CheckFormatAccountName] accountName:%v", accountName)
		return nil, errcode.ErrInvalidParam
	}
	account, err := l.account.GetAccountByAccountName(accountName)
	if err != nil {
		logx.Error("[GetAccountByAccountName] accountName:%v, err:%v", accountName, err)
		return nil, err
	}
	accountPk, assetsAccount, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Error("[getLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	resp = &types.RespGetAccountInfoByAccountName{
		AccountIndex:  uint32(account.AccountIndex),
		AccountPk:     accountPk,
		AssetsAccount: assetsAccount,
	}
	return resp, nil
}
