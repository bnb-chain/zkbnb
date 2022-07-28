package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountInfoByAccountIndexLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
	account   account.Model
}

func NewGetAccountInfoByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAccountIndexLogic {
	return &GetAccountInfoByAccountIndexLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
		account:   account.New(svcCtx),
	}
}

func (l *GetAccountInfoByAccountIndexLogic) GetAccountInfoByAccountIndex(req *types.ReqGetAccountInfoByAccountIndex) (*types.RespGetAccountInfoByAccountIndex, error) {
	account, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(l.ctx, (req.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetAccountInfoByAccountIndex{
		AccountStatus: uint32(account.Status),
		AccountName:   account.AccountName,
		AccountPk:     account.PublicKey,
		Nonce:         account.Nonce,
		Assets:        make([]*types.AccountAsset, 0),
	}
	for _, asset := range account.AccountAsset {
		resp.Assets = append(resp.Assets, &types.AccountAsset{
			AssetId:                  asset.AssetId,
			Balance:                  asset.Balance,
			LpAmount:                 asset.LpAmount,
			OfferCanceledOrFinalized: asset.OfferCanceledOrFinalized,
		})
	}
	return resp, nil
}
