package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/price"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

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
		price:   price.New(svcCtx.Config),
		l2asset: l2asset.New(svcCtx.Config),
	}
}

func (l *GetCurrencyPricesLogic) GetCurrencyPrices(req *types.ReqGetCurrencyPrices) (resp *types.RespGetCurrencyPrices, err error) {
	l2Assets, err := l.l2asset.GetL2AssetsList()
	if err != nil {
		logx.Error("[GetL2AssetsList] err:%v", err)
		return nil, err
	}
	resp.Data = make([]*types.DataCurrencyPrices, 0)
	for _, asset := range l2Assets {
		price, err := l.price.GetCurrencyPrice(asset.L2Symbol)
		if err != nil {
			logx.Error("[GetCurrencyPrice] err:%v", err)
			return nil, err
		}
		resp.Data = append(resp.Data, &types.DataCurrencyPrices{
			Pair:    asset.L2Symbol + "/" + "USDT",
			AssetId: int(asset.L2AssetId),
			Price:   price,
		})
	}
	return resp, nil
}
