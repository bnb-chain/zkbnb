package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountLiquidityPairsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountLiquidityPairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountLiquidityPairsLogic {
	return &GetAccountLiquidityPairsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountLiquidityPairsLogic) GetAccountLiquidityPairs(req *types.ReqGetAccountLiquidityPairs) (resp *types.RespGetAccountLiquidityPairs, err error) {
	// todo: add your logic here and delete this line

	return
}
