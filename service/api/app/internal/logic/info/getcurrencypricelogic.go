package info

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

const (
	queryBySymbol = "symbol"
)

type GetCurrencyPriceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrencyPriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrencyPriceLogic {
	return &GetCurrencyPriceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrencyPriceLogic) GetCurrencyPrice(req *types.ReqGetCurrencyPrice) (resp *types.CurrencyPrice, err error) {
	symbol := ""
	switch req.By {
	case queryBySymbol:
		if !utils.ValidateSymbol(req.Value) {
			logx.Errorf("invalid Symbol: %s", req.Value)
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid symbol")
		}
		symbol = strings.ToUpper(req.Value)
	default:
		return nil, errorcode.AppErrInvalidParam.RefineError("param by should be symbol")
	}

	asset, err := l.svcCtx.MemCache.GetAssetBySymbolWithFallback(symbol, func() (interface{}, error) {
		return l.svcCtx.AssetModel.GetAssetBySymbol(symbol)
	})
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	price, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, symbol)
	if err != nil {
		return nil, errorcode.AppErrInternal
	}
	resp = &types.CurrencyPrice{
		Pair:    asset.AssetSymbol + "/" + "USDT",
		Price:   strconv.FormatFloat(price, 'E', -1, 64),
		AssetId: uint32(asset.ID),
	}
	return resp, nil
}
