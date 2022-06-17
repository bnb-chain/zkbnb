package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTransactionCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTransactionCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTransactionCountLogic {
	return &GetTransactionCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTransactionCountLogic) GetTransactionCount(in *globalRPCProto.ReqGetTransactionCount) (*globalRPCProto.RespGetTransactionCount, error) {
	// todo: add your logic here and delete this line

	return &globalRPCProto.RespGetTransactionCount{}, nil
}
