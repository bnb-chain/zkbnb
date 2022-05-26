package pair

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSwapAmountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSwapAmountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSwapAmountLogic {
	return &GetSwapAmountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSwapAmountLogic) GetSwapAmount(req *types.ReqGetSwapAmount) (resp *types.RespGetSwapAmount, err error) {
	// todo: add your logic here and delete this line

	return
}
