package pair

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbas/types"
)

type GetPairLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPairLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPairLogic {
	return &GetPairLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPairLogic) GetPair(req *types.ReqGetPair) (resp *types.Pair, err error) {
	pair, err := l.svcCtx.StateFetcher.GetLatestLiquidity(int64(req.Index))
	if err != nil {
		logx.Errorf("fail to get pair info: %d, err: %s", req.Index, err.Error())
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNotFound
		}
		return nil, types2.AppErrInternal
	}
	resp = &types.Pair{
		AssetAId:      uint32(pair.AssetAId),
		AssetAAmount:  pair.AssetA.String(),
		AssetBId:      uint32(pair.AssetBId),
		AssetBAmount:  pair.AssetB.String(),
		TotalLpAmount: pair.LpAmount.String(),
	}
	return resp, nil
}
