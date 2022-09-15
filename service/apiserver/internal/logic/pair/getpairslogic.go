package pair

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
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

func (l *GetPairsLogic) GetPairs() (*types.Pairs, error) {
	resp := &types.Pairs{Pairs: make([]*types.Pair, 0)}

	liquidityAssets, err := l.svcCtx.LiquidityModel.GetAllLiquidity()
	if err != nil {
		if err == types2.DbErrNotFound {
			return resp, nil
		}
		return nil, types2.AppErrInternal
	}

	for _, liquidity := range liquidityAssets {
		assetA, err := l.svcCtx.AssetModel.GetAssetById(liquidity.AssetAId)
		if err != nil {
			return nil, types2.AppErrInternal
		}
		assetB, err := l.svcCtx.AssetModel.GetAssetById(liquidity.AssetBId)
		if err != nil {
			return nil, types2.AppErrInternal
		}
		assetAPrice, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, assetA.AssetSymbol)
		if err != nil {
			return nil, types2.AppErrInternal
		}
		assetBPrice, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, assetB.AssetSymbol)
		if err != nil {
			return nil, types2.AppErrInternal
		}
		resp.Pairs = append(resp.Pairs, &types.Pair{
			Index:         uint32(liquidity.PairIndex),
			AssetAId:      uint32(liquidity.AssetAId),
			AssetAName:    assetA.AssetName,
			AssetAAmount:  liquidity.AssetA,
			AssetAPrice:   strconv.FormatFloat(assetAPrice, 'E', -1, 64),
			AssetBId:      uint32(liquidity.AssetBId),
			AssetBName:    assetB.AssetName,
			AssetBAmount:  liquidity.AssetB,
			AssetBPrice:   strconv.FormatFloat(assetBPrice, 'E', -1, 64),
			FeeRate:       liquidity.FeeRate,
			TreasuryRate:  liquidity.TreasuryRate,
			TotalLpAmount: liquidity.LpAmount,
		})
	}
	return resp, nil
}
