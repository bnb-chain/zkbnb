package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetL1AmountByAssetidLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetL1AmountByAssetidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetL1AmountByAssetidLogic {
	return &GetL1AmountByAssetidLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx.Config, ctx),
	}
}

func (l *GetL1AmountByAssetidLogic) GetL1AmountByAssetid(req *types.ReqGetL1AmountByAssetid) (resp *types.RespGetL1AmountByAssetid, err error) {
	// TODO: globalRPC.GetLatestL1Amount
	resp.TotalAmount, err = l.globalRPC.GetLatestL1Amount(req.AssetId)
	if err != nil {
		logx.Error("[GetLatestL1Amount] err:%v", err)
		return nil, err
	}
	return resp, nil
}
