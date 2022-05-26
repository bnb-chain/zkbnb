package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetL1AmountByAssetidLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetL1AmountByAssetidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetL1AmountByAssetidLogic {
	return &GetL1AmountByAssetidLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetL1AmountByAssetidLogic) GetL1AmountByAssetid(req *types.ReqGetL1AmountByAssetid) (resp *types.RespGetL1AmountByAssetid, err error) {
	// todo: add your logic here and delete this line

	return
}
