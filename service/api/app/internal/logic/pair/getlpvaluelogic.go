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
	if !utils.ValidatePairIndex(req.PairIndex) {
		logx.Errorf("invalid PairIndex: %d", req.PairIndex)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid PairIndex")
	}
	amount, isTure := new(big.Int).SetString(req.LpAmount, 10)
	if !isTure {
		logx.Errorf("fail to convert string: %s to int", req.LpAmount)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid LpAmount")
	}

	liquidity, err := l.svcCtx.StateFetcher.GetLatestLiquidity(int64(req.PairIndex))
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	assetAAmount, assetBAmount := big.NewInt(0), big.NewInt(0)
	if liquidity.LpAmount.Cmp(big.NewInt(0)) > 0 {
		assetAAmount, assetBAmount, err = util.ComputeRemoveLiquidityAmount(liquidity, amount)
		if err != nil {
			logx.Errorf("fail to compute liquidity amount, err: %s", err.Error())
			return nil, errorcode.AppErrInternal
		}
	}

	resp = &types.LpValue{
		AssetAId:     uint32(liquidity.AssetAId),
		AssetAAmount: assetAAmount.String(),
		AssetBId:     uint32(liquidity.AssetBId),
		AssetBAmount: assetBAmount.String(),
	}

	return resp, nil
}
