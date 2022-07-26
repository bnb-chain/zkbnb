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
	price, err := l.price.GetCurrencyPrice(l.ctx, req.Symbol)
	if err != nil {
		logx.Errorf("[GetCurrencyPrice] err:%v", err)
		return nil, err
	}
	l2Asset, err := l.l2asset.GetL2AssetInfoBySymbol(l.ctx, req.Symbol)
	if err != nil {
		logx.Errorf("[GetL2AssetInfoBySymbol] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetCurrencyPriceBySymbol{
		Price:   strconv.FormatFloat(price, 'E', -1, 64),
		AssetId: uint32(l2Asset.ID),
	}
	return resp, nil
}
