package info

import (
	"context"
	"strconv"

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
		logx.Errorf("[GetL2AssetsList] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetCurrencyPrices{}
	for _, asset := range l2Assets {
		_price, err := l.price.GetCurrencyPrice(l.ctx, asset.AssetSymbol)
		_price2 := strconv.FormatFloat(_price, 'f', 30, 32) //float64 to string
		if err != nil {
			logx.Errorf("[GetCurrencyPrice] err:%v", err)
			return nil, err
		}
		resp.Data = append(resp.Data, &types.DataCurrencyPrices{
			Pair:    asset.AssetSymbol + "/" + "USDT",
			AssetId: asset.AssetId,
			Price:   _price2,
		})
	}
	return resp, nil
}
