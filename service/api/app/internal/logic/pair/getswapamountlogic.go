package pair

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetSwapAmountLogic struct {
	logx.Logger
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetSwapAmountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSwapAmountLogic {
	return &GetSwapAmountLogic{
		Logger:        logx.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetSwapAmountLogic) GetSwapAmount(req *types.ReqGetSwapAmount) (*types.RespGetSwapAmount, error) {
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
		logx.Errorf("[SetString] err, AssetAmount: %s", req.AssetAmount)
		return nil, errorcode.RpcErrInvalidParam
	}

	liquidity, err := l.commglobalmap.GetLatestLiquidityInfoForReadWithCache(l.ctx, int64(req.PairIndex))
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForReadWithCache] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrNotFound
		}
		return nil, errorcode.RpcErrInternal
	}
	if liquidity.AssetA == nil || liquidity.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidity.AssetB == nil || liquidity.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("liquidity: %v, err: %s", liquidity, errorcode.RpcErrLiquidityInvalidAssetAmount.Error())
		return nil, errorcode.RpcErrLiquidityInvalidAssetAmount
	}

	if int64(req.AssetId) != liquidity.AssetAId && int64(req.AssetId) != liquidity.AssetBId {
		logx.Errorf("input:%v,liquidity: %v, err: %s", req, liquidity, errorcode.RpcErrLiquidityInvalidAssetAmount.Error())
		return nil, errorcode.RpcErrLiquidityInvalidAssetID
	}
	logx.Errorf("[ComputeDelta] liquidity: %v", liquidity)
	logx.Errorf("[ComputeDelta] in: %v", req)
	logx.Errorf("[ComputeDelta] deltaAmount: %v", deltaAmount)

	var assetAmount *big.Int
	var toAssetId int64
	assetAmount, toAssetId, err = util.ComputeDelta(liquidity.AssetA, liquidity.AssetB, liquidity.AssetAId, liquidity.AssetBId,
		int64(req.AssetId), req.IsFrom, deltaAmount, liquidity.FeeRate)
	if err != nil {
		logx.Errorf("[ComputeDelta] err: %s", err.Error())
		return nil, errorcode.RpcErrInternal
	}
	logx.Errorf("[ComputeDelta] assetAmount:%v", assetAmount)
	return &types.RespGetSwapAmount{
		ResAssetAmount: assetAmount.String(),
		ResAssetId:     uint32(toAssetId),
	}, nil
}
