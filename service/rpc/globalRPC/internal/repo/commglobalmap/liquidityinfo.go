package commglobalmap

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	commGlobalmapHandler "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/errcode"
)

func (m *model) GetLatestLiquidityInfoForReadWithCache(ctx context.Context, pairIndex int64) (*commGlobalmapHandler.LiquidityInfo, error) {
	// f := func() (interface{}, error) {
	// 	tmpLiquidity, err := m.GetLatestLiquidityInfoForRead(ctx, pairIndex)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	infoBytes, err := json.Marshal(tmpLiquidity)
	// 	if err != nil {
	// 		logx.Errorf("[json.Marshal] unable to marshal: %v", err)
	// 		return nil, err
	// 	}
	// 	return &infoBytes, nil
	// }
	// var byteLiquidity []byte
	// value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyLiquidityByPairIndex(pairIndex), &byteLiquidity, 1, f)
	// if err != nil {
	// 	return nil, err
	// }
	// res, _ := value.(*[]byte)
	// liquidity := &commGlobalmapHandler.LiquidityInfo{}
	// err = json.Unmarshal([]byte(*res), &liquidity)
	// if err != nil {
	// 	logx.Errorf("[json.Unmarshal] unable to unmarshal liquidity info: %v", err)
	// 	return nil, err
	// }
	// return liquidity, nil
	return m.GetLatestLiquidityInfoForRead(ctx, pairIndex)

}
func (m *model) GetLatestLiquidityInfoForRead(ctx context.Context, pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error) {
	dbLiquidityInfo, err := m.liquidityModel.GetLiquidityByPairIndex(pairIndex)
	if err != nil {
		return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprint("GetLiquidityByPairIndex:", err.Error()))
	}
	mempoolTxs, err := m.mempoolModel.GetPendingLiquidityTxs()
	if err != nil {
		if err != mempool.ErrNotFound {
			return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprint("GetPendingLiquidityTxs:", err.Error()))
		}
	}
	liquidityInfo, err = commonAsset.ConstructLiquidityInfo(
		pairIndex,
		dbLiquidityInfo.AssetAId,
		dbLiquidityInfo.AssetA,
		dbLiquidityInfo.AssetBId,
		dbLiquidityInfo.AssetB,
		dbLiquidityInfo.LpAmount,
		dbLiquidityInfo.KLast,
		dbLiquidityInfo.FeeRate,
		dbLiquidityInfo.TreasuryAccountIndex,
		dbLiquidityInfo.TreasuryRate)
	if err != nil {
		return nil, errcode.ErrConstructLiquidityInfo.RefineError(fmt.Sprint("ConstructLiquidityInfo:", err.Error()))
	}
	for _, mempoolTx := range mempoolTxs {
		for _, txDetail := range mempoolTx.MempoolDetails {
			if txDetail.AssetType != commonAsset.LiquidityAssetType || liquidityInfo.PairIndex != txDetail.AssetId {
				continue
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				return nil, errcode.ErrComputeNewBalance.RefineError(err.Error())
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				return nil, errcode.ErrParseLiquidityInfo.RefineError(err.Error())
			}
		}
	}
	// TODO: this set cache operation will be deleted in the future, we should use GetLatestLiquidityInfoForReadWithCache anywhere
	// and delete the cache where mempool be changed
	infoBytes, err := json.Marshal(liquidityInfo)
	if err != nil {
		logx.Errorf("[json.Marshal] unable to marshal: %v", err)
		return nil, err
	}
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyLiquidityForReadByPairIndex(pairIndex), infoBytes, 1); err != nil {
		return nil, err
	}
	return liquidityInfo, nil
}

func (m *model) GetLatestLiquidityInfoForWrite(ctx context.Context, pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error) {
	dbLiquidityInfo, err := m.liquidityModel.GetLiquidityByPairIndex(pairIndex)
	if err != nil {
		return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprint("GetLiquidityByPairIndex:", err.Error()))
	}
	mempoolTxs, err := m.mempoolModel.GetPendingLiquidityTxs()
	if err != nil && err != mempool.ErrNotFound {
		return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprint("GetPendingLiquidityTxs:", err.Error()))
	}
	liquidityInfo, err = commonAsset.ConstructLiquidityInfo(
		pairIndex,
		dbLiquidityInfo.AssetAId,
		dbLiquidityInfo.AssetA,
		dbLiquidityInfo.AssetBId,
		dbLiquidityInfo.AssetB,
		dbLiquidityInfo.LpAmount,
		dbLiquidityInfo.KLast,
		dbLiquidityInfo.FeeRate,
		dbLiquidityInfo.TreasuryAccountIndex,
		dbLiquidityInfo.TreasuryRate,
	)
	if err != nil {
		logx.Errorf("[ConstructLiquidityInfo] err: %v", err)
		return nil, err
	}
	for _, mempoolTx := range mempoolTxs {
		for _, txDetail := range mempoolTx.MempoolDetails {
			if txDetail.AssetType != commonAsset.LiquidityAssetType || liquidityInfo.PairIndex != txDetail.AssetId {
				continue
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				return nil, errcode.ErrComputeNewBalance.RefineError(err.Error())
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				return nil, errcode.ErrParseLiquidityInfo.RefineError(err.Error())
			}
		}
	}
	// TODO: this set cache operation will be deleted in the future, we should use GetLatestLiquidityInfoForReadWithCache anywhere
	// and delete the cache where mempool be changed
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyLiquidityInfoForWriteByPairIndex(pairIndex), liquidityInfo, 10); err != nil {
		return nil, err
	}
	return liquidityInfo, nil
}
