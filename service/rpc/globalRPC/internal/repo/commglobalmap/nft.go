package commglobalmap

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/errcode"
)

type model struct {
	mempoolModel         mempool.MempoolModel
	mempoolTxDetailModel mempool.MempoolTxDetailModel
	accountModel         account.AccountModel
	liquidityModel       liquidity.LiquidityModel
	redisConnection      *redis.Redis
	offerModel           nft.OfferModel
	nftModel             nft.L2NftModel
	cache                multcache.MultCache
}

func (m *model) GetLatestOfferIdForWrite(ctx context.Context, accountIndex int64) (int64, error) {
	f := func() (interface{}, error) {
		lastOfferId, err := m.offerModel.GetLatestOfferId(accountIndex)
		if err != nil {
			return nil, err
		}
		return &lastOfferId, nil
	}
	var lastOfferId int64
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyOfferIdByAccountIndex(accountIndex), &lastOfferId, 1, f)
	if err != nil {
		if err.Error() == "OfferId not exist" {
			return 0, nil
		}
		return 0, err
	}
	nftIndex, _ := value.(*int64)
	return *nftIndex, nil
}

func (m *model) GetLatestNftInfoForRead(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error) {
	dbNftInfo, err := m.nftModel.GetNftAsset(nftIndex)
	if err != nil {
		return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprintf("GetNftAsset:%v", err))
	}
	mempoolTxs, err := m.mempoolModel.GetPendingNftTxs()
	if err != nil && err != mempool.ErrNotFound {
		return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprintf("GetPendingNftTxs:%v", err))
	}
	nftInfo := commonAsset.ConstructNftInfo(nftIndex, dbNftInfo.CreatorAccountIndex, dbNftInfo.OwnerAccountIndex, dbNftInfo.NftContentHash,
		dbNftInfo.NftL1TokenId, dbNftInfo.NftL1Address, dbNftInfo.CreatorTreasuryRate, dbNftInfo.CollectionId)
	for _, mempoolTx := range mempoolTxs {
		for _, txDetail := range mempoolTx.MempoolDetails {
			if txDetail.AssetType != commonAsset.NftAssetType || txDetail.AssetId != nftInfo.NftIndex {
				continue
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.NftAssetType, nftInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				return nil, errcode.ErrComputeNewBalance.RefineError(err)
			}
			nftInfo, err = commonAsset.ParseNftInfo(nBalance)
			if err != nil {
				return nil, errcode.ErrParseNftInfo.RefineError(err)
			}
		}
	}
	// TODO: this set cache operation will be deleted in the future, we should use GetLatestLiquidityInfoForReadWithCache anywhere
	// and delete the cache where mempool be changed
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyNftInfoByNftIndex(nftIndex), nftInfo, 10); err != nil {
		return nil, err
	}
	return nftInfo, nil
}

func (m *model) GetLatestNftInfoForReadWithCache(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error) {
	f := func() (interface{}, error) {
		tmpNftInfo, err := m.GetLatestNftInfoForRead(ctx, nftIndex)
		if err != nil {
			return nil, err
		}
		return tmpNftInfo, nil
	}
	nftInfoType := &commonAsset.NftInfo{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyNftInfoByNftIndex(nftIndex), nftInfoType, 10, f)
	if err != nil {
		return nil, err
	}
	nftInfo, _ := value.(*commonAsset.NftInfo)
	return nftInfo, nil
}

func (m *model) SetLatestNftInfoForReadInCache(ctx context.Context, nftIndex int64) error {
	nftInfo, err := m.GetLatestNftInfoForRead(ctx, nftIndex)
	if err != nil {
		return err
	}
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyNftInfoByNftIndex(nftIndex), nftInfo, 10); err != nil {
		return err
	}
	return nil
}

func (m *model) DeleteLatestNftInfoForReadInCache(ctx context.Context, nftIndex int64) error {
	return m.cache.Delete(ctx, multcache.SpliceCacheKeyNftInfoByNftIndex(nftIndex))
}
