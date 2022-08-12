package info

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetCurrencyPricesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrencyPricesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrencyPricesLogic {
	return &GetCurrencyPricesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrencyPricesLogic) GetCurrencyPrices() (*types.RespGetCurrencyPrices, error) {
	assets, err := l.svcCtx.MemCache.GetAssetsWithFallback(func() (interface{}, error) {
		return l.svcCtx.AssetModel.GetAssetsList()
	})
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp := &types.RespGetCurrencyPrices{
		CurrencyPrices: make([]*types.CurrencyPrice, 0),
	}
	for _, asset := range assets {
		price := 0.0
		if asset.AssetSymbol == "LEG" {
			price = 1.0
		} else if asset.AssetSymbol == "REY" {
			price = 0.5
		} else {
			price, err = l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, asset.AssetSymbol)
			if err != nil {
				logx.Errorf("fail to get price for symbol: %s, err: %s", asset.AssetSymbol, err.Error())
				if err == errorcode.AppErrQuoteNotExist {
					continue
				}
				return nil, errorcode.AppErrInternal
			}
		}

		resp.CurrencyPrices = append(resp.CurrencyPrices, &types.CurrencyPrice{
			Pair:    asset.AssetSymbol + "/" + "USDT",
			AssetId: asset.AssetId,
			Price:   strconv.FormatFloat(price, 'E', -1, 64),
		})
	}
	return resp, nil
}
