package info

import (
	"context"
	"strconv"

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
		price:   price.New(svcCtx),
		l2asset: l2asset.New(svcCtx),
	}
}

func (l *GetCurrencyPriceBySymbolLogic) GetCurrencyPriceBySymbol(req *types.ReqGetCurrencyPriceBySymbol) (*types.RespGetCurrencyPriceBySymbol, error) {
	_price, err := l.price.GetCurrencyPrice(l.ctx, req.Symbol)
	if err != nil {
		logx.Errorf("[GetCurrencyPrice] err:%v", err)
		return nil, err
	}
	_price2 := strconv.FormatFloat(_price, 'f', 30, 32) //float64 to string
	resp := &types.RespGetCurrencyPriceBySymbol{Price: _price2}
	l2Asset, err := l.l2asset.GetL2AssetInfoBySymbol(req.Symbol)
	if err != nil {
		logx.Errorf("[GetL2AssetInfoBySymbol] err:%v", err)
		return nil, err
	}
	resp.AssetId = uint32(l2Asset.ID)
	logx.Info("[GetL2AssetInfoBySymbol]", "Symbol:", req.Symbol, "response:", resp)
	return resp, nil
}
