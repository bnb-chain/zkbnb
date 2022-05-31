package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestTxsListByAccountIndexAndTxTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestTxsListByAccountIndexAndTxTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestTxsListByAccountIndexAndTxTypeLogic {
	return &GetLatestTxsListByAccountIndexAndTxTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLatestTxsListByAccountIndexAndTxTypeLogic) GetLatestTxsListByAccountIndexAndTxType(in *globalRPCProto.ReqGetLatestTxsListByAccountIndexAndTxType) (*globalRPCProto.RespGetLatestTxsListByAccountIndexAndTxType, error) {
	// todo: add your logic here and delete this line

	return &globalRPCProto.RespGetLatestTxsListByAccountIndexAndTxType{}, nil
}
