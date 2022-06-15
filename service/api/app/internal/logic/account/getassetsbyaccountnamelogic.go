package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetsByAccountNameLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	account   account.AccountModel
	globalRPC globalrpc.GlobalRPC
}

func NewGetAssetsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetsByAccountNameLogic {
	return &GetAssetsByAccountNameLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		account:   account.New(svcCtx.Config),
		globalRPC: globalrpc.New(svcCtx.Config, ctx),
	}
}

func (l *GetAssetsByAccountNameLogic) GetAssetsByAccountName(req *types.ReqGetAssetsByAccountName) (*types.RespGetAssetsByAccountName, error) {
	if utils.CheckAccountName(req.AccountName) {
		logx.Errorf("[CheckAccountName] param:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	account, err := l.account.GetAccountByAccountName(req.AccountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}
	assets, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(uint32(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetAssetsByAccountName{
		Assets: make([]*types.Asset, 0),
	}
	for _, asset := range assets {
		v := &types.Asset{
			AssetId: asset.AssetId,
			Balance: asset.Balance,
		}
		resp.Assets = append(resp.Assets, v)
	}
	return resp, nil
}
