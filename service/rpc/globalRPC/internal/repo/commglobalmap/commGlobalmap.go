package commglobalmap

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	commGlobalmapHandler "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
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
	cache                multcache.MultCache
}

func (m *model) GetLatestAccountInfoWithCache(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	f := func() (interface{}, error) {
		accountInfo, err := m.GetLatestAccountInfo(ctx, accountIndex)
		if err != nil {
			return nil, err
		}
		account, err := commonAsset.FromFormatAccountInfo(accountInfo)
		if err != nil {
			return nil, err
		}
		return account, nil
	}
	accountInfo := &account.Account{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), accountInfo, 1, f)
	if err != nil {
		return nil, err
	}
	account, _ := value.(*account.Account)
	res, err := commonAsset.ToFormatAccountInfo(account)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *model) DeleteLatestAccountInfoInCache(ctx context.Context, accountIndex int64) error {
	return m.cache.Delete(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex))
}

func (m *model) GetLatestAccountInfo(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	oAccountInfo, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		logx.Errorf("[GetAccountByAccountIndex]param:%v, err:%v", accountIndex, err)
		return nil, err
	}
	accountInfo, err := commonAsset.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		logx.Errorf("[ToFormatAccountInfo]param:%v, err:%v", oAccountInfo, err)
		return nil, err
	}
	mempoolTxs, err := m.mempoolModel.GetPendingMempoolTxsByAccountIndex(accountIndex)
	if err != nil && err != mempool.ErrNotFound {
		logx.Errorf("[GetPendingMempoolTxsByAccountIndex]param:%v, err:%v", accountIndex, err)
		return nil, err
	}
	for _, mempoolTx := range mempoolTxs {
		if mempoolTx.Nonce != commonConstant.NilNonce {
			accountInfo.Nonce = mempoolTx.Nonce
		}
		for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
			if mempoolTxDetail.AccountIndex != accountIndex {
				continue
			}
			switch mempoolTxDetail.AssetType {
			case commonAsset.GeneralAssetType:
				if accountInfo.AssetInfo[mempoolTxDetail.AssetId] == nil {
					accountInfo.AssetInfo[mempoolTxDetail.AssetId] = &commonAsset.AccountAsset{
						AssetId:                  mempoolTxDetail.AssetId,
						Balance:                  util.ZeroBigInt,
						LpAmount:                 util.ZeroBigInt,
						OfferCanceledOrFinalized: util.ZeroBigInt,
					}
				}
				nBalance, err := commonAsset.ComputeNewBalance(commonAsset.GeneralAssetType,
					accountInfo.AssetInfo[mempoolTxDetail.AssetId].String(), mempoolTxDetail.BalanceDelta)
				if err != nil {
					logx.Errorf("[ComputeNewBalance] err:%v", err)
					return nil, err
				}
				accountInfo.AssetInfo[mempoolTxDetail.AssetId], err = commonAsset.ParseAccountAsset(nBalance)
				if err != nil {
					logx.Errorf("[ParseAccountAsset]param:%v, err:%v", nBalance, err)
					return nil, err
				}
			case commonAsset.CollectionNonceAssetType:
				accountInfo.CollectionNonce, err = strconv.ParseInt(mempoolTxDetail.BalanceDelta, 10, 64)
				if err != nil {
					logx.Errorf("[ParseInt] unable to parse int: err:%v", err)
					return nil, err
				}
			case commonAsset.LiquidityAssetType:
			case commonAsset.NftAssetType:
			default:
				logx.Errorf("invalid asset type")
				return nil, errcode.ErrInvalidAssetType
			}
		}
	}
	accountInfo.Nonce = accountInfo.Nonce + 1
	accountInfo.CollectionNonce = accountInfo.CollectionNonce + 1
	// TODO: this set cache operation will be deleted in the future, we should use GetLatestAccountInfoWithCache anywhere
	// and delete the cache where mempool be changed
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), accountInfo, 1); err != nil {
		return nil, err
	}
	return accountInfo, nil
}

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
	var dbLiquidityInfo *liquidity.Liquidity
	dbLiquidityInfo, err = m.liquidityModel.GetLiquidityByPairIndex(pairIndex)
	if err != nil {
		logx.Errorf("[GetLiquidityByPairIndex] unable to get latest liquidity by pair index: %v", err)
		return nil, err
	}
	mempoolTxs, err := m.mempoolModel.GetPendingLiquidityTxs()
	if err != nil {
		if err != mempool.ErrNotFound {
			logx.Errorf("[GetPendingLiquidityTxs] unable to get mempool txs by account index: %v", err)
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
		logx.Errorf("[ConstructLiquidityInfo] unable to construct pool info: %v", err)
		return nil, err
	}
	for _, mempoolTx := range mempoolTxs {
		for _, txDetail := range mempoolTx.MempoolDetails {
			if txDetail.AssetType != commonAsset.LiquidityAssetType || liquidityInfo.PairIndex != txDetail.AssetId {
				continue
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[ComputeNewBalance] unable to compute new balance: %v", err)
				return nil, err
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				logx.Errorf("[ParseLiquidityInfo] unable to parse pool info: %v", err)
				return nil, err
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
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyLiquidityByPairIndex(pairIndex), infoBytes, 1); err != nil {
		return nil, err
	}
	return liquidityInfo, nil
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

func (m *model) GetBasicAccountInfo(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error) {
	oAccountInfo, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		logx.Errorf("[GetBasicAccountInfo] unable to get account by account index: %s", err.Error())
		return nil, err
	}
	accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		logx.Errorf("[GetBasicAccountInfo] unable to get basic account info: %s", err.Error())
		return nil, err
	}
	// TODO: this set cache operation will be deleted in the future, we should use GetLatestLiquidityInfoForReadWithCache anywhere
	// and delete the cache where mempool be changed
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyBasicAccountByAccountIndex(accountIndex), accountInfo, 10); err != nil {
		return nil, err
	}
	return accountInfo, nil
}
