package logic

import (
	"context"
	"math/big"

	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/utils"

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
		commglobalmap: commglobalmap.New(svcCtx.Config),
	}
}

func (l *GetSwapAmountLogic) GetSwapAmount(in *globalRPCProto.ReqGetSwapAmount) (*globalRPCProto.RespGetSwapAmount, error) {
	if utils.CheckPairIndex(in.PairIndex) {
		logx.Errorf("[CheckPairIndex] param:%v", in.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckAmount(in.AssetAmount) {
		logx.Errorf("[CheckAmount] param:%v", in.AssetAmount)
		return nil, errcode.ErrInvalidParam
	}
	liquidity, err := l.commglobalmap.GetLatestLiquidityInfoForRead(int64(in.PairIndex))
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForRead] err:%v", err)
		return nil, err
	}
	if liquidity.AssetA == nil || liquidity.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidity.AssetB == nil || liquidity.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("[sendSwapTx] invalid params")
		return nil, zerror.New(-1, "[sendSwapTx] invalid params")
	}
	deltaAmount, isTure := new(big.Int).SetString(in.AssetAmount, 10)
	if !isTure {
		logx.Errorf("[SetString] err, in.AssetAmount:%v", in.AssetAmount)
		return nil, errcode.ErrInvalidParam
	}
	var assetAmount *big.Int
	var toAssetId int64

	switch in.AssetId {
	case uint32(liquidity.AssetAId):
		assetAmount, toAssetId, err = util.ComputeDelta(liquidity.AssetA, liquidity.AssetB, liquidity.AssetAId, liquidity.AssetBId,
			int64(in.AssetId), in.IsFrom, deltaAmount, liquidity.FeeRate)
	case uint32(liquidity.AssetBId):
		assetAmount, toAssetId, err = util.ComputeDelta(liquidity.AssetA, liquidity.AssetB, liquidity.AssetAId, liquidity.AssetBId,
			int64(in.AssetId), in.IsFrom, deltaAmount, liquidity.FeeRate)
	default:
		return &globalRPCProto.RespGetSwapAmount{}, zerror.New(-1, "invalid pair assetIds")
	}
	if err != nil {
		logx.Errorf("[ComputeDelta] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetSwapAmount{
		SwapAssetAmount: assetAmount.String(),
		SwapAssetId:     uint32(toAssetId),
	}, nil
}
