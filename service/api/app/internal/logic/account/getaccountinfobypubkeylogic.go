package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByPubKeyLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	account   account.Model
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
	account, err := l.account.GetAccountByPk(req.AccountPk)
	if err != nil {
		logx.Errorf("[GetAccountByPk] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetAccountInfoByPubKey{
		AccountStatus: uint32(account.Status),
		AccountName:   account.AccountName,
		AccountIndex:  account.AccountIndex,
		Assets:        make([]*types.AccountAsset, 0),
	}
	assets, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(uint32(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	for _, asset := range assets {
		resp.Assets = append(resp.Assets, &types.AccountAsset{
			AssetId:                  asset.AssetId,
			Balance:                  asset.Balance,
			LpAmount:                 asset.LpAmount,
			OfferCanceledOrFinalized: asset.OfferCanceledOrFinalized,
		})
	}
	return resp, nil
}
