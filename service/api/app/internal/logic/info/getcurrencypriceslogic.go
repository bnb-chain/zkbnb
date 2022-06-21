package info

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/l2asset"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/price"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrencyPricesLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	price   price.Price
	l2asset l2asset.L2asset
}

func NewGetCurrencyPricesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrencyPricesLogic {
	return &GetCurrencyPricesLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		price:   price.New(svcCtx),
		l2asset: l2asset.New(svcCtx),
	}
}

func (l *GetCurrencyPricesLogic) GetCurrencyPrices(_ *types.ReqGetCurrencyPrices) (*types.RespGetCurrencyPrices, error) {
	l2Assets, err := l.l2asset.GetL2AssetsList()
	if err != nil {
		logx.Error("[GetL2AssetsList] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetCurrencyPrices{}
	for _, asset := range l2Assets {
		currencyPrice, err := l.price.GetCurrencyPrice(l.ctx, asset.AssetSymbol)
		if err != nil {
			logx.Error("[GetCurrencyPrice] err:%v", err)
			return nil, err
		}
		resp.Data = append(resp.Data, &types.DataCurrencyPrices{
			Pair:    asset.AssetSymbol + "/" + "USDT",
			AssetId: asset.AssetId,
			Price:   currencyPrice,
		})
	}
	return resp, nil
}
