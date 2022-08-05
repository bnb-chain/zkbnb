package logic

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/checker"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type GetLpValueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetLpValueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLpValueLogic {
	return &GetLpValueLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetLpValueLogic) GetLpValue(in *globalRPCProto.ReqGetLpValue) (*globalRPCProto.RespGetLpValue, error) {
	if checker.CheckPairIndex(in.PairIndex) {
		logx.Errorf("[CheckPairIndex] param: %d", in.PairIndex)
		return nil, errorcode.RpcErrInvalidParam.RefineError("invalid PairIndex")
	}
	amount, isTure := new(big.Int).SetString(in.LPAmount, 10)
	if !isTure {
		logx.Errorf("[SetString] err: %s", in.LPAmount)
		return nil, errorcode.RpcErrInvalidParam
	}

	liquidity, err := l.commglobalmap.GetLatestLiquidityInfoForReadWithCache(l.ctx, int64(in.PairIndex))
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForReadWithCache] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrNotFound
		}
		return nil, errorcode.RpcErrInternal
	}
	assetAAmount, assetBAmount, err := util.ComputeRemoveLiquidityAmount(liquidity, amount)
	if err != nil {
		logx.Errorf("[ComputeRemoveLiquidityAmount] err: %s", err.Error())
		return nil, errorcode.RpcErrInternal
	}
	return &globalRPCProto.RespGetLpValue{
		AssetAId:     uint32(liquidity.AssetAId),
		AssetAAmount: assetAAmount.String(),
		AssetBId:     uint32(liquidity.AssetBId),
		AssetBAmount: assetBAmount.String(),
	}, nil
}
