package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLpValueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLpValueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLpValueLogic {
	return &GetLpValueLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLpValueLogic) GetLpValue(in *globalRPCProto.ReqGetLpValue) (*globalRPCProto.RespGetLpValue, error) {
	// todo: add your logic here and delete this line

	return &globalRPCProto.RespGetLpValue{}, nil
}
