package pair

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
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
	if !utils.ValidatePairIndex(req.PairIndex) {
		logx.Errorf("invalid PairIndex: %d", req.PairIndex)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid PairIndex")
	}
	if !utils.ValidateAssetId(req.AssetId) {
		logx.Errorf("invalid AssetId: %d", req.AssetId)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AssetId")
	}
	resAssetAmount, resAssetId, err := l.globalRPC.GetSwapAmount(l.ctx, uint64(req.PairIndex), uint64(req.AssetId), req.AssetAmount, req.IsFrom)
	if err != nil {
		logx.Errorf("fail to get swap amount from rpc for pair: %d, asset: %d, err: %s", req.PairIndex, req.AssetId, err.Error())
		if err == errorcode.RpcErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	return &types.RespGetSwapAmount{
		ResAssetAmount: resAssetAmount,
		ResAssetId:     resAssetId,
	}, nil
}
