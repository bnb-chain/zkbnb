package pair

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
	if !utils.ValidatePairIndex(req.Index) {
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid PairIndex")
	}

	pair, err := l.svcCtx.StateFetcher.GetLatestLiquidity(int64(req.Index))
	if err != nil {
		logx.Errorf("fail to get pair info: %d, err: %s", req.Index, err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
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
