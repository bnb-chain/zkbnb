package account

import (
	"context"

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
	accountName := utils.FormatSting(req.AccountName)
	account, err := l.account.GetAccountByAccountName(l.ctx, accountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}
	assets, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(uint32(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetAccountInfoByAccountName{
		AccountIndex: uint32(account.AccountIndex),
		AccountPk:    account.PublicKey,
		Assets:       make([]*types.AssetInfo, 0),
	}
	for _, asset := range assets {
		resp.Assets = append(resp.Assets, &types.AssetInfo{
			AssetId: asset.AssetId,
			Balance: asset.Balance,
		})
	}
	return resp, nil
}
