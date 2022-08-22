package pair

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AssetAmount")
	}

	liquidity, err := l.svcCtx.StateFetcher.GetLatestLiquidity(int64(req.PairIndex))
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	if liquidity.AssetA == nil || liquidity.AssetB == nil {
		logx.Errorf("invalid liquidity: %v", liquidity)
		return nil, errorcode.AppErrInternal
	}

	if int64(req.AssetId) != liquidity.AssetAId && int64(req.AssetId) != liquidity.AssetBId {
		logx.Errorf("invalid liquidity asset ids: %v", liquidity)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AssetId")
	}

	if liquidity.AssetA.Cmp(big.NewInt(0)) == 0 || liquidity.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("invalid liquidity asset amount: %v", liquidity)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid PairIndex, empty liquidity or invalid pair")
	}

	var assetAmount *big.Int
	var toAssetId int64
	assetAmount, toAssetId, err = util.ComputeDelta(liquidity.AssetA, liquidity.AssetB, liquidity.AssetAId, liquidity.AssetBId,
		int64(req.AssetId), req.IsFrom, deltaAmount, liquidity.FeeRate)
	if err != nil {
		logx.Errorf("fail to compute delta, err: %s", err.Error())
		return nil, errorcode.AppErrInternal
	}
	assetName, _ := l.svcCtx.MemCache.GetAssetNameById(toAssetId)
	return &types.SwapAmount{
		AssetId:     uint32(toAssetId),
		AssetName:   assetName,
		AssetAmount: assetAmount.String(),
	}, nil
}
