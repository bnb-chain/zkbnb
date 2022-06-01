package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestPairInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestPairInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestPairInfoLogic {
	return &GetLatestPairInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

//  Liquidity
func (l *GetLatestPairInfoLogic) GetLatestPairInfo(in *globalRPCProto.ReqGetLatestPairInfo) (*globalRPCProto.RespGetLatestPairInfo, error) {
	// todo: add your logic here and delete this line

	return &globalRPCProto.RespGetLatestPairInfo{}, nil
}
