package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/liquidity"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestAccountLpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	liquidity liquidity.Liquidity
	mempool   mempool.Mempool
}

func NewGetLatestAccountLpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAccountLpLogic {
	return &GetLatestAccountLpLogic{
		ctx:       ctx,
		svcCtx:    svcCtx,
		Logger:    logx.WithContext(ctx),
		liquidity: liquidity.New(svcCtx.Config),
		mempool:   mempool.New(svcCtx.Config),
	}
}

func (l *GetLatestAccountLpLogic) GetLatestAccountLp(in *globalRPCProto.ReqGetLatestAccountLp) (*globalRPCProto.RespGetLatestAccountLp, error) {
	if utils.CheckAccountIndex(in.AccountIndex) {
		logx.Error("[CheckAccountIndex] param:%v", in.AccountIndex)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckPairIndex(in.PairIndex) {
		logx.Error("[CheckPairIndex] param:%v", in.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	accountLpEnc, err := l.getLatestAccountLpInfo(uint32(in.AccountIndex), uint32(in.PairIndex))
	if err != nil {
		logx.Error("[GetLatestAccountLpInfo] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetLatestAccountLp{
		LpAmount: accountLpEnc,
	}, nil
}

func (l *GetLatestAccountLpLogic) getLatestAccountLpInfo(accountIndex uint32, pairIndex uint32) (
	accountLpEnc string,
	err error) {
	globalKey := globalmapHandler.GetAccountLPGlobalKey(accountIndex, pairIndex)
	ifExisted, globalValue := globalmapHandler.HandleGlobalMapGet(globalKey)
	if ifExisted {
		return globalValue, nil
	}
	mempoolDetails, err := l.mempool.GetAccountAssetMempoolDetails(
		//TODO: pairIndex == AssetId ?
		int64(accountIndex), int64(pairIndex),
		// TODOL: LiquidityLpAssetType not found in legend, but found GeneralAssetType
		commonAsset.GeneralAssetType, // commonAsset.LiquidityLpAssetType
	)
	// TODO:  err != mempool.ErrNotFound {
	if err != nil {
		logx.Error("[GetLiquidityByPairIndex] err:%v", err)
		return "", err
	}
	pairInfo, err := l.liquidity.GetLiquidityByPairIndex(int64(pairIndex))
	if err != nil {
		logx.Error("[GetLiquidityByPairIndex] err:%v", err)
		return "", err
	}
	finalBalance, err := globalmapHandler.UpdateSingleAssetGlobalMapByMempoolDetails(mempoolDetails,
		&globalmapHandler.GlobalAssetInfo{
			AccountIndex: int64(accountIndex),
			AssetId:      int64(pairIndex),
			// TODOL: LiquidityLpAssetType not found in legend, but found GeneralAssetType
			AssetType:      commonAsset.GeneralAssetType, // commonAsset.LiquidityLpAssetType
			BaseBalanceEnc: pairInfo.LpAmount,
		})
	if err != nil {
		logx.Error("[UpdateSingleAssetGlobalMapByMempoolDetails] err:%v", err)
		return "", err
	}
	return finalBalance, nil
}
