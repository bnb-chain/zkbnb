package commglobalmap

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/multcache"
	commGlobalmapHandler "github.com/bnb-chain/zkbas/common/util/globalmapHandler"
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
		return nil, err
	}
	mempoolTxs, err := m.mempoolModel.GetPendingLiquidityTxs()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, err
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
		return nil, err
	}
	for _, mempoolTx := range mempoolTxs {
		for _, txDetail := range mempoolTx.MempoolDetails {
			if txDetail.AssetType != commonAsset.LiquidityAssetType || liquidityInfo.PairIndex != txDetail.AssetId {
				continue
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				return nil, err
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				return nil, err
			}
		}
	}
	// TODO: this set cache operation will be deleted in the future, we should use GetLatestLiquidityInfoForReadWithCache anywhere
	// and delete the cache where mempool be changed
	infoBytes, err := json.Marshal(liquidityInfo)
	if err != nil {
		logx.Errorf("[json.Marshal] unable to marshal, err : %s", err.Error())
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
		return nil, err
	}
	mempoolTxs, err := m.mempoolModel.GetPendingLiquidityTxs()
	if err != nil && err != errorcode.DbErrNotFound {
		return nil, err
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
		logx.Errorf("[ConstructLiquidityInfo] err: %s", err.Error())
		return nil, err
	}
	for _, mempoolTx := range mempoolTxs {
		for _, txDetail := range mempoolTx.MempoolDetails {
			if txDetail.AssetType != commonAsset.LiquidityAssetType || liquidityInfo.PairIndex != txDetail.AssetId {
				continue
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				return nil, err
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				return nil, err
			}
		}
	}
	return liquidityInfo, nil
}

func (m *model) SetLatestLiquidityInfoForWrite(ctx context.Context, pairIndex int64) error {
	liquidityInfo, err := m.GetLatestLiquidityInfoForWrite(ctx, pairIndex)
	if err != nil {
		return err
	}
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyLiquidityInfoForWriteByPairIndex(pairIndex), liquidityInfo, 10); err != nil {
		return err
	}
	return nil
}

func (m *model) DeleteLatestLiquidityInfoForWriteInCache(ctx context.Context, pairIndex int64) error {
	return m.cache.Delete(ctx, multcache.SpliceCacheKeyLiquidityInfoForWriteByPairIndex(pairIndex))
}
