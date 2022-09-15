package pair

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
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
	liquidity, err := l.svcCtx.StateFetcher.GetLatestLiquidity(int64(req.Index))
	if err != nil {
		logx.Errorf("fail to get pair info: %d, err: %s", req.Index, err.Error())
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNotFound
		}
		return nil, types2.AppErrInternal
	}
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
	resp = &types.Pair{
		Index:         uint32(liquidity.PairIndex),
		AssetAId:      uint32(liquidity.AssetAId),
		AssetAName:    assetA.AssetName,
		AssetAAmount:  liquidity.AssetA.String(),
		AssetAPrice:   strconv.FormatFloat(assetAPrice, 'E', -1, 64),
		AssetBId:      uint32(liquidity.AssetBId),
		AssetBName:    assetB.AssetName,
		AssetBAmount:  liquidity.AssetB.String(),
		AssetBPrice:   strconv.FormatFloat(assetBPrice, 'E', -1, 64),
		TotalLpAmount: liquidity.LpAmount.String(),
	}
	return resp, nil
}
