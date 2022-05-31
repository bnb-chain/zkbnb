package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestAssetInfoByAccountIndexAndAssetIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestAssetInfoByAccountIndexAndAssetIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAssetInfoByAccountIndexAndAssetIdLogic {
	return &GetLatestAssetInfoByAccountIndexAndAssetIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLatestAssetInfoByAccountIndexAndAssetIdLogic) GetLatestAssetInfoByAccountIndexAndAssetId(in *globalRPCProto.ReqGetLatestAssetInfoByAccountIndexAndAssetId) (*globalRPCProto.RespGetLatestAssetInfoByAccountIndexAndAssetId, error) {
	// todo: add your logic here and delete this line

	return &globalRPCProto.RespGetLatestAssetInfoByAccountIndexAndAssetId{}, nil
}
