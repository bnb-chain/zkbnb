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

type GetLPValueLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetLPValueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLPValueLogic {
	return &GetLPValueLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetLPValueLogic) GetLPValue(req *types.ReqGetLPValue) (resp *types.RespGetLPValue, err error) {
	if checker.CheckPairIndex(req.PairIndex) {
		logx.Error("[CheckPairIndex] param:%v", req.PairIndex)
		return nil, errorcode.AppErrInvalidParam
	}
	lpValue, err := l.globalRPC.GetLpValue(l.ctx, req.PairIndex, req.LpAmount)
	if err != nil {
		logx.Error("[GetLpValue] err:%v", err)
		return nil, err
	}
	resp = &types.RespGetLPValue{
		AssetAId:     lpValue.AssetAId,
		AssetAAmount: lpValue.AssetAAmount,
		AssetBid:     lpValue.AssetBId,
		AssetBAmount: lpValue.AssetBAmount,
	}
	return resp, nil
}
