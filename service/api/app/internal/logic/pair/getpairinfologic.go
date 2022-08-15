package pair

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetPairInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPairInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPairInfoLogic {
	return &GetPairInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPairInfoLogic) GetPairInfo(req *types.ReqGetPairInfo) (*types.RespGetPairInfo, error) {
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
	resp := &types.RespGetPairInfo{
		AssetAId:      uint32(pair.AssetAId),
		AssetAAmount:  pair.AssetA.String(),
		AssetBId:      uint32(pair.AssetBId),
		AssetBAmount:  pair.AssetB.String(),
		TotalLpAmount: pair.LpAmount.String(),
	}
	return resp, nil
}
