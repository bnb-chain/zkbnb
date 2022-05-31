package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestAssetsListByAccountIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestAssetsListByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAssetsListByAccountIndexLogic {
	return &GetLatestAssetsListByAccountIndexLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLatestAssetsListByAccountIndexLogic) GetLatestAssetsListByAccountIndex(in *globalRPCProto.ReqGetLatestAssetsListByAccountIndex) (*globalRPCProto.RespGetLatestAssetsListByAccountIndex, error) {
	// todo: add your logic here and delete this line

	return &globalRPCProto.RespGetLatestAssetsListByAccountIndex{}, nil
}
