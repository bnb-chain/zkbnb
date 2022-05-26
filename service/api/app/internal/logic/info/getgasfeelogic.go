package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGasFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGasFeeLogic {
	return &GetGasFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGasFeeLogic) GetGasFee(req *types.ReqGetGasFee) (resp *types.RespGetGasFee, err error) {
	// todo: add your logic here and delete this line

	return
}
