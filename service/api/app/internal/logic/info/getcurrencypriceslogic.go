package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GetCurrencyPricesLogic) GetCurrencyPrices(req *types.ReqGetCurrencyPrices) (resp *types.RespGetCurrencyPrices, err error) {
	// todo: add your logic here and delete this line

	return
}
