package pair

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/errcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
	"github.com/bnb-chain/zkbas/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPairInfoLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetPairInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPairInfoLogic {
	return &GetPairInfoLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetPairInfoLogic) GetPairInfo(req *types.ReqGetPairInfo) (*types.RespGetPairInfo, error) {
	if utils.CheckPairIndex(req.PairIndex) {
		logx.Error("[CheckPairIndex] param:%v", req.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	pair, err := l.globalRPC.GetPairInfo(req.PairIndex)
	if err != nil {
		logx.Error("[GetPairRatio] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetPairInfo{
		AssetAId:      pair.AssetAId,
		AssetAAmount:  pair.AssetAAmount,
		AssetBId:      pair.AssetBId,
		AssetBAmount:  pair.AssetBAmount,
		TotalLpAmount: pair.LpAmount,
	}
	return resp, nil
}
