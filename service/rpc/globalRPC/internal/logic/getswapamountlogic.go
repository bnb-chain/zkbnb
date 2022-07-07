package logic

import (
	"context"
	"math/big"

	"github.com/zecrey-labs/zecrey-legend/common/checker"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSwapAmountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetSwapAmountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSwapAmountLogic {
	return &GetSwapAmountLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetSwapAmountLogic) GetSwapAmount(in *globalRPCProto.ReqGetSwapAmount) (*globalRPCProto.RespGetSwapAmount, error) {
	if checker.CheckPairIndex(in.PairIndex) {
		logx.Errorf("[CheckPairIndex] Parameter mismatch:%v", in.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	liquidity, err := l.commglobalmap.GetLatestLiquidityInfoForReadWithCache(l.ctx, int64(in.PairIndex))
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForReadWithCache] err:%v", err)
		return nil, err
	}
	if liquidity.AssetA == nil || liquidity.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidity.AssetB == nil || liquidity.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("liquidity:%v, err:%v", liquidity, errcode.ErrInvalidAsset)
		return &globalRPCProto.RespGetSwapAmount{}, errcode.ErrInvalidAsset
	}
	deltaAmount, isTure := new(big.Int).SetString(in.AssetAmount, 10)
	if !isTure {
		logx.Errorf("[SetString] err, AssetAmount:%v", in.AssetAmount)
		return nil, errcode.ErrInvalidParam
	}
	var assetAmount *big.Int
	var toAssetId int64
	if int64(in.AssetId) != liquidity.AssetAId && int64(in.AssetId) != liquidity.AssetBId {
		logx.Errorf("input:%v,liquidity:%v, err:%v", in, liquidity, errcode.ErrInvalidAsset)
		return &globalRPCProto.RespGetSwapAmount{}, errcode.ErrInvalidAssetID
	}
	assetAmount, toAssetId, err = util.ComputeDelta(liquidity.AssetA, liquidity.AssetB, liquidity.AssetAId, liquidity.AssetBId,
		int64(in.AssetId), in.IsFrom, deltaAmount, liquidity.FeeRate)
	if err != nil {
		logx.Errorf("[ComputeDelta] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetSwapAmount{
		SwapAssetAmount: assetAmount.String(),
		SwapAssetId:     uint32(toAssetId),
	}, nil
}
