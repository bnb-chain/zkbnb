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

type GetAccountLiquidityPairsByAccountIndexLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
	liquidity liquidity.Liquidity
	mempool   mempool.Mempool
}

func NewGetAccountLiquidityPairsByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountLiquidityPairsByAccountIndexLogic {
	return &GetAccountLiquidityPairsByAccountIndexLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx.Config, ctx),
		liquidity: liquidity.New(svcCtx.Config),
		mempool:   mempool.New(svcCtx.Config),
	}
}

// TODO: check this func's logic
func (l *GetAccountLiquidityPairsByAccountIndexLogic) GetAccountLiquidityPairsByAccountIndex(req *types.ReqGetAccountLiquidityPairsByAccountIndex) (resp *types.RespGetAccountLiquidityPairsByAccountIndex, err error) {
	if utils.CheckAccountIndex(req.AccountIndex) {
		logx.Error("[CheckAccountIndex] param:%v", req.AccountIndex)
		return nil, errcode.ErrInvalidParam
	}
	entitie, err := l.liquidity.GetLiquidityByPairIndex(int64(req.AccountIndex))
	if err != nil {
		logx.Error("[GetLiquidityByPairIndex] err:%v", err)
		return nil, err
	}
	pair := &types.AccountLiquidityPairs{
		PairIndex:   uint32(entitie.PairIndex),
		AssetAId:    uint32(entitie.AssetAId),
		AssetAName:  entitie.AssetA,
		AssetBId:    uint32(entitie.AssetBId),
		AssetBName:  entitie.AssetB,
		LpAmountEnc: entitie.LpAmount,
		// CreatedAt  : entitie.
	}
	resp.Pairs = append(resp.Pairs, pair)
	return resp, nil
}
