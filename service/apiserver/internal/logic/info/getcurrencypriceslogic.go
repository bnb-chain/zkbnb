package info

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
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

func (l *GetCurrencyPricesLogic) GetCurrencyPrices(req *types.ReqGetRange) (resp *types.CurrencyPrices, err error) {
	total, err := l.svcCtx.MemCache.GetAssetTotalCountWithFallback(func() (interface{}, error) {
		return l.svcCtx.AssetModel.GetAssetsTotalCount()
	})
	if err != nil {
		return nil, types2.AppErrInternal
	}

	resp = &types.CurrencyPrices{
		CurrencyPrices: make([]*types.CurrencyPrice, 0),
		Total:          uint32(total),
	}
	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	assets, err := l.svcCtx.AssetModel.GetAssetsList(int64(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, types2.AppErrInternal
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
				return nil, types2.AppErrInternal
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
