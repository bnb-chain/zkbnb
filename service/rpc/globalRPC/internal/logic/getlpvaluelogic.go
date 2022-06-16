package logic

import (
	"context"
	"math/big"

	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
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
	if utils.CheckPairIndex(in.PairIndex) {
		logx.Errorf("[CheckPairIndex] param:%v", in.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckAmount(in.LPAmount) {
		logx.Errorf("[CheckAmount] param:%v", in.LPAmount)
		return nil, errcode.ErrInvalidParam
	}
	liquidity, err := l.commglobalmap.GetLatestLiquidityInfoForRead(int64(in.PairIndex))
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForRead] err:%v", err)
		return nil, err
	}
	amount, isTure := new(big.Int).SetString(in.LPAmount, 10)
	if !isTure {
		logx.Errorf("[SetString] err:%v", in.LPAmount)
		return nil, errcode.ErrInvalidParam
	}
	assetAAmount, assetBAmount := util.ComputeRemoveLiquidityAmount(liquidity, amount)
	return &globalRPCProto.RespGetLpValue{
		AssetAId:     uint32(liquidity.AssetAId),
		AssetAAmount: assetAAmount.String(),
		AssetBId:     uint32(liquidity.AssetBId),
		AssetBAmount: assetBAmount.String(),
	}, nil
}
