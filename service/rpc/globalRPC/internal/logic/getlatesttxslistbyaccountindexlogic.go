package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestTxsListByAccountIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestTxsListByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestTxsListByAccountIndexLogic {
	return &GetLatestTxsListByAccountIndexLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

//  Transaction
func (l *GetLatestTxsListByAccountIndexLogic) GetLatestTxsListByAccountIndex(in *globalRPCProto.ReqGetLatestTxsListByAccountIndex) (*globalRPCProto.RespGetLatestTxsListByAccountIndex, error) {
	// todo: add your logic here and delete this line

	return &globalRPCProto.RespGetLatestTxsListByAccountIndex{}, nil
}
