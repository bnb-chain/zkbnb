package pair

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalrpc"
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
		globalRPC: globalrpc.New(svcCtx.Config),
	}
}

func (l *GetSwapAmountLogic) GetSwapAmount(req *types.ReqGetSwapAmount) (resp *types.RespGetSwapAmount, err error) {
	if utils.CheckPairIndex(req.PairIndex) {
		logx.Error("[CheckPairIndex] param:%v", req.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckLPAmount(req.LpAmount) {
		logx.Error("[CheckLPAmount] param:%v", req.LpAmount)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckAssetId(req.AssetId) {
		logx.Error("[CheckAssetId] param:%v", req.AssetId)
		return nil, errcode.ErrInvalidParam
	}
	resp.PairIndex, resp.ResAssetAmount, resp.ResAssetId, err = l.globalRPC.GetSwapAmount(req.PairIndex, req.AssetId, req.AssetAmount, req.IsFrom)
	if err != nil {
		logx.Error("[GetSwapAmount] err:%v", err)
		return nil, err
	}
	return resp, nil
}
