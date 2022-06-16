package pair

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSwapAmountLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetSwapAmountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSwapAmountLogic {
	return &GetSwapAmountLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetSwapAmountLogic) GetSwapAmount(req *types.ReqGetSwapAmount) (*types.RespGetSwapAmount, error) {
	if utils.CheckPairIndex(req.PairIndex) {
		logx.Error("[CheckPairIndex] param:%v", req.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckAssetId(req.AssetId) {
		logx.Error("[CheckAssetId] param:%v", req.AssetId)
		return nil, errcode.ErrInvalidParam
	}
	resAssetAmount, resAssetId, err := l.globalRPC.GetSwapAmount(l.ctx, uint64(req.PairIndex), uint64(req.AssetId), req.AssetAmount, req.IsFrom)
	if err != nil {
		logx.Error("[GetSwapAmount] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetSwapAmount{resAssetAmount, resAssetId}

	return resp, nil
}
