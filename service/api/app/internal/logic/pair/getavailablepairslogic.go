package pair

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAvailablePairsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAvailablePairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvailablePairsLogic {
	return &GetAvailablePairsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAvailablePairsLogic) GetAvailablePairs(_ *types.ReqGetAvailablePairs) (*types.RespGetAvailablePairs, error) {
	resp := &types.RespGetAvailablePairs{}
	liquidityAssets, err := l.svcCtx.LiquidityModel.GetAllLiquidityAssets()
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return resp, nil
		}
		return nil, errorcode.AppErrInternal
	}

	for _, asset := range liquidityAssets {
		assetA, err := l.svcCtx.AssetModel.GetAssetByAssetId(asset.AssetAId)
		if err != nil {
			return nil, errorcode.AppErrInternal
		}
		assetB, err := l.svcCtx.AssetModel.GetAssetByAssetId(asset.AssetBId)
		if err != nil {
			return nil, errorcode.AppErrInternal
		}
		resp.Pairs = append(resp.Pairs, &types.Pair{
			PairIndex:    uint32(asset.PairIndex),
			AssetAId:     uint32(asset.AssetAId),
			AssetAName:   assetA.AssetName,
			AssetAAmount: asset.AssetA,
			AssetBId:     uint32(asset.AssetBId),
			AssetBName:   assetB.AssetName,
			AssetBAmount: asset.AssetB,
			FeeRate:      asset.FeeRate,
			TreasuryRate: asset.TreasuryRate,
		})
	}
	return resp, nil
}
