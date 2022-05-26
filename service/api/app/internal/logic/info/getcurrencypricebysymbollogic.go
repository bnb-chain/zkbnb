package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/price"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrencyPriceBySymbolLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	price   price.Price
	l2asset l2asset.L2asset
}

func NewGetCurrencyPriceBySymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrencyPriceBySymbolLogic {
	return &GetCurrencyPriceBySymbolLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		price:   price.New(svcCtx.Config),
		l2asset: l2asset.New(svcCtx.Config),
	}
}

func (l *GetCurrencyPriceBySymbolLogic) GetCurrencyPriceBySymbol(req *types.ReqGetCurrencyPriceBySymbol) (resp *types.RespGetCurrencyPriceBySymbol, err error) {
	resp.Price, err = l.price.GetCurrencyPrice(req.Symbol)
	if err != nil {
		logx.Error("[GetAssets] err:%v", err)
		return nil, err
	}
	l2Asset, err := l.l2asset.GetL2AssetInfoBySymbol(req.Symbol)
	if err != nil {
		logx.Error("[GetL2AssetInfoBySymbol] err:%v", err)
		return nil, err
	}
	resp.AssetId = int(l2Asset.ID)
	return resp, nil
}
