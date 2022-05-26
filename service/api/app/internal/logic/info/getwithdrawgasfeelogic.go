package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWithdrawGasFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetWithdrawGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawGasFeeLogic {
	return &GetWithdrawGasFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetWithdrawGasFeeLogic) GetWithdrawGasFee(req *types.ReqGetWithdrawGasFee) (resp *types.RespGetWithdrawGasFee, err error) {
	// todo: add your logic here and delete this line

	return
}
