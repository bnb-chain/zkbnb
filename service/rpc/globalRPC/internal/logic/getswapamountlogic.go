package logic

import (
	"context"
	"math/big"

	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/pkg/zerror"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
	"github.com/bnb-chain/zkbas/utils"

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
	if utils.CheckPairIndex(in.PairIndex) {
		logx.Errorf("[GetSwapAmount CheckPairIndex] Parameter mismatch:%v", in.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	liquidity, err := l.commglobalmap.GetLatestLiquidityInfoForRead(int64(in.PairIndex))
	if err != nil {
		logx.Errorf("[GetSwapAmount GetLatestLiquidityInfoForRead] err:%v", err)
		return nil, err
	}
	if liquidity.AssetA == nil || liquidity.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidity.AssetB == nil || liquidity.AssetB.Cmp(big.NewInt(0)) == 0 {
		return &globalRPCProto.RespGetSwapAmount{}, zerror.New(-1, "[sendSwapTx func:GetSwapAmount] AssetA or AssetB is 0 in the liquidity Table")
	}
	deltaAmount, isTure := new(big.Int).SetString(in.AssetAmount, 10)
	if !isTure {
		logx.Errorf("[GetSwapAmount SetString] err, AssetAmount:%v", in.AssetAmount)
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
		logx.Errorf("[GetSwapAmount ComputeDelta] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetSwapAmount{
		SwapAssetAmount: assetAmount.String(),
		SwapAssetId:     uint32(toAssetId),
	}, nil
}
