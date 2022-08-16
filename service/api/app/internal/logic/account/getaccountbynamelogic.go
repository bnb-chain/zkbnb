package account

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountByNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountByNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountByNameLogic {
	return &GetAccountByNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountByNameLogic) GetAccountByName(req *types.ReqGetAccountByName) (resp *types.RespGetAccountByName, err error) {
	if !utils.ValidateAccountName(req.AccountName) {
		logx.Errorf("invalid AccountName: %s", req.AccountName)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}

	accountIndex, err := l.svcCtx.MemCache.GetAccountIndexByName(req.AccountName)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	account, err := l.svcCtx.MemCache.GetLatestAccountWithFallback(accountIndex, func() (interface{}, error) {
		return l.svcCtx.StateFetcher.GetLatestAccount(l.ctx, accountIndex)
	})
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp = &types.RespGetAccountByName{
		AccountIndex: uint32(account.AccountIndex),
		AccountPk:    account.PublicKey,
		Nonce:        account.Nonce,
		Assets:       make([]*types.AccountAsset, 0),
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
