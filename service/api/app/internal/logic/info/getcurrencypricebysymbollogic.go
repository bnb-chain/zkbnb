package info

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/price"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetCurrencyPriceBySymbolLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	price  price.Price
}

func NewGetCurrencyPriceBySymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrencyPriceBySymbolLogic {
	return &GetCurrencyPriceBySymbolLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		price:  price.New(svcCtx),
	}
}

func (l *GetCurrencyPriceBySymbolLogic) GetCurrencyPriceBySymbol(req *types.ReqGetCurrencyPriceBySymbol) (*types.RespGetCurrencyPriceBySymbol, error) {
	if utils.CheckSymbol(req.Symbol) {
		logx.Errorf("invalid Symbol: %s", req.Symbol)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid Symbol")
	}
	symbol := strings.ToUpper(req.Symbol)

	l2Asset, err := l.svcCtx.L2AssetModel.GetAssetInfoBySymbol(symbol)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	price, err := l.price.GetCurrencyPrice(l.ctx, symbol)
	if err != nil {
		if err == errorcode.AppErrQuoteNotExist {
			return nil, err
		}
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetCurrencyPriceBySymbol{
		Price:   strconv.FormatFloat(price, 'E', -1, 64),
		AssetId: uint32(l2Asset.ID),
	}
	return resp, nil
}
