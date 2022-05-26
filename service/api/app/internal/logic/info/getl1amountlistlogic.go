package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetL1AmountListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetL1AmountListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetL1AmountListLogic {
	return &GetL1AmountListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetL1AmountListLogic) GetL1AmountList(req *types.ReqGetL1AmountList) (resp *types.RespGetL1AmountList, err error) {
	// todo: add your logic here and delete this line

	return
}
