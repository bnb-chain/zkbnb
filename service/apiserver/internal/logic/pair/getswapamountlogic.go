package pair

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetSwapAmountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSwapAmountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSwapAmountLogic {
	return &GetSwapAmountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSwapAmountLogic) GetSwapAmount(req *types.ReqGetSwapAmount) (*types.SwapAmount, error) {
	deltaAmount, isTure := new(big.Int).SetString(req.AssetAmount, 10)
	if !isTure {
		logx.Errorf("fail to convert string: %s to int", req.AssetAmount)
		return nil, types2.AppErrInvalidParam.RefineError("invalid AssetAmount")
	}

	liquidity, err := l.svcCtx.StateFetcher.GetLatestLiquidity(int64(req.PairIndex))
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNotFound
		}
		return nil, types2.AppErrInternal
	}

	if liquidity.AssetA == nil || liquidity.AssetB == nil {
		logx.Errorf("invalid liquidity: %v", liquidity)
		return nil, types2.AppErrInternal
	}

	if int64(req.AssetId) != liquidity.AssetAId && int64(req.AssetId) != liquidity.AssetBId {
		logx.Errorf("invalid liquidity asset ids: %v", liquidity)
		return nil, types2.AppErrInvalidParam.RefineError("invalid AssetId")
	}

	if liquidity.AssetA.Cmp(big.NewInt(0)) == 0 || liquidity.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("invalid liquidity asset amount: %v", liquidity)
		return nil, types2.AppErrInvalidParam.RefineError("invalid PairIndex, empty liquidity or invalid pair")
	}

	var assetAmount *big.Int
	var toAssetId int64
	assetAmount, toAssetId, err = chain.ComputeDelta(liquidity.AssetA, liquidity.AssetB, liquidity.AssetAId, liquidity.AssetBId,
		int64(req.AssetId), req.IsFrom, deltaAmount, liquidity.FeeRate)
	if err != nil {
		logx.Errorf("fail to compute delta, err: %s", err.Error())
		return nil, types2.AppErrInternal
	}
	assetName, _ := l.svcCtx.MemCache.GetAssetNameById(toAssetId)
	return &types.SwapAmount{
		AssetId:     uint32(toAssetId),
		AssetName:   assetName,
		AssetAmount: assetAmount.String(),
	}, nil
}
