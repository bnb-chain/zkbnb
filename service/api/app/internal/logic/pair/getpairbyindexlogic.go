package pair

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPairByIndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPairByIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPairByIndexLogic {
	return &GetPairByIndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPairByIndexLogic) GetPairByIndex(req *types.ReqGetPairByIndex) (resp *types.RespGetPairByIndex, err error) {
	if !utils.ValidatePairIndex(req.PairIndex) {
		logx.Errorf("invalid PairIndex: %d", req.PairIndex)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid PairIndex")
	}

	pair, err := l.svcCtx.StateFetcher.GetLatestLiquidityInfo(l.ctx, int64(req.PairIndex))
	if err != nil {
		logx.Errorf("fail to get pair info: %d, err: %s", req.PairIndex, err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp = &types.RespGetPairByIndex{
		AssetAId:      uint32(pair.AssetAId),
		AssetAAmount:  pair.AssetA.String(),
		AssetBId:      uint32(pair.AssetBId),
		AssetBAmount:  pair.AssetB.String(),
		TotalLpAmount: pair.LpAmount.String(),
	}
	return resp, nil
}
