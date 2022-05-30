package pair

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/accountliquidity"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAvailablePairsLogic struct {
	logx.Logger
	ctx              context.Context
	svcCtx           *svc.ServiceContext
	accountliquidity accountliquidity.AccountLiquidity
}

func NewGetAvailablePairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvailablePairsLogic {
	return &GetAvailablePairsLogic{
		Logger:           logx.WithContext(ctx),
		ctx:              ctx,
		svcCtx:           svcCtx,
		accountliquidity: accountliquidity.New(svcCtx.Config),
	}
}

func (l *GetAvailablePairsLogic) GetAvailablePairs(req *types.ReqGetAvailablePairs) (resp *types.RespGetAvailablePairs, err error) {
	// todo: add your logic here and delete this line
	liquidityAssets, err := l.accountliquidity.GetAllLiquidityAssets()
	if err != nil {
		logx.Error("[GetAllLiquidityAssets] err:%v", err)
		return nil, err
	}
	for _, asset := range liquidityAssets {
		resp.Pairs = append(resp.Pairs, &types.Pair{
			PairIndex:  uint16(asset.PairIndex),
			AssetAId:   uint16(asset.AssetA),
			AssetAName: asset.AssetAR,
			AssetBId:   uint16(asset.AssetB),
			AssetBName: asset.AssetBR,
			// FeeRate:      asset.LpEnc,
			// TreasuryRate: asset.TreasuryRate,
		})
	}
	return resp, nil
}
