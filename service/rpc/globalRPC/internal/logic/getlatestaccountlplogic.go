package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestAccountLpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestAccountLpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAccountLpLogic {
	return &GetLatestAccountLpLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

//  Asset
func (l *GetLatestAccountLpLogic) GetLatestAccountLp(in *globalRPCProto.ReqGetLatestAccountLp) (*globalRPCProto.RespGetLatestAccountLp, error) {
	// todo: add your logic here and delete this line

	return &globalRPCProto.RespGetLatestAccountLp{}, nil
}
