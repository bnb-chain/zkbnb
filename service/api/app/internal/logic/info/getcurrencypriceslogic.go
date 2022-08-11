package info

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/price"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetCurrencyPricesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	price  price.Price
}

func NewGetCurrencyPricesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrencyPricesLogic {
	return &GetCurrencyPricesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		price:  price.New(svcCtx),
	}
}

func (l *GetCurrencyPricesLogic) GetCurrencyPrices(req *types.ReqGetCurrencyPrices) (*types.RespGetCurrencyPrices, error) {
	l2Assets, err := l.svcCtx.L2AssetModel.GetAssetsList()
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	//TODO: performance issue here
	//TODO: why use cmc here?
	resp := &types.RespGetCurrencyPrices{}
	for _, asset := range l2Assets {
		price := 0.0
		if asset.AssetSymbol == "LEG" {
			price = 1.0
		} else if asset.AssetSymbol == "REY" {
			price = 0.5
		} else {
			price, err = l.price.GetCurrencyPrice(l.ctx, asset.AssetSymbol)
			if err != nil {
				logx.Errorf("fail to get price for symbol: %s, err: %s", asset.AssetSymbol, err.Error())
				if err == errorcode.AppErrQuoteNotExist {
					continue
				}
				return nil, errorcode.AppErrInternal
			}
		}

		resp.Data = append(resp.Data, &types.DataCurrencyPrices{
			Pair:    asset.AssetSymbol + "/" + "USDT",
			AssetId: asset.AssetId,
			Price:   strconv.FormatFloat(price, 'E', -1, 64),
		})
	}
	return resp, nil
}
