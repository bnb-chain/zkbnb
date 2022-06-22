package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByPubKeyLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	account   account.AccountModel
	globalRPC globalrpc.GlobalRPC
}

func NewGetAccountInfoByPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByPubKeyLogic {
	return &GetAccountInfoByPubKeyLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		account:   account.New(svcCtx),
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetAccountInfoByPubKeyLogic) GetAccountInfoByPubKey(req *types.ReqGetAccountInfoByPubKey) (*types.RespGetAccountInfoByPubKey, error) {
	account, err := l.account.GetAccountByAccountPk(l.ctx, req.AccountPk)
	if err != nil {
		logx.Errorf("[GetAccountByAccountPk] err:%v", err)
		return nil, err
	}
	assets, err := l.globalRPC.GetLatestAssetsListByAccountIndex(uint32(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAssetsListByAccountIndex] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetAccountInfoByPubKey{
		AccountIndex: uint32(account.AccountIndex),
		AccountName:  account.PublicKey,
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
