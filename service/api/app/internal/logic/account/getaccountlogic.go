package account

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountLogic {
	return &GetAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountLogic) GetAccount(req *types.ReqGetAccount) (resp *types.Account, err error) {
	accountIndex := int64(0)
	switch req.By {
	case "index":
		accountIndex, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid value for account index")
		}
	case "name":
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByName(req.Value)
	case "pk":
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByPk(req.Value)
	default:
		return nil, errorcode.AppErrInvalidParam.RefineError("param by should be index|name|pk")
	}

	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	account, err := l.svcCtx.StateFetcher.GetLatestAccount(l.ctx, accountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp = &types.Account{
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
