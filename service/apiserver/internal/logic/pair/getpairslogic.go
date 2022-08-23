package pair

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbas/types"
)

type GetPairsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPairsLogic {
	return &GetPairsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPairsLogic) GetPairs() (resp *types.Pairs, err error) {
	liquidityAssets, err := l.svcCtx.LiquidityModel.GetAllLiquidityAssets()
	if err != nil {
		if err == types2.DbErrNotFound {
			return resp, nil
		}
		return nil, types2.AppErrInternal
	}

	for _, asset := range liquidityAssets {
		assetA, err := l.svcCtx.AssetModel.GetAssetById(asset.AssetAId)
		if err != nil {
			return nil, types2.AppErrInternal
		}
		assetB, err := l.svcCtx.AssetModel.GetAssetById(asset.AssetBId)
		if err != nil {
			return nil, types2.AppErrInternal
		}
		resp.Pairs = append(resp.Pairs, &types.Pair{
			Index:        uint32(asset.PairIndex),
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
