package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

	account, err := l.account.GetAccountByAccountIndex(req.AccountIndex)
	if err != nil {
		logx.Errorf("[GetAccountByAccountIndex] accountName:%v, err:%v", req.AccountIndex, err)
		return nil, err
	}
	resp := &types.RespGetAccountInfoByAccountIndex{
		AccountStatus: uint32(account.Status),
		AccountName:   account.AccountName,
		AccountPk:     account.PublicKey,
		Assets:        make([]*types.AccountAsset, 0),
	}
	assets, err := l.globalRPC.GetLatestAssetsListByAccountIndex(uint32(req.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAssetsListByAccountIndex] err:%v", err)
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
