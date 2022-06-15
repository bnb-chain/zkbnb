package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/accounthistory"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
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
	globalRPC      globalrpc.GlobalRPC
	account        account.AccountModel
}

func NewGetAccountInfoByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAccountNameLogic {
	return &GetAccountInfoByAccountNameLogic{
		Logger:         logx.WithContext(ctx),
		ctx:            ctx,
		svcCtx:         svcCtx,
		accountHistory: accounthistory.New(svcCtx.Config),
		globalRPC:      globalrpc.New(svcCtx.Config, ctx),
		account:        account.New(svcCtx.Config),
	}
}

func (l *GetAccountInfoByAccountNameLogic) GetAccountInfoByAccountName(req *types.ReqGetAccountInfoByAccountName) (resp *types.RespGetAccountInfoByAccountName, err error) {
	resp.AssetsAccount = make([]*types.Asset, 0)
	if utils.CheckAccountName(req.AccountName) {
		logx.Errorf("[CheckAccountName] req.AccountName:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	accountName := utils.FormatSting(req.AccountName)
	if utils.CheckFormatAccountName(accountName) {
		logx.Errorf("[CheckFormatAccountName] accountName:%v", accountName)
		return nil, errcode.ErrInvalidParam
	}
	account, err := l.account.GetAccountByAccountName(accountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] accountName:%v, err:%v", accountName, err)
		return nil, err
	}
	assets, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(uint32(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	for _, asset := range assets {
		resp.AssetsAccount = append(resp.AssetsAccount, &types.Asset{
			AssetId: asset.AssetId,
			Balance: asset.Balance,
		})
	}
	resp.AccountIndex = uint32(account.AccountIndex)
	resp.AccountPk = account.PublicKey
	return resp, nil
}
