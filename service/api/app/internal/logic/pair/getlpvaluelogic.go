package pair

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/common/checker"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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
		return nil, errcode.ErrInvalidParam
	}
	lpValue, err := l.globalRPC.GetLpValue(req.PairIndex, req.LpAmount)
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
