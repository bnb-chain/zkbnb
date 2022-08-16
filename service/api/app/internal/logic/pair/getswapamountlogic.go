package pair

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
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
	if !utils.ValidatePairIndex(req.PairIndex) {
		logx.Errorf("invalid PairIndex: %d", req.PairIndex)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid PairIndex")
	}
	if !utils.ValidateAssetId(req.AssetId) {
		logx.Errorf("invalid AssetId: %d", req.AssetId)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AssetId")
	}

	deltaAmount, isTure := new(big.Int).SetString(req.AssetAmount, 10)
	if !isTure {
		logx.Errorf("fail to convert string: %s to int", req.AssetAmount)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AssetAmount")
	}

	liquidity, err := l.svcCtx.StateFetcher.GetLatestLiquidity(l.ctx, int64(req.PairIndex))
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	if liquidity.AssetA == nil || liquidity.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidity.AssetB == nil || liquidity.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("invalid liquidity asset amount: %v, err: %s", liquidity, errorcode.AppErrLiquidityInvalidAssetAmount.Error())
		return nil, errorcode.AppErrLiquidityInvalidAssetAmount
	}

	if int64(req.AssetId) != liquidity.AssetAId && int64(req.AssetId) != liquidity.AssetBId {
		logx.Errorf("invalid liquidity asset id: %v, err: %s", liquidity, errorcode.AppErrLiquidityInvalidAssetAmount.Error())
		return nil, errorcode.AppErrLiquidityInvalidAssetID
	}

	var assetAmount *big.Int
	var toAssetId int64
	assetAmount, toAssetId, err = util.ComputeDelta(liquidity.AssetA, liquidity.AssetB, liquidity.AssetAId, liquidity.AssetBId,
		int64(req.AssetId), req.IsFrom, deltaAmount, liquidity.FeeRate)
	if err != nil {
		logx.Errorf("fail to compute delta, err: %s", err.Error())
		return nil, errorcode.AppErrInternal
	}
	return &types.SwapAmount{
		AssetId:     uint32(toAssetId),
		AssetAmount: assetAmount.String(),
	}, nil
}
