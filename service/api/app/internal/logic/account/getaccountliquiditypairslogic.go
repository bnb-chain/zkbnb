package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/liquidity"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountLiquidityPairsLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
	liquidity liquidity.Liquidity
	mempool   mempool.Mempool
}

func NewGetAccountLiquidityPairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountLiquidityPairsLogic {
	return &GetAccountLiquidityPairsLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx.Config, ctx),
		liquidity: liquidity.New(svcCtx.Config),
		mempool:   mempool.New(svcCtx.Config),
	}
}

func (l *GetAccountLiquidityPairsLogic) GetLiquidityPairs(req *types.ReqGetAccountLiquidityPairs) (resp *types.RespGetAccountLiquidityPairs, err error) {
	if utils.CheckAccountIndex(req.AccountIndex) {
		logx.Error("[CheckAccountIndex] param:%v", req.AccountIndex)
		return nil, errcode.ErrInvalidParam
	}
	// AccountIndex or pairIndex？
	// entities, err := l.liquidity.GetLiquidityByPairIndex(int64(req.AccountIndex))
	// if err != nil {
	// 	logx.Error("[GetLiquidityByPairIndex] err:%v", err)
	// 	return nil, err
	// }
	// get created_at
	// mempoolDetails, err := l.mempool.GetLatestMempoolDetailUnscopedGroupByAssetIdAndChainId(
	// 	int64(req.AccountIndex), commonAsset.LiquidityLpAssetType)
	// for _, entity := range entities {
	// 	resRpc, err := l.globalRPC.GetLatestAccountLp(l.ctx, &globalrpc.ReqGetLatestAccountLp{
	// 		PairIndex:    uint64(entity.PairIndex),
	// 		AccountIndex: uint64(entity.AccountIndex),
	// 	})
	// 	result := resRpc.Result
	// 	temp := &types.ResultGetAccountLiquidityPairs{
	// 		PairIndex:   uint16(result.PairIndex),
	// 		AssetAId:    uint16(result.AssetAId),
	// 		AssetAName:  result.AssetAName,
	// 		AssetBId:    uint16(result.AssetBId),
	// 		AssetBName:  result.AssetBName,
	// 		LpAmountEnc: result.LpEnc,
	// 		CreatedAt:   entity.Model.CreatedAt.UnixMilli(),
	// 	}
	// 	for _, mempoolDetail := range mempoolDetails {
	// 		if mempoolDetail.AssetId == entity.PairIndex {
	// 			temp.CreatedAt = mempoolDetail.Max.UnixMilli()
	// 			break
	// 		}
	// 	}
	// }
	return resp, nil
}
