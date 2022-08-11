package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountInfoByAccountIndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountInfoByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAccountIndexLogic {
	return &GetAccountInfoByAccountIndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountInfoByAccountIndexLogic) GetAccountInfoByAccountIndex(req *types.ReqGetAccountInfoByAccountIndex) (*types.RespGetAccountInfoByAccountIndex, error) {
	account, err := l.svcCtx.StateFetcher.GetLatestAccountInfo(l.ctx, req.AccountIndex)
	if err != nil {
		logx.Errorf("fail to get account info: %d, err: %s", req.AccountIndex, err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetAccountInfoByAccountIndex{
		AccountStatus: uint32(account.Status),
		AccountName:   account.AccountName,
		AccountPk:     account.PublicKey,
		Nonce:         account.Nonce,
		Assets:        make([]*types.AccountAsset, 0),
	}
	for _, asset := range account.AssetInfo {
		resp.Assets = append(resp.Assets, &types.AccountAsset{
			AssetId:                  uint32(asset.AssetId),
			Balance:                  asset.Balance.String(),
			LpAmount:                 asset.LpAmount.String(),
			OfferCanceledOrFinalized: asset.OfferCanceledOrFinalized.String(),
		})
	}
	return resp, nil
}
