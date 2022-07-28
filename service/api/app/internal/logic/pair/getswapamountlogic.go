package pair

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/checker"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
	if checker.CheckPairIndex(req.PairIndex) {
		logx.Error("[CheckPairIndex] param:%v", req.PairIndex)
		return nil, errorcode.AppErrInvalidParam
	}
	if checker.CheckAssetId(req.AssetId) {
		logx.Error("[CheckAssetId] param:%v", req.AssetId)
		return nil, errorcode.AppErrInvalidParam
	}
	resAssetAmount, resAssetId, err := l.globalRPC.GetSwapAmount(l.ctx, uint64(req.PairIndex), uint64(req.AssetId), req.AssetAmount, req.IsFrom)
	if err != nil {
		logx.Error("[GetSwapAmount] err:%v", err)
		return nil, err
	}
	return &types.RespGetSwapAmount{
		ResAssetAmount: resAssetAmount,
		ResAssetId:     resAssetId,
	}, nil
}
