package pair

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPairInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPairInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPairInfoLogic {
	return &GetPairInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPairInfoLogic) GetPairInfo(req *types.ReqGetPairInfo) (resp *types.RespGetPairInfo, err error) {
	// todo: add your logic here and delete this line

	return
}
