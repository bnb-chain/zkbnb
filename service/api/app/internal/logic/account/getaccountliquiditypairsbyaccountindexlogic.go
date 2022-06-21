package account

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/errcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/liquidity"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
	"github.com/bnb-chain/zkbas/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountLiquidityPairsByAccountIndexLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	liquidity liquidity.Liquidity
}

func NewGetAccountLiquidityPairsByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountLiquidityPairsByAccountIndexLogic {
	return &GetAccountLiquidityPairsByAccountIndexLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		liquidity: liquidity.New(svcCtx),
	}
}

func (l *GetAccountLiquidityPairsByAccountIndexLogic) GetAccountLiquidityPairsByAccountIndex(req *types.ReqGetAccountLiquidityPairsByAccountIndex) (resp *types.RespGetAccountLiquidityPairsByAccountIndex, err error) {
	if utils.CheckAccountIndex(req.AccountIndex) {
		logx.Errorf("[CheckAccountIndex] param:%v", req.AccountIndex)
		return nil, errcode.ErrInvalidParam
	}
	entitie, err := l.liquidity.GetLiquidityByPairIndex(int64(req.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLiquidityByPairIndex] err:%v", err.Error())
		return nil, err
	}
	if entitie == nil {
		logx.Errorf("[GetLiquidityByPairIndex] err:%v", err.Error())
		return nil, err
	}
	pair := &types.AccountLiquidityPairs{
		PairIndex:   uint32(entitie.PairIndex),
		AssetAId:    uint32(entitie.AssetAId),
		AssetAName:  entitie.AssetA,
		AssetBId:    uint32(entitie.AssetBId),
		AssetBName:  entitie.AssetB,
		LpAmountEnc: entitie.LpAmount,
		CreatedAt:   entitie.CreatedAt.Unix(),
	}
	resp = &types.RespGetAccountLiquidityPairsByAccountIndex{}
	resp.Pairs = append(resp.Pairs, pair)
	return resp, nil
}
