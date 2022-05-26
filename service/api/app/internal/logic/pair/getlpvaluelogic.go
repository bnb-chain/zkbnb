package pair

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLPValueLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLPValueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLPValueLogic {
	return &GetLPValueLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLPValueLogic) GetLPValue(req *types.ReqGetLPValue) (resp *types.RespGetLPValue, err error) {
	// todo: add your logic here and delete this line

	return
}
