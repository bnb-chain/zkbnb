package pair

import (
	"context"
	"math/big"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetLpValueLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLpValueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLpValueLogic {
	return &GetLpValueLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLpValueLogic) GetLPValue(req *types.ReqGetLpValue) (resp *types.LpValue, err error) {
	amount, isTure := new(big.Int).SetString(req.LpAmount, 10)
	if !isTure {
		logx.Errorf("fail to convert string: %s to int", req.LpAmount)
		return nil, types2.AppErrInvalidParam.RefineError("invalid LpAmount")
	}

	liquidity, err := l.svcCtx.StateFetcher.GetLatestLiquidity(int64(req.PairIndex))
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNotFound
		}
		return nil, types2.AppErrInternal
	}
	assetAAmount, assetBAmount := big.NewInt(0), big.NewInt(0)
	if liquidity.LpAmount.Cmp(types2.ZeroBigInt) > 0 {
		assetAAmount, assetBAmount, err = chain.ComputeRemoveLiquidityAmount(liquidity, amount)
		if err != nil {
			logx.Errorf("fail to compute liquidity amount, err: %s", err.Error())
			return nil, types2.AppErrInternal
		}
	}
	assetA, err := l.svcCtx.AssetModel.GetAssetById(liquidity.AssetAId)
	if err != nil {
		return nil, types2.AppErrInternal
	}
	assetB, err := l.svcCtx.AssetModel.GetAssetById(liquidity.AssetBId)
	if err != nil {
		return nil, types2.AppErrInternal
	}
	assetAPrice, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, assetA.AssetSymbol)
	if err != nil {
		return nil, types2.AppErrInternal
	}
	assetBPrice, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, assetB.AssetSymbol)
	if err != nil {
		return nil, types2.AppErrInternal
	}
	resp = &types.LpValue{
		AssetAId:     uint32(liquidity.AssetAId),
		AssetAName:   assetA.AssetName,
		AssetAAmount: assetAAmount.String(),
		AssetAPrice:  strconv.FormatFloat(assetAPrice, 'E', -1, 64),
		AssetBId:     uint32(liquidity.AssetBId),
		AssetBName:   assetB.AssetName,
		AssetBAmount: assetBAmount.String(),
		AssetBPrice:  strconv.FormatFloat(assetBPrice, 'E', -1, 64),
	}

	return resp, nil
}
