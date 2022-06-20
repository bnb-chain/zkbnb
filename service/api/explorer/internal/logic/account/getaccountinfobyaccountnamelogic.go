package account

import (
	"context"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByAccountNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext

	account   account.AccountModel
	globalRPC globalrpc.GlobalRPC
}

func NewGetAccountInfoByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAccountNameLogic {
	return &GetAccountInfoByAccountNameLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		account:   account.New(svcCtx),
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetAccountInfoByAccountNameLogic) GetAccountInfoByAccountName(req *types.ReqGetAccountInfoByAccountName) (*types.RespGetAccountInfoByAccountName, error) {
	resp := &types.RespGetAccountInfoByAccountName{}
	if utils.CheckAccountName(req.AccountName) {
		err := fmt.Errorf("[CheckAccountName] req.AccountName:%v", req.AccountName)
		l.Error(err)
		return nil, err
	}
	accountName := utils.FormatSting(req.AccountName)
	if utils.CheckFormatAccountName(accountName) {
		err := fmt.Errorf("[CheckFormatAccountName] accountName:%v", accountName)
		l.Error(err)
		return nil, err
	}
	account, err := l.account.GetAccountByAccountName(l.ctx, accountName)
	if err != nil {
		err = fmt.Errorf("[GetAccountByAccountName] accountName:%v, err:%v", accountName, err)
		l.Error(err)
		return nil, err
	}
	assets, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(uint32(account.AccountIndex))
	if err != nil {
		err = fmt.Errorf("[GetLatestAccountInfoByAccountIndex] err:%v", err)
		l.Error(err)
		return nil, err
	}

	for _, asset := range assets {
		resp.Account.Assets = append(resp.Account.Assets, &types.AssetInfo{
			AssetId: asset.AssetId,
			Balance: asset.Balance,
		})
	}

	resp.Account.AccountIndex = uint32(account.AccountIndex)
	resp.Account.AccountName = accountName
	resp.Account.AccountPk = account.PublicKey
	return resp, nil
}
