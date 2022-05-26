package pair

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAvailablePairsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAvailablePairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvailablePairsLogic {
	return &GetAvailablePairsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAvailablePairsLogic) GetAvailablePairs(req *types.ReqGetAvailablePairs) (resp *types.RespGetAvailablePairs, err error) {
	// todo: add your logic here and delete this line

	return
}
